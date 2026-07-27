package middleware

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// timeoutBody is the 504 payload, kept byte-identical to the
// gin.H{"error": "request timeout"} render this middleware used before.
const timeoutBody = `{"error":"request timeout"}`

// timeoutWriter serialises access to the real gin.ResponseWriter so the handler
// goroutine and the timeout path can never write a response at the same time.
// Whoever claims the writer first owns the response; the loser's writes are
// swallowed instead of producing "http: superfluous WriteHeader" or a body
// interleaved with the 504 already on the wire.
type timeoutWriter struct {
	gin.ResponseWriter

	mu       sync.Mutex
	timedOut bool
	hijacked bool
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return
	}
	tw.ResponseWriter.WriteHeader(code)
}

func (tw *timeoutWriter) WriteHeaderNow() {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return
	}
	tw.ResponseWriter.WriteHeaderNow()
}

func (tw *timeoutWriter) Write(b []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		// Report success: the handler is a zombie at this point and an error
		// here would only send it down an error path nobody can observe.
		return len(b), nil
	}
	return tw.ResponseWriter.Write(b)
}

func (tw *timeoutWriter) WriteString(s string) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return len(s), nil
	}
	return tw.ResponseWriter.WriteString(s)
}

func (tw *timeoutWriter) Flush() {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return
	}
	tw.ResponseWriter.Flush()
}

// Hijack records that the connection was taken over (WebSocket upgrades do
// this) so the timeout path knows there is no HTTP response left to write.
func (tw *timeoutWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	tw.mu.Lock()
	tw.hijacked = true
	tw.mu.Unlock()
	return tw.ResponseWriter.Hijack()
}

// claimTimeout marks the writer as timed out and reports whether the caller is
// allowed to emit the 504 - it is not if the handler already started writing or
// hijacked the connection.
func (tw *timeoutWriter) claimTimeout() bool {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return false
	}
	tw.timedOut = true
	return !tw.hijacked && !tw.ResponseWriter.Written()
}

// Timeout sets a per-request timeout via context.
//
// The handler chain runs on its own goroutine so a handler that ignores the
// request context cannot pin the client past the deadline. gin.Context is
// explicitly not safe for concurrent use, so on the timeout path we never render
// through the context: the 504 goes straight to the response writer, which has
// been wrapped so exactly one of the two paths can produce a body.
//
// Residual limitation: Go cannot cancel a goroutine, so a handler that ignores
// ctx.Done() runs to completion regardless. The client is released as soon as
// the complete 504 has been flushed, but this middleware then blocks until the
// handler unwinds, which means the deadline bounds the client's wait and not the
// server-side work - a wedged handler still holds its DB rows, its Docker
// connection and its HTTP connection. Returning early instead would be worse:
// gin recycles the gin.Context into its pool the moment this handler returns, so
// a zombie goroutine still walking the chain would race with, and corrupt, an
// unrelated request. Handlers that must be genuinely interruptible have to
// honour c.Request.Context() themselves.
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		tw := &timeoutWriter{ResponseWriter: c.Writer}
		c.Writer = tw

		// Buffered so the send can never block even if this goroutine stops
		// receiving: an unbuffered channel would pin the handler goroutine
		// forever, leaking it and everything its stack holds open.
		done := make(chan struct{}, 1)
		go func() {
			c.Next()
			done <- struct{}{}
		}()

		select {
		case <-done:
			// Handler finished within the deadline and owns the response.
		case <-ctx.Done():
			if tw.claimTimeout() {
				// Content-Length is set so the client sees a complete response
				// and can stop reading immediately, even though this connection
				// is not released until the handler unwinds below.
				header := tw.ResponseWriter.Header()
				header.Set("Content-Type", "application/json; charset=utf-8")
				header.Set("Content-Length", strconv.Itoa(len(timeoutBody)))
				tw.ResponseWriter.WriteHeader(http.StatusGatewayTimeout)
				_, _ = tw.ResponseWriter.WriteString(timeoutBody)
				tw.ResponseWriter.Flush()
			}
			// Wait the handler out. Anything it writes from here on is dropped
			// by timeoutWriter, and by not returning we keep gin.Context owned
			// by a single goroutine at a time and keep cancel() below from
			// firing while the handler is still using the request context.
			<-done
		}
	}
}
