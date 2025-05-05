package swiftapi

import (
	"fmt"
	"net/http"

	"github.com/AbrahamBass/swifapi/internal/builders"
	"github.com/AbrahamBass/swifapi/internal/logger"
	"github.com/AbrahamBass/swifapi/internal/middlewares"
	"github.com/AbrahamBass/swifapi/internal/tasks"
	"github.com/AbrahamBass/swifapi/internal/types"

	"go.uber.org/dig"
	"go.uber.org/zap"
)

type Application struct {
	logger            *zap.Logger
	di                types.IDigContainer
	routers           []*APIRouter
	staticFiles       map[string]string
	globalMiddlewares []types.Middleware
	jwtConfig         types.IJWTConfig
}

func NewApplication() *Application {
	application := &Application{
		routers:           []*APIRouter{},
		staticFiles:       map[string]string{},
		globalMiddlewares: []types.Middleware{},
	}
	application.logger = logger.NewZapLogger()
	application.di = dig.New()
	return application
}

func (s *Application) GetLogger() *zap.Logger {
	return s.logger
}

func (s *Application) AddRouter(rtrg func(types.IAPIRouter)) {
	baseRouter := newAPIRouter()
	rtrg(baseRouter)
	s.routers = append(s.routers, baseRouter)
}

func (s *Application) AddMiddleware(middleware types.Middleware) {
	s.globalMiddlewares = append(s.globalMiddlewares, middleware)
}

func (s *Application) SetJwtConfig(jwtConfig types.IJWTConfig) {
	s.jwtConfig = jwtConfig
}

func (s *Application) Di() types.IContainerBuilder {
	return builders.NewDi(s, s.di)
}

func (s *Application) CSRF() types.ICSRFBuilder {
	return builders.NewCSRFBuilder(s)
}

func (s *Application) Include() types.IIncludeBuilder {
	return builders.NewInclude(s)
}

func (s *Application) JWTBearer() types.IJWTBuilder {
	return builders.NewJWTBearer(s)
}

func (s *Application) RateLimiter() types.IRateLimiterBuilder {
	return builders.NewRateLimiterBuilder(s)
}

func (s *Application) Cors() types.ICORSBuilder {
	return builders.NewCorsBuilder(s)
}

func (s *Application) HTTPSRedirect() types.IHTTPSRedirectBuilder {
	return builders.NewHTTPSRedirectBuilder(s)
}

func (s *Application) Sanitization() types.ISanitizationBuilder {
	return builders.NewSanitizationBuilder(s)
}

func (s *Application) Mux() http.Handler {
	defer s.logger.Sync()

	mux := newMux()
	mux.SetDig(s.di)
	mux.SetLogger(s.logger)
	mux.SetRouters(s.routers)
	mux.SetStaticFile(s.staticFiles)
	mux.SetGlobalMiddlewares(s.globalMiddlewares)
	mux.SetJwtConfig(s.jwtConfig)

	_ = s.Di().
		Provide(s.GetLogger).
		Provide(tasks.NewBackgroundTaskManager)

	return mux
}

func (s *Application) Build(port int) types.IApplication {
	addr := fmt.Sprintf(":%d", port)

	server := &http.Server{
		Addr:    addr,
		Handler: middlewares.LoggingMiddleware(s.logger)(s.Mux()),
	}

	s.logger.Info("🚀 Server ready!",
		zap.String("url", fmt.Sprintf("http://localhost:%d", port)),
		zap.Int("port", port),
	)

	if err := server.ListenAndServe(); err != nil {
		s.logger.Fatal("🔥 Server failed to start",
			zap.Error(err),
			zap.Int("port", port),
		)
	}

	return s
}
