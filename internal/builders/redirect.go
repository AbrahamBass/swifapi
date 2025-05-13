package builders

import (
	"github.com/AbrahamBass/swiftapi/internal/middlewares"
	"github.com/AbrahamBass/swiftapi/internal/types"
)

type HTTPSRedirectBuilder struct {
	app types.IApplication
}

func NewHTTPSRedirectBuilder(app types.IApplication) *HTTPSRedirectBuilder {
	return &HTTPSRedirectBuilder{app: app}
}

func (hr *HTTPSRedirectBuilder) Apply() types.IApplication {
	middleware := middlewares.HTTPSRedirectMiddleware()
	hr.app.AddMiddleware(middleware)
	return hr.app
}
