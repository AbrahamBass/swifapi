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

type Middleware func(scope IRequestScope, handler func())

type ILogger interface {
	GetLogger() *zap.Logger
}

type IStaticFile interface {
	ILogger
	AddStaticFile(prefix, dir string)
	StaticFileExists(prefix string) bool
}

type IAPIRoute interface {
	Wrap(middlewares ...Middleware)
}

type IAPIWebsocketRoute interface {
	Wrap(middlewares ...Middleware)
}

type IAPIRouter interface {
	Stream(path string, handler interface{}, origin func(r *http.Request) bool) IAPIRoute
	Handle(method string, path string, handler interface{}) IAPIRoute
	Secure(secure bool)
	AddRoute(path string, handler interface{}, methods ...string) IAPIRoute
	AddWebsocketRoute(path string, handler interface{}, origin func(r *http.Request) bool, methods ...string) IAPIRoute
	Wrap(middlewares ...Middleware)
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

type IRequestScope interface {
	Request() *http.Request
	Response() http.ResponseWriter
	Protocol() string
	Location() *url.URL
	Pathway() string
	ClientIP() string
	Referral() string

	QueryVal(key string) (string, bool)
	UriParam(key string) (string, bool)
	MetaVal(key string) (string, bool)
	CrumbVal(name string) (*http.Cookie, bool)

	Respond(status int, v interface{})
	Throw(status int, err interface{})
	SetStatus(status int)
	MediaType(mt MediaType)

	SetHeader(key, value string)
	SetCrumb(cookie *http.Cookie)
	SetBaggage(key string, value any)

	SecureChannel() *tls.ConnectionState
	Hostname() string
	RedirectTo(code int, url string)
}
