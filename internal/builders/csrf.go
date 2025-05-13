package builders

import (
	"github.com/AbrahamBass/swiftapi/internal/middlewares"
	"github.com/AbrahamBass/swiftapi/internal/types"

	"github.com/gorilla/csrf"
)

type CSRFBuilder struct {
	app    types.IApplication
	config types.ICsrfConfig
}

func NewCSRFBuilder(app types.IApplication) *CSRFBuilder {
	return &CSRFBuilder{
		app:    app,
		config: middlewares.NewCsrfConfig(),
	}
}

func (c *CSRFBuilder) SecretKey(key string) types.ICSRFBuilder {
	c.config.SetSecretKey([]byte(key))
	return c
}

func (c *CSRFBuilder) TokenLookup(lookup string) types.ICSRFBuilder {
	c.config.SetTokenLookup(lookup)
	return c
}

func (c *CSRFBuilder) CookiePath(path string) types.ICSRFBuilder {
	c.config.SetCookiePath(path)
	return c
}

func (c *CSRFBuilder) CookieName(name string) types.ICSRFBuilder {
	c.config.SetCookieName(name)
	return c
}

func (c *CSRFBuilder) CookieDomain(domain string) types.ICSRFBuilder {
	c.config.SetCookieDomain(domain)
	return c
}

func (c *CSRFBuilder) CookieSecure(secure bool) types.ICSRFBuilder {
	c.config.SetCookieSecure(secure)
	return c
}

func (c *CSRFBuilder) CookieHTTPOnly(httpOnly bool) types.ICSRFBuilder {
	c.config.SetCookieHTTPOnly(httpOnly)
	return c
}

func (c *CSRFBuilder) CookieSameSite(sameSite csrf.SameSiteMode) types.ICSRFBuilder {
	c.config.SetCookieSameSite(sameSite)
	return c
}

func (c *CSRFBuilder) Apply() types.IApplication {
	middleware := middlewares.CsrfMiddleware(c.config)
	c.app.AddMiddleware(middleware)
	return c.app
}
