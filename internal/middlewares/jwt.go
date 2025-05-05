package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AbrahamBass/swifapi/internal/types"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	key        []byte
	algorithms []string
	audience   []string
	issuer     []string
}

func NewJWTConfig() *JWTConfig {
	return &JWTConfig{
		key:        []byte(""),
		algorithms: []string{"HS256"},
		audience:   []string{},
		issuer:     []string{},
	}
}

func (j *JWTConfig) Key() []byte {
	return j.key
}

func (j *JWTConfig) Algorithms() []string {
	return j.algorithms
}

func (j *JWTConfig) Audience() []string {
	return j.audience
}

func (j *JWTConfig) Issuer() []string {
	return j.issuer
}
func (j *JWTConfig) SetKey(key string) {
	j.key = []byte(key)
}

func (j *JWTConfig) SetAlgorithms(algorithms []string) {
	j.algorithms = algorithms
}

func (j *JWTConfig) SetAudience(audience []string) {
	j.audience = audience
}

func (j *JWTConfig) SetIssuer(issuer []string) {
	j.issuer = issuer
}

func JWTMiddleware(jwtConfig types.IJWTConfig) types.Middleware {
	return func(c types.IMiddlewareContext, next func()) {
		authHeader, ok := c.HdVal("Authorization")
		if authHeader == "" || !ok {
			c.Exception(401, "Unauthorized")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Exception(http.StatusUnauthorized, "Invalid authorization format")
			return
		}
		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if !contains(jwtConfig.Algorithms(), token.Method.Alg()) {
				return nil, fmt.Errorf("algoritmo de firma no permitido: %v", token.Header["alg"])
			}
			return jwtConfig.Key(), nil
		})

		if err != nil {
			c.Exception(401, "Invalid token: "+err.Error())
			return
		}

		if !token.Valid {
			c.Exception(http.StatusUnauthorized, "Token invalid")
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if expClaim, err := claims.GetExpirationTime(); err == nil {
				if time.Now().After(expClaim.Time) {
					c.Exception(401, "Token expired")
					return
				}
			} else {
				c.Exception(401, "Expiration claim required")
				return
			}

			if len(jwtConfig.Issuer()) > 0 {
				iss, ok := claims["iss"].(string)
				if !ok || !contains(jwtConfig.Issuer(), iss) {
					c.Exception(http.StatusUnauthorized, "Emisor (issuer) no permitido")
					return
				}
			}

			if len(jwtConfig.Audience()) > 0 {
				aud, ok := claims["aud"].(string)
				if !ok || !contains(jwtConfig.Audience(), aud) {
					c.Exception(http.StatusUnauthorized, "Audiencia (audience) no permitida")
					return
				}
			}

		}

		c.SetCtx("claims", token.Claims)

		next()
	}
}
