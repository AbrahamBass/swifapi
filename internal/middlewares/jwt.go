package middlewares

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AbrahamBass/swiftapi/internal/types"

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
	return func(scope types.IRequestScope, handler func()) {
		authHeader, ok := scope.MetaVal("Authorization")
		if authHeader == "" || !ok {
			scope.Throw(401, "Unauthorized")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			scope.Throw(http.StatusUnauthorized, "Invalid authorization format")
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
			scope.Throw(401, "Invalid token: "+err.Error())
			return
		}

		if !token.Valid {
			scope.Throw(http.StatusUnauthorized, "Token invalid")
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if expClaim, err := claims.GetExpirationTime(); err == nil {
				if time.Now().After(expClaim.Time) {
					scope.Throw(401, "Token expired")
					return
				}
			} else {
				scope.Throw(401, "Expiration claim required")
				return
			}

			if len(jwtConfig.Issuer()) > 0 {
				iss, ok := claims["iss"].(string)
				if !ok || !contains(jwtConfig.Issuer(), iss) {
					scope.Throw(http.StatusUnauthorized, "Emisor (issuer) no permitido")
					return
				}
			}

			if len(jwtConfig.Audience()) > 0 {
				aud, ok := claims["aud"].(string)
				if !ok || !contains(jwtConfig.Audience(), aud) {
					scope.Throw(http.StatusUnauthorized, "Audiencia (audience) no permitida")
					return
				}
			}

		}

		scope.SetBaggage("claims", token.Claims)

		handler()
	}
}
