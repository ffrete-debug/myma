package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"ark-server-commander/config"
	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/storage"
	"ark-server-commander/utils"

	"go.uber.org/zap"
)

// tickInterval is how often the scheduler looks for due work. Schedules are
// expressed in hours, so a minute of granularity is far finer than needed and
// keeps the loop cheap.
const tickInterval = time.Minute

// uploadTimeoutPerBackup bounds a single archive upload so one slow transfer
// cannot stall the scheduler indefinitely.
const uploadTimeoutPerBackup = 30 * time.Minute

// Scheduler runs automated backups.
//
// A single goroutine processes due schedules serially. That is deliberate:
// backups are disk- and CPU-heavy (tar over a game volume), and running several
// at once on one host degrades the live servers the backups exist to protect.
type Scheduler struct {
	svc  *BackupService
	stop chan struct{}
	once sync.Once
}

func NewScheduler(svc *BackupService) *Scheduler {
	return &Scheduler{svc: svc, stop: make(chan struct{})}
}

// Start begins the scheduling loop. Safe to call once; later calls are no-ops.
func (s *Scheduler) Start() {
	go func() {
		ticker := time.NewTicker(tickInterval)
		defer ticker.Stop()

		utils.Info("backup scheduler started", zap.Duration("tick", tickInterval))

		for {
			select {
			case <-s.stop:
				utils.Info("backup scheduler stopped")
				return
			case <-ticker.C:
				s.runDue()
			}
		}
	}()
}

// Stop halts the loop. Idempotent, so it is safe on a shutdown path that may
// run more than once.
func (s *Scheduler) Stop() {
	s.once.Do(func() { close(s.stop) })
}

// runDue processes every schedule whose interval has elapsed.
func (s *Scheduler) runDue() {
	var schedules []models.BackupSchedule
	if err := database.DB.Where("enabled = ?", true).Find(&schedules).Error; err != nil {
		utils.Error("backup scheduler: load schedules failed", zap.Error(err))
		return
	}

	now := time.Now()
	for i := range schedules {
		sched := &schedules[i]
		if !sched.Due(now) {
			continue
		}
		// A failure on one server must not stop the others.
		if err := s.runOne(sched); err != nil {
			utils.Error("scheduled backup failed",
				zap.Uint("server_id", sched.ServerID), zap.Error(err))
		}
	}
}

// runOne performs a single scheduled backup and records the outcome.
//
// LastRunAt is written even when the backup fails. Otherwise a server whose
// backup keeps failing would be retried every tick, hammering the disk.
func (s *Scheduler) runOne(sched *models.BackupSchedule) error {
	started := time.Now()
	next := started.Add(time.Duration(sched.IntervalHours) * time.Hour)

	record := func(status, errMsg string) {
		if err := database.DB.Model(&models.BackupSchedule{}).
			Where("id = ?", sched.ID).
			Updates(map[string]any{
				"last_run_at": started,
				"next_run_at": next,
				"last_status": status,
				"last_error":  errMsg,
			}).Error; err != nil {
			utils.Error("backup scheduler: record outcome failed", zap.Error(err))
		}
	}

	backup, err := s.svc.CreateBackup(sched.UserID, fmt.Sprint(sched.ServerID))
	if err != nil {
		record("failed", err.Error())
		return fmt.Errorf("create backup: %w", err)
	}

	// CreateBackup archives asynchronously, so wait for the row to settle
	// before uploading or pruning; otherwise we would upload a partial file.
	final, err := s.awaitCompletion(backup.ID)
	if err != nil {
		record("failed", err.Error())
		return err
	}

	if sched.UploadToCloud {
		if err := UploadBackupToCloud(final.Filename); err != nil {
			// The local archive exists and is valid; only the off-host copy
			// failed. That is worth surfacing but is not a failed backup.
			record("uploaded_failed", err.Error())
			utils.Warn("scheduled backup: cloud upload failed",
				zap.Uint("server_id", sched.ServerID), zap.Error(err))
			s.prune(sched)
			return nil
		}
	}

	record("completed", "")
	s.prune(sched)
	return nil
}

// awaitCompletion polls the backup row until the async archive finishes.
func (s *Scheduler) awaitCompletion(backupID uint) (*models.Backup, error) {
	deadline := time.Now().Add(uploadTimeoutPerBackup)

	for time.Now().Before(deadline) {
		var b models.Backup
		if err := database.DB.Where("id = ?", backupID).First(&b).Error; err != nil {
			return nil, fmt.Errorf("reload backup: %w", err)
		}
		switch b.Status {
		case "completed":
			return &b, nil
		case "failed":
			return nil, fmt.Errorf("backup failed: %s", b.Error)
		}
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("backup did not complete within %s", uploadTimeoutPerBackup)
}

// prune enforces RetainCount, deleting the oldest automated archives beyond it.
//
// Both the file and the row are removed. A retained count of 0 disables pruning
// entirely, which is why it is not the default: ARK saves are large enough to
// fill a disk quietly.
func (s *Scheduler) prune(sched *models.BackupSchedule) {
	if sched.RetainCount <= 0 {
		return
	}

	var backups []models.Backup
	if err := database.DB.
		Where("server_id = ? AND status = ?", sched.ServerID, "completed").
		Order("created_at desc").
		Find(&backups).Error; err != nil {
		utils.Error("backup prune: list failed", zap.Error(err))
		return
	}

	if len(backups) <= sched.RetainCount {
		return
	}

	for _, old := range backups[sched.RetainCount:] {
		path := filepath.Join(BackupDir, old.Filename)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			utils.Warn("backup prune: remove file failed",
				zap.String("file", old.Filename), zap.Error(err))
			// Leave the row in place so the file is not orphaned invisibly.
			continue
		}
		if err := database.DB.Delete(&models.Backup{}, old.ID).Error; err != nil {
			utils.Warn("backup prune: delete row failed", zap.Error(err))
		}
	}
}

// UploadBackupToCloud mirrors a completed local archive to the configured
// cloud destination.
func UploadBackupToCloud(filename string) error {
	provider, err := CloudProvider()
	if err != nil {
		return err
	}
	if provider == nil || !provider.Configured() {
		return fmt.Errorf("cloud backup is not configured")
	}

	path := filepath.Join(BackupDir, filename)
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open backup: %w", err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat backup: %w", err)
	}

	if err := provider.Upload(filename, f, info.Size(), "application/gzip"); err != nil {
		return err
	}

	utils.Info("backup uploaded to cloud storage",
		zap.String("provider", provider.Name()),
		zap.String("file", filename),
		zap.Int64("bytes", info.Size()))
	return nil
}

// CloudProvider builds the configured destination, or nil when cloud upload is
// disabled. An unrecognised BACKUP_PROVIDER is an error rather than a silent
// fallback: an operator who mistyped it should be told, not left believing
// backups are being uploaded.
func CloudProvider() (storage.Provider, error) {
	kind, err := storage.ParseKind(config.BackupProvider)
	if err != nil {
		return nil, err
	}
	if kind == storage.KindNone {
		return nil, nil
	}
	return storage.New(cloudSettings(kind)), nil
}

// CloudConfigured reports whether a destination is usable, for the status
// endpoint and the schedule validation.
func CloudConfigured() bool {
	p, err := CloudProvider()
	return err == nil && p != nil && p.Configured()
}

// CloudDestination is a non-secret description of where uploads go.
func CloudDestination() string {
	p, err := CloudProvider()
	if err != nil || p == nil {
		return ""
	}
	return p.Destination()
}

// CloudProviderName is the selected provider's name, or "" when disabled.
func CloudProviderName() string {
	p, err := CloudProvider()
	if err != nil || p == nil {
		return ""
	}
	return p.Name()
}

func cloudSettings(kind storage.Kind) storage.Settings {
	return storage.Settings{
		Kind:                kind,
		S3:                  CloudConfig(),
		DropboxAccessToken:  config.DropboxAccessToken,
		DropboxRefreshToken: config.DropboxRefreshToken,
		DropboxAppKey:       config.DropboxAppKey,
		DropboxAppSecret:    config.DropboxAppSecret,
		DropboxPath:         config.DropboxPath,
		GDriveClientID:      config.GDriveClientID,
		GDriveClientSecret:  config.GDriveClientSecret,
		GDriveRefreshToken:  config.GDriveRefreshToken,
		GDriveFolderID:      config.GDriveFolderID,
		WebDAVURL:           config.WebDAVURL,
		WebDAVUsername:      config.WebDAVUsername,
		WebDAVPassword:      config.WebDAVPassword,
	}
}

// CloudConfig builds the storage config from the process configuration.
func CloudConfig() storage.Config {
	return storage.Config{
		Endpoint:  config.S3Endpoint,
		Region:    config.S3Region,
		Bucket:    config.S3Bucket,
		AccessKey: config.S3AccessKey,
		SecretKey: config.S3SecretKey,
		Prefix:    config.S3Prefix,
		PathStyle: config.S3PathStyle,
	}
}
