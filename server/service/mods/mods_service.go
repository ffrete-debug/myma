// Package mods manages the Steam Workshop mods attached to a server.
package mods

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"ark-server-commander/database"
	"ark-server-commander/models"
	"ark-server-commander/service/steam"

	"gorm.io/gorm"
)

// workshopIDPattern guards the one value that reaches a launch argument.
// Steam publishedfileids are decimal only; rejecting anything else keeps
// separators and shell metacharacters out of the generated -mods= string.
var workshopIDPattern = regexp.MustCompile(`^[0-9]{1,20}$`)

// maxModsPerServer bounds the launch line. ARK itself becomes unstable long
// before this, and an unbounded list would let one request build an enormous
// command line.
const maxModsPerServer = 100

type Service struct {
	steam *steam.Client
}

func NewService() *Service {
	return &Service{steam: steam.NewClient()}
}

// ownedServer loads a server the user owns, or reports not-found. Callers must
// never distinguish "does not exist" from "belongs to someone else".
func ownedServer(serverID, userID uint) (*models.Server, error) {
	var server models.Server
	if err := database.DB.Where("id = ? AND user_id = ?", serverID, userID).First(&server).Error; err != nil {
		return nil, fmt.Errorf("server not found")
	}
	return &server, nil
}

// ListMods returns a server's mods in load order.
func (s *Service) ListMods(serverID, userID uint) ([]models.ServerMod, error) {
	if _, err := ownedServer(serverID, userID); err != nil {
		return nil, err
	}

	var mods []models.ServerMod
	if err := database.DB.Where("server_id = ?", serverID).Order("position asc").Find(&mods).Error; err != nil {
		return nil, fmt.Errorf("list mods: %w", err)
	}
	return mods, nil
}

// AddMod attaches a Workshop item to a server, appending it to the load order.
//
// The item is looked up on Steam first so the stored name and preview are real
// rather than caller-supplied, and so a bogus ID is rejected before it can end
// up on a launch line.
func (s *Service) AddMod(ctx context.Context, serverID, userID uint, workshopID string) (*models.ServerMod, error) {
	if _, err := ownedServer(serverID, userID); err != nil {
		return nil, err
	}

	workshopID = strings.TrimSpace(workshopID)
	if !workshopIDPattern.MatchString(workshopID) {
		return nil, fmt.Errorf("invalid workshop id")
	}

	var count int64
	if err := database.DB.Model(&models.ServerMod{}).Where("server_id = ?", serverID).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("count mods: %w", err)
	}
	if count >= maxModsPerServer {
		return nil, fmt.Errorf("mod limit reached (%d)", maxModsPerServer)
	}

	items, err := s.steam.GetItems(ctx, []string{workshopID})
	if err != nil {
		return nil, fmt.Errorf("look up workshop item: %w", err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("workshop item not found for ARK")
	}
	item := items[0]

	mod := models.ServerMod{
		ServerID:   serverID,
		WorkshopID: workshopID,
		Name:       item.Title,
		PreviewURL: item.PreviewURL,
		Position:   int(count),
		Enabled:    true,
	}

	// The unique index on (server_id, workshop_id) is what actually prevents
	// duplicates; checking first would still race two concurrent requests.
	if err := database.DB.Create(&mod).Error; err != nil {
		if isDuplicate(err) {
			return nil, fmt.Errorf("mod already added to this server")
		}
		return nil, fmt.Errorf("add mod: %w", err)
	}

	if err := syncGameModIds(serverID); err != nil {
		return nil, err
	}

	return &mod, nil
}

// RemoveMod detaches a mod and closes the gap it leaves in the load order.
func (s *Service) RemoveMod(serverID, userID uint, workshopID string) error {
	if _, err := ownedServer(serverID, userID); err != nil {
		return err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("server_id = ? AND workshop_id = ?", serverID, workshopID).Delete(&models.ServerMod{})
		if res.Error != nil {
			return fmt.Errorf("remove mod: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("mod not found on this server")
		}
		if err := renumber(tx, serverID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return syncGameModIds(serverID)
}

// SetEnabled toggles a mod without losing its place in the load order.
func (s *Service) SetEnabled(serverID, userID uint, workshopID string, enabled bool) error {
	if _, err := ownedServer(serverID, userID); err != nil {
		return err
	}

	res := database.DB.Model(&models.ServerMod{}).
		Where("server_id = ? AND workshop_id = ?", serverID, workshopID).
		Update("enabled", enabled)
	if res.Error != nil {
		return fmt.Errorf("update mod: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("mod not found on this server")
	}

	return syncGameModIds(serverID)
}

// Reorder applies a complete desired load order.
//
// Load order is meaningful in ARK — later mods override earlier ones — so this
// takes the full list and rejects a partial one, which would otherwise leave
// unlisted mods at arbitrary positions.
func (s *Service) Reorder(serverID, userID uint, workshopIDs []string) error {
	if _, err := ownedServer(serverID, userID); err != nil {
		return err
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		var existing []models.ServerMod
		if err := tx.Where("server_id = ?", serverID).Find(&existing).Error; err != nil {
			return fmt.Errorf("load mods: %w", err)
		}

		if len(existing) != len(workshopIDs) {
			return fmt.Errorf("reorder must list all %d mods, got %d", len(existing), len(workshopIDs))
		}

		known := make(map[string]bool, len(existing))
		for _, m := range existing {
			known[m.WorkshopID] = true
		}
		seen := make(map[string]bool, len(workshopIDs))
		for _, id := range workshopIDs {
			if !known[id] {
				return fmt.Errorf("mod %s is not attached to this server", id)
			}
			if seen[id] {
				return fmt.Errorf("mod %s listed twice", id)
			}
			seen[id] = true
		}

		for i, id := range workshopIDs {
			if err := tx.Model(&models.ServerMod{}).
				Where("server_id = ? AND workshop_id = ?", serverID, id).
				Update("position", i).Error; err != nil {
				return fmt.Errorf("reorder mods: %w", err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return syncGameModIds(serverID)
}

// Search proxies Workshop search.
func (s *Service) Search(ctx context.Context, text string, page, limit int) ([]models.WorkshopItem, error) {
	return s.steam.Search(ctx, text, page, limit)
}

// GetItems proxies Workshop lookup by ID.
func (s *Service) GetItems(ctx context.Context, ids []string) ([]models.WorkshopItem, error) {
	for _, id := range ids {
		if !workshopIDPattern.MatchString(id) {
			return nil, fmt.Errorf("invalid workshop id")
		}
	}
	return s.steam.GetItems(ctx, ids)
}

// ActiveModIDs returns the enabled mods for a server in load order. This is
// what the launch-argument builder consumes.
func ActiveModIDs(serverID uint) ([]string, error) {
	var mods []models.ServerMod
	if err := database.DB.
		Where("server_id = ? AND enabled = ?", serverID, true).
		Order("position asc").
		Find(&mods).Error; err != nil {
		return nil, fmt.Errorf("load active mods: %w", err)
	}

	ids := make([]string, 0, len(mods))
	for _, m := range mods {
		// Defence in depth: the column is validated on write, but this value
		// goes onto a launch line, so it is re-checked on the way out.
		if workshopIDPattern.MatchString(m.WorkshopID) {
			ids = append(ids, m.WorkshopID)
		}
	}
	return ids, nil
}

// renumber makes positions dense and 0-based again after a deletion.
func renumber(tx *gorm.DB, serverID uint) error {
	var mods []models.ServerMod
	if err := tx.Where("server_id = ?", serverID).Order("position asc").Find(&mods).Error; err != nil {
		return fmt.Errorf("reload mods: %w", err)
	}
	for i, m := range mods {
		if m.Position == i {
			continue
		}
		if err := tx.Model(&models.ServerMod{}).Where("id = ?", m.ID).Update("position", i).Error; err != nil {
			return fmt.Errorf("renumber mods: %w", err)
		}
	}
	return nil
}

func isDuplicate(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "unique")
}

// syncGameModIds writes the enabled mods, in load order, into the server's
// GameModIds field.
//
// This is the step that actually makes mods take effect. The game image installs
// Workshop mods itself from the GameModIds environment variable
// (start_server.sh: steamcmd +workshop_download_item 346110 <id>), so the mod
// list has to reach that field. Nothing else consumes the server_mods table.
//
// Note it does NOT go on the command line as -mods=: that is an ASA/CurseForge
// flag which ASE ignores (docs/wiki.md:358).
//
// Installation is gated on a first-startup flag inside the container, so a
// changed mod set only takes effect when the container is recreated. Callers
// surface that to the user.
func syncGameModIds(serverID uint) error {
	ids, err := ActiveModIDs(serverID)
	if err != nil {
		return err
	}

	joined := strings.Join(ids, ",")
	if err := database.DB.Model(&models.Server{}).
		Where("id = ?", serverID).
		Update("game_mod_ids", joined).Error; err != nil {
		return fmt.Errorf("update server mod ids: %w", err)
	}
	return nil
}
