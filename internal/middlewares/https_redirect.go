package middlewares

import (
	"net/http"
	"strings"

	"github.com/AbrahamBass/swiftapi/internal/types"
)

func HTTPSRedirectMiddleware() types.Middleware {
	return func(scope types.IRequestScope, handler func()) {
		if scope.SecureChannel() == nil {
			host := strings.TrimPrefix(scope.Hostname(), "www.")
			httpsURL := "https://" + host + scope.Pathway()
			scope.RedirectTo(http.StatusMovedPermanently, httpsURL)
			return
		}
		handler()
	}
}
