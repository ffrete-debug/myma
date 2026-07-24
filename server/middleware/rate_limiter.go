package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int
	burst    int
	window   time.Duration
}

type visitor struct {
	tokens int
	last   time.Time
}

func NewRateLimiter(rate, burst int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		burst:    burst,
		window:   window,
	}
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		now := time.Now()
		if !exists {
			v = &visitor{tokens: rl.burst, last: now}
			rl.visitors[ip] = v
		}
		elapsed := now.Sub(v.last)
		v.tokens += int(elapsed / rl.window) * rl.rate
		if v.tokens > rl.burst {
			v.tokens = rl.burst
		}
		v.last = now
		if v.tokens <= 0 {
			rl.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "muitas requisicoes"})
			return
		}
		v.tokens--
		rl.mu.Unlock()
		c.Next()
	}
}
