package middlewares

import (
	"github.com/AbrahamBass/swiftapi/internal/types"

	"github.com/microcosm-cc/bluemonday"
)

func SanitizationMiddleware() types.Middleware {
	return func(scope types.IRequestScope, handler func()) {
		r := scope.Request()

		p := bluemonday.StrictPolicy()

		sanitizeQueryParams(r, p)
		sanitizeCookies(r, p)
		sanitizeHeaders(r, p)
		sanitizeRequestBody(r, p)
		sanitizeFragment(r, p)
		sanitizeFormData(r, p)

		handler()
	}
}
