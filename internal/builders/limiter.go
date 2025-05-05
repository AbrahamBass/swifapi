package builders

import (
	"time"

	"github.com/AbrahamBass/swifapi/internal/middlewares"
	"github.com/AbrahamBass/swifapi/internal/types"
)

type RateLimiterBuilder struct {
	app         types.IApplication
	rateLimiter types.IRateLimiter
}

func NewRateLimiterBuilder(app types.IApplication) *RateLimiterBuilder {
	return &RateLimiterBuilder{
		app:         app,
		rateLimiter: middlewares.NewRateLimiter(),
	}
}

func (rt *RateLimiterBuilder) ReqLimit(maxRequests int) types.IRateLimiterBuilder {
	rt.rateLimiter.SetMaxRequests(maxRequests)
	return rt
}

func (rt *RateLimiterBuilder) Duration(window time.Duration) types.IRateLimiterBuilder {
	rt.rateLimiter.SetWindow(window)
	return rt
}

func (rt *RateLimiterBuilder) Apply() types.IApplication {
	middleware := middlewares.RateLimiterMiddleware(rt.rateLimiter)
	rt.app.AddMiddleware(middleware)
	return rt.app
}
