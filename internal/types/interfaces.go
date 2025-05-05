package types

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/csrf"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

type Middleware func(ctx IMiddlewareContext, next func())

type ILogger interface {
	GetLogger() *zap.Logger
}

type IStaticFile interface {
	ILogger
	AddStaticFile(prefix, dir string)
	StaticFileExists(prefix string) bool
}

type IAPIRoute interface {
	Use(middlewares ...Middleware)
}

type IAPIWebsocketRoute interface {
	Use(middlewares ...Middleware)
}

type IAPIRouter interface {
	Get(path string, handler interface{}) IAPIRoute
	Post(path string, handler interface{}) IAPIRoute
	Put(path string, handler interface{}) IAPIRoute
	Patch(path string, handler interface{}) IAPIRoute
	Delete(path string, handler interface{}) IAPIRoute
	Head(path string, handler interface{}) IAPIRoute
	Options(path string, handler interface{}) IAPIRoute
	Websocket(path string, handler interface{}, origin func(r *http.Request) bool) IAPIRoute
	AddRoute(path string, handler interface{}, methods ...string) IAPIRoute
	AddWebsocketRoute(path string, handler interface{}, origin func(r *http.Request) bool, methods ...string) IAPIRoute
	Use(middlewares ...Middleware)
	Authorization(authorization bool)
	Prefix(prefix string)
	Version(version string)
}

type IInclude interface {
	AddRouter(rtrg func(IAPIRouter))
}

type IJWTConfig interface {
	Key() []byte
	Algorithms() []string
	Audience() []string
	Issuer() []string

	SetKey(key string)
	SetAlgorithms(algorithms []string)
	SetAudience(audience []string)
	SetIssuer(issuer []string)
}

type IJwt interface {
	SetJwtConfig(IJWTConfig)
}

type IMiddleware interface {
	AddMiddleware(Middleware)
}

type ICsrfConfig interface {
	SecretKey() []byte
	SetSecretKey([]byte)
	TokenLookup() string
	SetTokenLookup(string)
	CookiePath() string
	SetCookiePath(string)
	CookieName() string
	SetCookieName(string)
	CookieDomain() string
	SetCookieDomain(string)
	CookieSecure() bool
	SetCookieSecure(bool)
	CookieHTTPOnly() bool
	SetCookieHTTPOnly(bool)
	CookieSameSite() csrf.SameSiteMode
	SetCookieSameSite(csrf.SameSiteMode)
}

type IRateLimiter interface {
	Mu() *sync.Mutex
	Requests() map[string]IRateRecord
	MaxRequests() int
	Window() time.Duration
	SetMaxRequests(maxRequests int)
	SetWindow(window time.Duration)
	CleanupExpired()
}

type IRateRecord interface {
	Count() int
	ResetTime() time.Time
	Increment()
}

type ICORSConfigurer interface {
	AllowedOrigins() []string
	AllowedMethods() []string
	AllowedHeaders() []string
	AllowCredentials() bool
	SetAllowedOrigins([]string)
	SetAllowedMethods([]string)
	SetAllowedHeaders([]string)
	SetAllowCredentials(bool)
}

type IApplication interface {
	ILogger
	IInclude
	IJwt
	IMiddleware
	Build(port int) IApplication
	Mux() http.Handler
	Di() IContainerBuilder // Agregamos los builders
	Include() IIncludeBuilder
	CSRF() ICSRFBuilder
	JWTBearer() IJWTBuilder
	RateLimiter() IRateLimiterBuilder
	Sanitization() ISanitizationBuilder
	Cors() ICORSBuilder
	HTTPSRedirect() IHTTPSRedirectBuilder
}

type IContainerBuilder interface {
	// Métodos de Container
	Provide(instance interface{}) IContainerBuilder
	Apply() IApplication
}

type IDigContainer interface {
	Provide(constructor interface{}, opts ...dig.ProvideOption) error
	Invoke(function interface{}, opts ...dig.InvokeOption) error
}

type IIncludeBuilder interface {
	Add(rtrg func(IAPIRouter)) IIncludeBuilder
	Apply() IApplication
}

type IJWTBuilder interface {
	Key(key string) IJWTBuilder
	Algorithms(algorithms []string) IJWTBuilder
	Audience(audience []string) IJWTBuilder
	Issue(issuer []string) IJWTBuilder
	Apply() IApplication
}

type IRateLimiterBuilder interface {
	ReqLimit(maxRequests int) IRateLimiterBuilder
	Duration(window time.Duration) IRateLimiterBuilder
	Apply() IApplication
}

type ICORSBuilder interface {
	Origins(origins ...string) ICORSBuilder
	Methods(methods ...string) ICORSBuilder
	Headers(headers ...string) ICORSBuilder
	Credentials(allow bool) ICORSBuilder
	Apply() IApplication
}

type IHTTPSRedirectBuilder interface {
	Apply() IApplication
}

type ISanitizationBuilder interface {
	Apply() IApplication
}

type ICSRFBuilder interface {
	SecretKey(key string) ICSRFBuilder
	TokenLookup(lookup string) ICSRFBuilder
	CookiePath(path string) ICSRFBuilder
	CookieName(name string) ICSRFBuilder
	CookieDomain(domain string) ICSRFBuilder
	CookieSecure(secure bool) ICSRFBuilder
	CookieHTTPOnly(httpOnly bool) ICSRFBuilder
	CookieSameSite(sameSite csrf.SameSiteMode) ICSRFBuilder
	Apply() IApplication
}

type IMiddlewareContext interface {
	Req() *http.Request
	Res() http.ResponseWriter
	Method() string
	URL() *url.URL
	Path() string
	RemoteAddr() string
	Referer() string
	Qry(key string) (string, bool)
	Prm(key string) (string, bool)
	HdVal(key string) (string, bool)
	CkVal(name string) (*http.Cookie, bool)
	Response(status int, v interface{})
	Exception(status int, err interface{})
	SetStatus(status int)
	MtType(mt MediaType)
	Set(key, value string)
	SetCk(cookie *http.Cookie)
	SetCtx(key string, value any)
	TLS() *tls.ConnectionState
	Host() string
	Redirect(code int, url string)
}
