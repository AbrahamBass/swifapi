package builders

import (
	"github.com/AbrahamBass/swiftapi/internal/middlewares"
	"github.com/AbrahamBass/swiftapi/internal/types"
)

type CORSBuilder struct {
	app    types.IApplication
	config types.ICORSConfigurer
}

func NewCorsBuilder(app types.IApplication) *CORSBuilder {
	return &CORSBuilder{
		app:    app,
		config: middlewares.NewCORSConfig(),
	}
}

func (cb *CORSBuilder) Origins(origins ...string) types.ICORSBuilder {
	cb.config.SetAllowedOrigins(origins)
	return cb
}

func (cb *CORSBuilder) Methods(methods ...string) types.ICORSBuilder {
	cb.config.SetAllowedMethods(methods)
	return cb
}

func (cb *CORSBuilder) Headers(headers ...string) types.ICORSBuilder {
	cb.config.SetAllowedHeaders(headers)
	return cb
}

func (cb *CORSBuilder) Credentials(allow bool) types.ICORSBuilder {
	cb.config.SetAllowCredentials(allow)
	return cb
}

func (cb *CORSBuilder) Apply() types.IApplication {
	middleware := middlewares.CORSMiddleware(cb.config)
	cb.app.AddMiddleware(middleware)
	return cb.app
}
