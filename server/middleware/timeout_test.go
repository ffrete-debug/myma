package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestTimeoutPassesThroughFastHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Timeout(2 * time.Second))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(201, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	if w.Code != 201 {
		t.Errorf("expected 201, got %d", w.Code)
	}
	if w.Body.String() != `{"ok":true}` {
		t.Errorf("unexpected body %q", w.Body.String())
	}
}

// The client must get a complete 504 at the deadline, and the slow handler's own
// response - written after the deadline - must be discarded rather than
// appended to it.
func TestTimeoutReleasesClientAndDiscardsLateWrite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	release := make(chan struct{})
	router := gin.New()
	router.Use(Timeout(50 * time.Millisecond))
	router.GET("/slow", func(c *gin.Context) {
		<-release
		c.JSON(200, gin.H{"ok": true})
	})
	srv := httptest.NewServer(router)
	defer srv.Close()
	defer close(release)

	start := time.Now()
	resp, err := http.Get(srv.URL + "/slow")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	elapsed := time.Since(start)

	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("expected 504, got %d", resp.StatusCode)
	}
	if string(body) != timeoutBody {
		t.Errorf("expected %q, got %q", timeoutBody, string(body))
	}
	// Generous bound: the point is that the client is not held for the whole
	// handler, not that the deadline is precise.
	if elapsed > time.Second {
		t.Errorf("client held for %v, expected release near the 50ms deadline", elapsed)
	}
}

// Once the handler has started writing, the timeout path must stay silent.
func TestTimeoutDoesNotOverwriteStartedResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	release := make(chan struct{})
	router := gin.New()
	router.Use(Timeout(50 * time.Millisecond))
	router.GET("/partial", func(c *gin.Context) {
		c.String(200, "hello")
		c.Writer.Flush()
		<-release
	})
	srv := httptest.NewServer(router)
	defer srv.Close()

	go func() {
		time.Sleep(200 * time.Millisecond)
		close(release)
	}()

	resp, err := http.Get(srv.URL + "/partial")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if string(body) != "hello" {
		t.Errorf("expected \"hello\", got %q", string(body))
	}
}
