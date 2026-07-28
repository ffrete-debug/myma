package models

import "time"

// BackupSchedule drives automated backups for one server.
//
// One row per server: a server either has automated backups or it does not, and
// allowing several overlapping schedules for the same volume would mostly
// produce concurrent archives of the same data.
type BackupSchedule struct {
	ID       uint `json:"id" gorm:"primarykey"`
	ServerID uint `json:"server_id" gorm:"not null;uniqueIndex"`
	UserID   uint `json:"user_id" gorm:"not null;index"`

	Enabled bool `json:"enabled" gorm:"not null;default:false"`
	// IntervalHours is the gap between runs. Hours rather than a cron
	// expression: this is the only granularity the UI offers, and a cron parser
	// would be a dependency and a support burden for no gained capability.
	IntervalHours int `json:"interval_hours" gorm:"not null;default:24"`
	// RetainCount caps how many automated archives are kept per server. 0 means
	// keep everything, which is a deliberate opt-in because ARK saves are large.
	RetainCount int `json:"retain_count" gorm:"not null;default:7"`
	// UploadToCloud mirrors each completed archive to object storage. Requires
	// the S3 settings to be configured; the scheduler logs and continues if not.
	UploadToCloud bool `json:"upload_to_cloud" gorm:"not null;default:false"`

	LastRunAt  *time.Time `json:"last_run_at"`
	LastStatus string     `json:"last_status"`
	LastError  string     `json:"last_error,omitempty"`
	NextRunAt  *time.Time `json:"next_run_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// BackupScheduleRequest is the upsert payload from the UI.
type BackupScheduleRequest struct {
	Enabled       bool `json:"enabled"`
	IntervalHours int  `json:"interval_hours" binding:"required,min=1,max=168"`
	RetainCount   int  `json:"retain_count" binding:"min=0,max=365"`
	UploadToCloud bool `json:"upload_to_cloud"`
}

// Due reports whether the schedule should run now.
func (s *BackupSchedule) Due(now time.Time) bool {
	if !s.Enabled || s.IntervalHours <= 0 {
		return false
	}
	if s.LastRunAt == nil {
		// Never run: start the cycle immediately rather than waiting a full
		// interval, so enabling a schedule visibly does something.
		return true
	}
	return now.Sub(*s.LastRunAt) >= time.Duration(s.IntervalHours)*time.Hour
}
