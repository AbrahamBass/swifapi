package middlewares

import (
	"github.com/AbrahamBass/swifapi/internal/types"

	"github.com/microcosm-cc/bluemonday"
)

func SanitizationMiddleware() types.Middleware {
	return func(c types.IMiddlewareContext, next func()) {
		r := c.Req()

		p := bluemonday.StrictPolicy()

		sanitizeQueryParams(r, p)
		sanitizeCookies(r, p)
		sanitizeHeaders(r, p)
		sanitizeRequestBody(r, p)
		sanitizeFragment(r, p)
		sanitizeFormData(r, p)

	}
}
