package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// sweepInterval bounds how often the visitors map is scanned for eviction.
	sweepInterval = time.Minute
	// minVisitorTTL floors how long an idle visitor is kept, so a limiter with a
	// very short window does not churn its map on every sweep.
	minVisitorTTL = time.Minute
)

type RateLimiter struct {
	mu        sync.Mutex
	visitors  map[string]*visitor
	rate      int
	burst     int
	window    time.Duration
	ttl       time.Duration
	lastSweep time.Time
}

type visitor struct {
	// tokens is fractional so that refills shorter than one window are carried
	// instead of being truncated away.
	tokens float64
	last   time.Time
}

func NewRateLimiter(rate, burst int, window time.Duration) *RateLimiter {
	// A visitor may only be evicted once it has been idle long enough for its
	// bucket to have refilled all the way to burst - dropping it any sooner
	// would hand the client a free reset and defeat the limiter.
	ttl := minVisitorTTL
	if rate > 0 && window > 0 {
		if full := time.Duration(float64(burst) / float64(rate) * float64(window)); full > ttl {
			ttl = full
		}
	}
	return &RateLimiter{
		visitors:  make(map[string]*visitor),
		rate:      rate,
		burst:     burst,
		window:    window,
		ttl:       ttl,
		lastSweep: time.Now(),
	}
}

// refill credits the tokens accrued since v.last, carrying fractional windows.
// The previous integer division truncated any elapsed time shorter than window
// down to zero tokens while still advancing v.last, so a client sending faster
// than one request per window never refilled at all.
func (rl *RateLimiter) refill(v *visitor, now time.Time) {
	elapsed := now.Sub(v.last)
	if elapsed <= 0 {
		return
	}
	if rl.window > 0 {
		v.tokens += float64(elapsed) / float64(rl.window) * float64(rl.rate)
	}
	if v.tokens > float64(rl.burst) {
		v.tokens = float64(rl.burst)
	}
	v.last = now
}

// sweepLocked drops visitors idle for longer than ttl. It must be called with
// rl.mu held and is a no-op until sweepInterval has elapsed, which amortises the
// O(n) scan across requests instead of needing a background goroutine.
func (rl *RateLimiter) sweepLocked(now time.Time) {
	if now.Sub(rl.lastSweep) < sweepInterval {
		return
	}
	rl.lastSweep = now
	for ip, v := range rl.visitors {
		if now.Sub(v.last) > rl.ttl {
			delete(rl.visitors, ip)
		}
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()
		rl.mu.Lock()
		rl.sweepLocked(now)
		v, exists := rl.visitors[ip]
		if !exists {
			v = &visitor{tokens: float64(rl.burst), last: now}
			rl.visitors[ip] = v
		}
		rl.refill(v, now)
		if v.tokens < 1 {
			rl.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "muitas requisicoes"})
			return
		}
		v.tokens--
		rl.mu.Unlock()
		c.Next()
	}
}
