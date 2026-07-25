package rcon

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// fakeExecutor implements RCONExecutor for tests.
type fakeExecutor struct {
	// inputs captured from the last call
	gotUserID   uint
	gotServerID string
	gotCommand  string
	// outputs to return
	out string
	err error
}

func (f *fakeExecutor) ExecuteRCONCommand(userID uint, serverID string, command string) (string, error) {
	f.gotUserID = userID
	f.gotServerID = serverID
	f.gotCommand = command
	return f.out, f.err
}

func newTestRouter(userID uint, executor RCONExecutor) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Simulate the AuthMiddleware setting user_id into context.
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	r.POST("/servers/:id/rcon/execute", ExecuteRCON)
	// Replace the package-level executor with the fake. The package-level
	// var is preserved so other callers (main) keep using the real service.
	prev := executorTesting()
	setExecutorForTest(executor)
	_ = prev
	return r
}

// executorTesting is exposed separately so that tests can restore the
// previously-installed executor between cases. It is not used outside tests.
func executorTesting() RCONExecutor { return executor }

func doExecuteRequest(r *gin.Engine, serverID, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/servers/"+serverID+"/rcon/execute", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestExecuteRCON_HappyPath verifies binding -> service call -> 200 with output.
func TestExecuteRCON_HappyPath(t *testing.T) {
	fake := &fakeExecutor{out: "Server: TheIsland, Players: 0/70\n", err: nil}
	r := newTestRouter(7, fake)
	defer setExecutorForTest(nil) // reset after test

	w := doExecuteRequest(r, "3", `{"command":"status"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, w.Body.String())
	}

	if fake.gotUserID != 7 {
		t.Errorf("expected userID 7, got %d", fake.gotUserID)
	}
	if fake.gotServerID != "3" {
		t.Errorf("expected serverID \"3\", got %q", fake.gotServerID)
	}
	if fake.gotCommand != "status" {
		t.Errorf("expected command \"status\", got %q", fake.gotCommand)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON body: %v, body=%s", err, w.Body.String())
	}
	if resp["message"] != "Command executed" {
		t.Errorf("expected message \"Command executed\", got %v", resp["message"])
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["output"] == nil || strings.TrimSpace(data["output"].(string)) == "" {
		t.Errorf("expected non-empty output, got %v", data["output"])
	}
}

// TestExecuteRCON_ServiceError verifies backend errors are surfaced as 500.
func TestExecuteRCON_ServiceError(t *testing.T) {
	fake := &fakeExecutor{err: errors.New("rcon dial localhost:32330: connection refused")}
	r := newTestRouter(7, fake)
	defer setExecutorForTest(nil)

	w := doExecuteRequest(r, "5", `{"command":"listplayers"}`)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d, body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid JSON body: %v", err)
	}
	if resp["message"] != "RCON execution failed" {
		t.Errorf("expected error message, got %v", resp["message"])
	}
}

// TestExecuteRCON_InvalidJSON verifies malformed bodies are 400.
func TestExecuteRCON_InvalidJSON(t *testing.T) {
	fake := &fakeExecutor{}
	r := newTestRouter(7, fake)
	defer setExecutorForTest(nil)

	w := doExecuteRequest(r, "1", `not-json`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	// Backend must NOT have been invoked.
	if fake.gotCommand != "" {
		t.Errorf("expected executor NOT to be called, but it captured %q", fake.gotCommand)
	}
}

// TestExecuteRCON_EmptyCommand verifies missing/binding-required rejects empty command.
func TestExecuteRCON_EmptyCommand(t *testing.T) {
	fake := &fakeExecutor{}
	r := newTestRouter(7, fake)
	defer setExecutorForTest(nil)

	w := doExecuteRequest(r, "1", `{"command":""}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty command, got %d, body=%s", w.Code, w.Body.String())
	}
}

// TestExecuteRCON_MissingCommand verifies absence of `command` key is rejected.
func TestExecuteRCON_MissingCommand(t *testing.T) {
	fake := &fakeExecutor{}
	r := newTestRouter(7, fake)
	defer setExecutorForTest(nil)

	w := doExecuteRequest(r, "1", `{}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing command, got %d, body=%s", w.Code, w.Body.String())
	}
}
