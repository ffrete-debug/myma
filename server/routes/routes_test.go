package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ark-server-commander/service/update"
	"ark-server-commander/websocket"

	"github.com/gin-gonic/gin"
)

// publicRoutes is the complete, explicit list of routes that may be reached
// without a token. Everything else RegisterRoutes registers must sit behind
// middleware.AuthMiddleware. Adding an entry here is a security decision — do
// not widen it just to make TestAllNonPublicRoutesRequireAuth pass
var publicRoutes = map[string]bool{
	"GET /health": true,

	// Authentication — these mint or revoke tokens, so they cannot demand one
	"GET /api/auth/check-init": true,
	"POST /api/auth/init":      true,
	"POST /api/auth/login":     true,
	"POST /api/auth/refresh":   true,
	"POST /api/auth/logout":    true,

	// Static assets, SPA fallback and swagger
	"GET /_next/*filepath":   true,
	"HEAD /_next/*filepath":  true,
	"GET /public/*filepath":  true,
	"HEAD /public/*filepath": true,
	"GET /favicon.ico":       true,
	"HEAD /favicon.ico":      true,
	"GET /swagger/*any":      true,
}

// alwaysRegisteredPublicRoutes are the public routes that exist regardless of
// whether a ./static build is present, so the allowlist above cannot rot
var alwaysRegisteredPublicRoutes = []string{
	"GET /health",
	"GET /api/auth/check-init",
	"POST /api/auth/init",
	"POST /api/auth/login",
	"POST /api/auth/refresh",
	"POST /api/auth/logout",
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	hub := websocket.NewHub()
	// The handlers are never reached in these tests — auth aborts first — so a
	// nil DB is enough to build the update service
	RegisterRoutes(r, update.NewUpdateService(nil, hub), hub)
	return r
}

// concretePath turns a gin route pattern into a requestable URL by filling in
// ":id" style params and "*filepath" catch-alls
func concretePath(pattern string) string {
	segs := strings.Split(pattern, "/")
	for i, seg := range segs {
		switch {
		case strings.HasPrefix(seg, ":"):
			segs[i] = "1"
		case strings.HasPrefix(seg, "*"):
			segs[i] = "x"
		}
	}
	return strings.Join(segs, "/")
}

func TestAllNonPublicRoutesRequireAuth(t *testing.T) {
	r := newTestRouter()

	registered := r.Routes()
	if len(registered) == 0 {
		t.Fatal("RegisterRoutes registered no routes")
	}

	for _, route := range registered {
		key := route.Method + " " + route.Path
		if publicRoutes[key] {
			continue
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest(route.Method, concretePath(route.Path), nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s is reachable without authentication (got %d, want 401): register it on the protected group, or add it to publicRoutes deliberately", key, w.Code)
		}
	}
}

func TestPublicRoutesAreRegistered(t *testing.T) {
	r := newTestRouter()

	registered := make(map[string]bool)
	for _, route := range r.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	for _, key := range alwaysRegisteredPublicRoutes {
		if !registered[key] {
			t.Errorf("%s is in the public allowlist but is not registered", key)
		}
	}
}

// TestUpdateStatusRouteRequiresAuth is the regression test for the update
// status endpoint, which was registered on the unauthenticated group while
// visually nested inside the protected block
func TestUpdateStatusRouteRequiresAuth(t *testing.T) {
	r := newTestRouter()

	var found bool
	for _, route := range r.Routes() {
		if route.Method == http.MethodGet && route.Path == "/api/updates/:id/status" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("GET /api/updates/:id/status is not registered")
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/updates/1/status", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated update status, got %d", w.Code)
	}
}

// TestWebSocketRoutesRequireAuth guards the WebSocket upgrades, which accept a
// token via query param and are easy to leave unprotected
func TestWebSocketRoutesRequireAuth(t *testing.T) {
	r := newTestRouter()

	paths := []string{"/api/ws/updates/1", "/api/ws/rcon/1", "/api/ws/logs/1"}
	for _, path := range paths {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("GET %s: expected 401 without a token, got %d", path, w.Code)
		}
	}
}
