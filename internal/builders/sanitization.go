package builders

import (
	"github.com/AbrahamBass/swifapi/internal/middlewares"
	"github.com/AbrahamBass/swifapi/internal/types"
)

type SanitizationBuilder struct {
	app types.IApplication
}

func NewSanitizationBuilder(app types.IApplication) *SanitizationBuilder {
	return &SanitizationBuilder{app: app}
}

func (hr *SanitizationBuilder) Apply() types.IApplication {
	middleware := middlewares.SanitizationMiddleware()
	hr.app.AddMiddleware(middleware)
	return hr.app
}
