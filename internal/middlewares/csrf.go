package middlewares

import (
	"net/http"
	"strings"

	"github.com/AbrahamBass/swifapi/internal/types"

	"github.com/gorilla/csrf"
)

type CsrfConfig struct {
	secretKey      []byte
	tokenLookup    string
	cookiePath     string
	cookieName     string
	cookieDomain   string
	cookieSecure   bool
	cookieHTTPOnly bool
	cookieSameSite csrf.SameSiteMode
}

// NewCsrfConfig crea una nueva configuración con valores por defecto
func NewCsrfConfig() *CsrfConfig {
	return &CsrfConfig{
		tokenLookup:    "cookie:_csrf",
		cookieName:     "_csrf",
		cookieSecure:   false,
		cookieHTTPOnly: false,
		cookieSameSite: csrf.SameSiteStrictMode,
	}
}

func (c *CsrfConfig) SecretKey() []byte {
	return c.secretKey
}

func (c *CsrfConfig) TokenLookup() string {
	return c.tokenLookup
}

func (c *CsrfConfig) CookiePath() string {
	return c.cookiePath
}

func (c *CsrfConfig) CookieName() string {
	return c.cookieName
}

func (c *CsrfConfig) CookieDomain() string {
	return c.cookieDomain
}

func (c *CsrfConfig) CookieSecure() bool {
	return c.cookieSecure
}

func (c *CsrfConfig) CookieHTTPOnly() bool {
	return c.cookieHTTPOnly
}

func (c *CsrfConfig) CookieSameSite() csrf.SameSiteMode {
	return c.cookieSameSite
}

func (c *CsrfConfig) SetSecretKey(key []byte) {
	c.secretKey = key
}

func (c *CsrfConfig) SetTokenLookup(lookup string) {
	c.tokenLookup = lookup
}

func (c *CsrfConfig) SetCookiePath(path string) {
	c.cookiePath = path
}

func (c *CsrfConfig) SetCookieName(name string) {
	c.cookieName = name
}

func (c *CsrfConfig) SetCookieDomain(domain string) {
	c.cookieDomain = domain
}

func (c *CsrfConfig) SetCookieSecure(secure bool) {
	c.cookieSecure = secure
}

func (c *CsrfConfig) SetCookieHTTPOnly(httpOnly bool) {
	c.cookieHTTPOnly = httpOnly
}

func (c *CsrfConfig) SetCookieSameSite(sameSite csrf.SameSiteMode) {
	c.cookieSameSite = sameSite
}

func CsrfMiddleware(config types.ICsrfConfig) types.Middleware {

	opts := []csrf.Option{
		csrf.CookieName(config.CookieName()),
		csrf.Path(config.CookiePath()),
		csrf.Domain(config.CookieDomain()),
		csrf.Secure(config.CookieSecure()),
		csrf.HttpOnly(config.CookieHTTPOnly()),
		csrf.SameSite(config.CookieSameSite()),
	}

	parts := strings.SplitN(config.TokenLookup(), ":", 2)
	if len(parts) != 2 {
		panic("Formato de TokenLookup inválido. Usa 'cookie:nombre', 'header:nombre', o 'form:nombre'")
	}

	sourceType := parts[0]
	sourceName := parts[1]

	switch sourceType {
	case "cookie":
		opts = append(opts, csrf.CookieName(sourceName))
	case "header":
		opts = append(opts, csrf.RequestHeader(sourceName))
	case "form":
		opts = append(opts, csrf.FieldName(sourceName))
	default:
		panic("Fuente de token CSRF no soportada: " + sourceType)
	}

	csrfMiddleware := csrf.Protect(
		config.SecretKey(),
		opts...,
	)

	return func(c types.IMiddlewareContext, next func()) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next()
		})
		csrfMiddleware(handler).ServeHTTP(c.Res(), c.Req())
	}

}
