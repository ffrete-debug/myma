package docker_manager

import (
	"ark-server-commander/utils"
	"fmt"
	"go.uber.org/zap"
)

// RollbackAction 
type RollbackAction struct {
	Type        string       // : "volume", "container", "config"
	ResourceID  string       // ID
	Action      func() error // 
	Description string       // 
}

// RollbackManager 
type RollbackManager struct {
	actions []RollbackAction
}

// NewRollbackManager Create
func NewRollbackManager() *RollbackManager {
	return &RollbackManager{
		actions: make([]RollbackAction, 0),
	}
}

// AddAction 
func (rm *RollbackManager) AddAction(actionType, resourceID, description string, action func() error) {
	rm.actions = append(rm.actions, RollbackAction{
		Type:        actionType,
		ResourceID:  resourceID,
		Action:      action,
		Description: description,
	})
	utils.Debug(" ",
		zap.String("type", actionType),
		zap.String("resource", resourceID),
		zap.String("description", description))
}

// Rollback （）
func (rm *RollbackManager) Rollback() error {
	if len(rm.actions) == 0 {
		utils.Debug(" ")
		return nil
	}

	utils.Info("On ", zap.Int("count", len(rm.actions)))

	var rollbackErrors []error

	// 
	for i := len(rm.actions) - 1; i >= 0; i-- {
		action := rm.actions[i]
		utils.Info(" ",
			zap.String("type", action.Type),
			zap.String("resource", action.ResourceID),
			zap.String("description", action.Description))

		if err := action.Action(); err != nil {
			utils.Error(" Operation failed",
				zap.String("type", action.Type),
				zap.String("resource", action.ResourceID),
				zap.Error(err))
			rollbackErrors = append(rollbackErrors, fmt.Errorf("%s  : %w", action.Description, err))
		} else {
			utils.Info(" Operation successful",
				zap.String("type", action.Type),
				zap.String("resource", action.ResourceID))
		}
	}

	if len(rollbackErrors) > 0 {
		return fmt.Errorf(" Operation failed: %v", rollbackErrors)
	}

	utils.Info(" ")
	return nil
}

// Clear 
func (rm *RollbackManager) Clear() {
	rm.actions = make([]RollbackAction, 0)
	utils.Debug(" ")
}

// Count 
func (rm *RollbackManager) Count() int {
	return len(rm.actions)
}
