// Package ratelimiting provides simple per-client-IP request throttling to
// give the API a baseline defense against brute-force and abusive traffic.
package ratelimiting

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"mission-note/internal/core/response"
)

const staleAfter = 3 * time.Minute

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// PerIP returns a Gin middleware that rate-limits requests per client IP
// using a token bucket: rps sustained requests/sec, burst allowed at once.
func PerIP(rps float64, burst int) gin.HandlerFunc {
	var (
		mu       sync.Mutex
		visitors = make(map[string]*visitor)
	)

	go evictStaleVisitors(&mu, visitors)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		v, ok := visitors[ip]
		if !ok {
			v = &visitor{limiter: rate.NewLimiter(rate.Limit(rps), burst)}
			visitors[ip] = v
		}
		v.lastSeen = time.Now()
		limiter := v.limiter
		mu.Unlock()

		if !limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, "too many requests, slow down")
			c.Abort()
			return
		}

		c.Next()
	}
}

// evictStaleVisitors periodically drops idle IP entries so the map doesn't
// grow unbounded over the life of the process.
func evictStaleVisitors(mu *sync.Mutex, visitors map[string]*visitor) {
	for range time.Tick(time.Minute) {
		mu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > staleAfter {
				delete(visitors, ip)
			}
		}
		mu.Unlock()
	}
}
