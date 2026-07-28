package middleware

import (
	"math"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

const tokenEpsilon = 1e-9

func TestRateLimiterRefill(t *testing.T) {
	tests := []struct {
		name    string
		rate    int
		burst   int
		window  time.Duration
		tokens  float64
		elapsed time.Duration
		want    float64
	}{
		// The bug this pins down: a tenth of a window used to credit 0 tokens.
		{"fraction of a window credits fractional tokens", 100, 200, time.Second, 0, 100 * time.Millisecond, 10},
		{"single millisecond still credits", 100, 200, time.Second, 0, time.Millisecond, 0.1},
		{"exactly one window credits the full rate", 100, 200, time.Second, 0, time.Second, 100},
		{"one and a half windows credits one and a half rates", 100, 200, time.Second, 0, 1500 * time.Millisecond, 150},
		{"refill is capped at burst", 100, 200, time.Second, 0, time.Hour, 200},
		{"no elapsed time credits nothing", 100, 200, time.Second, 5, 0, 5},
		{"clock going backwards credits nothing", 100, 200, time.Second, 5, -time.Second, 5},
		// The auth limiter wired up in main.go: 1 token per 5s, burst 10.
		{"auth limiter over one window", 1, 10, 5 * time.Second, 0, 5 * time.Second, 1},
		{"auth limiter over half a window", 1, 10, 5 * time.Second, 0, 2500 * time.Millisecond, 0.5},
		{"auth limiter caps at burst", 1, 10, 5 * time.Second, 0, time.Hour, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.rate, tt.burst, tt.window)
			base := time.Now()
			v := &visitor{tokens: tt.tokens, last: base}
			rl.refill(v, base.Add(tt.elapsed))

			if math.Abs(v.tokens-tt.want) > tokenEpsilon {
				t.Errorf("tokens = %v, want %v", v.tokens, tt.want)
			}
		})
	}
}

// A busy client refilling in many small steps must end up with the same budget
// as an idle one refilling in a single step. The old integer division truncated
// every sub-window step to zero, so sustained traffic never refilled at all.
func TestRateLimiterRefillAccumulatesAcrossSmallSteps(t *testing.T) {
	tests := []struct {
		name  string
		steps int
		step  time.Duration
		want  float64
	}{
		{"one window in ten steps", 10, 100 * time.Millisecond, 100},
		{"one window in a thousand steps", 1000, time.Millisecond, 100},
		{"half a window in fifty steps", 50, 10 * time.Millisecond, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(100, 200, time.Second)
			base := time.Now()
			v := &visitor{tokens: 0, last: base}
			for i := 1; i <= tt.steps; i++ {
				rl.refill(v, base.Add(time.Duration(i)*tt.step))
			}

			if math.Abs(v.tokens-tt.want) > tokenEpsilon {
				t.Errorf("tokens = %v, want %v", v.tokens, tt.want)
			}
		})
	}
}

func TestRateLimiterEviction(t *testing.T) {
	tests := []struct {
		name        string
		sinceSweep  time.Duration
		idle        time.Duration
		wantEvicted bool
	}{
		{"idle visitor is evicted", 2 * sweepInterval, 2 * time.Hour, true},
		{"recently seen visitor is kept", 2 * sweepInterval, 0, false},
		{"sweep not due yet keeps everything", 0, 2 * time.Hour, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(1, 10, 5*time.Second)
			now := time.Now()
			rl.visitors["192.0.2.1"] = &visitor{tokens: 10, last: now.Add(-tt.idle)}
			rl.lastSweep = now.Add(-tt.sinceSweep)

			rl.mu.Lock()
			rl.sweepLocked(now)
			rl.mu.Unlock()

			_, exists := rl.visitors["192.0.2.1"]
			if exists == tt.wantEvicted {
				t.Errorf("visitor present = %v, want %v", exists, !tt.wantEvicted)
			}
		})
	}
}

// A visitor must never be evicted before its bucket would have refilled to
// burst, otherwise eviction itself hands the client a free reset.
func TestRateLimiterTTLCoversFullRefill(t *testing.T) {
	tests := []struct {
		name   string
		rate   int
		burst  int
		window time.Duration
		want   time.Duration
	}{
		{"global limiter floors at the minimum", 100, 200, time.Second, minVisitorTTL},
		{"auth limiter floors at the minimum", 1, 10, 5 * time.Second, minVisitorTTL},
		{"slow limiter uses its full refill time", 1, 60, time.Minute, time.Hour},
		{"degenerate rate falls back to the minimum", 0, 10, time.Second, minVisitorTTL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter(tt.rate, tt.burst, tt.window)
			if rl.ttl != tt.want {
				t.Errorf("ttl = %v, want %v", rl.ttl, tt.want)
			}
		})
	}
}

func TestRateLimiterMiddlewareAllowsBurstThenBlocks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rl := NewRateLimiter(1, 3, time.Hour)
	router := gin.New()
	router.Use(rl.Middleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	for i := 1; i <= 4; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		want := 200
		if i == 4 {
			want = 429
		}
		if w.Code != want {
			t.Errorf("request %d: expected %d, got %d", i, want, w.Code)
		}
	}
}
