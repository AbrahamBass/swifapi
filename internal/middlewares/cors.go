package middlewares

import (
	"net/http"
	"strings"

	"github.com/AbrahamBass/swiftapi/internal/types"
)

type CORSConfig struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	allowCredentials bool
}

func (c *CORSConfig) AllowedOrigins() []string {
	return c.allowedOrigins
}

func (c *CORSConfig) AllowedMethods() []string {
	return c.allowedMethods
}

func (c *CORSConfig) AllowedHeaders() []string {
	return c.allowedHeaders
}

func (c *CORSConfig) AllowCredentials() bool {
	return c.allowCredentials
}

func (c *CORSConfig) SetAllowedOrigins(origins []string) {
	c.allowedOrigins = origins
}

func (c *CORSConfig) SetAllowedMethods(methods []string) {
	c.allowedMethods = methods
}

func (c *CORSConfig) SetAllowedHeaders(headers []string) {
	c.allowedHeaders = headers
}

func (c *CORSConfig) SetAllowCredentials(allow bool) {
	c.allowCredentials = allow
}

func NewCORSConfig() *CORSConfig {
	return &CORSConfig{
		allowedOrigins:   []string{"*"},
		allowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		allowedHeaders:   []string{"Content-Type"},
		allowCredentials: false,
	}
}

func CORSMiddleware(config types.ICORSConfigurer) types.Middleware {
	return func(scope types.IRequestScope, handler func()) {
		origin, _ := scope.MetaVal("Origin")
		if isOriginAllowed(origin, config.AllowedOrigins()) {
			scope.SetHeader("Access-Control-Allow-Origin", origin)
			if config.AllowCredentials() {
				scope.SetHeader("Access-Control-Allow-Credentials", "true")
			}
		}

		if scope.Protocol() == "OPTIONS" {
			scope.SetHeader("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods(), ", "))
			scope.SetHeader("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders(), ", "))
			scope.Respond(http.StatusNoContent, nil)
			return
		}

		if !isMethodAllowed(scope.Protocol(), config.AllowedMethods()) {
			scope.Throw(http.StatusMethodNotAllowed, "Method Not Allowed")
			return
		}

		handler()
	}
}
