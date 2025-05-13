package builders

import (
	"github.com/AbrahamBass/swiftapi/internal/middlewares"
	"github.com/AbrahamBass/swiftapi/internal/types"
)

type JWTBearer struct {
	app    types.IApplication
	config types.IJWTConfig
}

func NewJWTBearer(app types.IApplication) *JWTBearer {
	return &JWTBearer{
		app:    app,
		config: middlewares.NewJWTConfig(),
	}
}

func (jb *JWTBearer) Key(key string) types.IJWTBuilder {
	jb.config.SetKey(key)
	return jb
}

func (jb *JWTBearer) Algorithms(algorithms []string) types.IJWTBuilder {
	jb.config.SetAlgorithms(algorithms)
	return jb
}

func (jb *JWTBearer) Audience(audience []string) types.IJWTBuilder {
	jb.config.SetAudience(audience)
	return jb
}

func (jb *JWTBearer) Issue(issuer []string) types.IJWTBuilder {
	jb.config.SetIssuer(issuer)
	return jb
}

func (jb *JWTBearer) Apply() types.IApplication {
	jb.app.SetJwtConfig(jb.config)
	return jb.app
}
