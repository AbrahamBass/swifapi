package middlewares

import (
	"net/http"
	"strings"

	"github.com/AbrahamBass/swifapi/internal/types"
)

func HTTPSRedirectMiddleware() types.Middleware {
	return func(c types.IMiddlewareContext, next func()) {
		if c.TLS() == nil {
			host := strings.TrimPrefix(c.Host(), "www.")
			httpsURL := "https://" + host + c.Path()
			c.Redirect(http.StatusMovedPermanently, httpsURL)
			return
		}
		next()
	}
}
