package docker_manager

import (
	"ark-server-commander/utils"
	"testing"
)

func init() {
	// Initialize logger
	utils.InitLogger()
}

// TestRollbackManager 
func TestRollbackManager(t *testing.T) {
	rm := NewRollbackManager()

	// 
	executed := false
	rm.AddAction("test", "resource1", " ", func() error {
		executed = true
		return nil
	})

	if rm.Count() != 1 {
		t.Errorf(" 1， %d", rm.Count())
	}

	// 
	if err := rm.Rollback(); err != nil {
		t.Errorf(" : %v", err)
	}

	if !executed {
		t.Error(" ")
	}
}

// TestRollbackManagerMultipleActions 
func TestRollbackManagerMultipleActions(t *testing.T) {
	rm := NewRollbackManager()

	var executionOrder []int

	// 
	rm.AddAction("test", "resource1", " 1", func() error {
		executionOrder = append(executionOrder, 1)
		return nil
	})

	rm.AddAction("test", "resource2", " 2", func() error {
		executionOrder = append(executionOrder, 2)
		return nil
	})

	rm.AddAction("test", "resource3", " 3", func() error {
		executionOrder = append(executionOrder, 3)
		return nil
	})

	// 
	if err := rm.Rollback(); err != nil {
		t.Errorf(" : %v", err)
	}

	// （Yes）
	if len(executionOrder) != 3 {
		t.Errorf(" 3 ， %d ", len(executionOrder))
	}

	if executionOrder[0] != 3 || executionOrder[1] != 2 || executionOrder[2] != 1 {
		t.Errorf(" Error， [3,2,1]， %v", executionOrder)
	}
}

// TestRollbackManagerClear 
func TestRollbackManagerClear(t *testing.T) {
	rm := NewRollbackManager()

	rm.AddAction("test", "resource1", " 1", func() error {
		return nil
	})

	if rm.Count() != 1 {
		t.Errorf(" 1， %d", rm.Count())
	}

	rm.Clear()

	if rm.Count() != 0 {
		t.Errorf(" 0， %d", rm.Count())
	}
}
