package middlewares

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/AbrahamBass/swiftapi/internal/types"
)

type RateLimiter struct {
	mu          sync.Mutex
	requests    map[string]types.IRateRecord
	maxRequests int
	window      time.Duration
}

func (rl *RateLimiter) Mu() *sync.Mutex {
	return &rl.mu
}

func (rl *RateLimiter) Requests() map[string]types.IRateRecord {
	return rl.requests
}

func (rl *RateLimiter) MaxRequests() int {
	return rl.maxRequests
}

func (rl *RateLimiter) Window() time.Duration {
	return rl.window
}

func (rl *RateLimiter) SetMaxRequests(maxRequests int) {
	rl.maxRequests = maxRequests
}

func (rl *RateLimiter) SetWindow(window time.Duration) {
	rl.window = window
}

func (rl *RateLimiter) CleanupExpired() {
	for ip, record := range rl.Requests() {
		if time.Now().After(record.ResetTime()) {
			delete(rl.Requests(), ip)
		}
	}
}

type RateRecord struct {
	count     int
	resetTime time.Time
}

func (rr *RateRecord) Count() int {
	return rr.count
}

func (r *RateRecord) Increment() {
	r.count++
}

func (rr *RateRecord) ResetTime() time.Time {
	return rr.resetTime
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests:    make(map[string]types.IRateRecord),
		maxRequests: 1,
		window:      time.Second,
	}
}

func RateLimiterMiddleware(rl types.IRateLimiter) types.Middleware {
	return func(scope types.IRequestScope, next func()) {
		ip := scope.ClientIP()

		rl.Mu().Lock()
		defer rl.Mu().Unlock()

		rl.CleanupExpired()

		record, exists := rl.Requests()[ip]
		if !exists {
			record = &RateRecord{
				resetTime: time.Now().Add(rl.Window()),
			}
			rl.Requests()[ip] = record
		}

		if record.Count() >= rl.MaxRequests() {
			scope.SetHeader("Retry-After", record.ResetTime().Format(time.RFC1123))
			scope.Throw(http.StatusTooManyRequests, "Too Many Requests")
			return
		}

		record.Increment()

		scope.SetHeader("X-RateLimit-Limit", fmt.Sprintf("%d", rl.MaxRequests()))
		scope.SetHeader("X-RateLimit-Remaining", fmt.Sprintf("%d", rl.MaxRequests()-record.Count()))
		scope.SetHeader("X-RateLimit-Reset", fmt.Sprintf("%d", record.ResetTime().Unix()))

		next()
	}

}
