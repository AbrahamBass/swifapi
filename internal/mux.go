package swiftapi

import (
	"context"

	"github.com/AbrahamBass/swiftapi/internal/dependencies"

	md "github.com/AbrahamBass/swiftapi/internal/middlewares"

	"net/http"
	"slices"

	"github.com/AbrahamBass/swiftapi/internal/types"

	"go.uber.org/zap"
)

type Mux struct {
	dig               types.IDigContainer
	logger            *zap.Logger
	routers           []*APIRouter
	static            map[string]string
	globalMiddlewares []types.Middleware
	jwtConfig         types.IJWTConfig
}

func newMux() *Mux {
	return &Mux{}
}

func (m *Mux) SetDig(dig types.IDigContainer) {
	m.dig = dig
}

func (m *Mux) SetRouters(routers []*APIRouter) {
	m.routers = routers
}

func (m *Mux) SetLogger(Logger *zap.Logger) {
	m.logger = Logger
}

func (m *Mux) SetStaticFile(static map[string]string) {
	m.static = static
}

func (m *Mux) SetGlobalMiddlewares(middlewares []types.Middleware) {
	m.globalMiddlewares = middlewares
}

func (m *Mux) SetJwtConfig(jwtConfig types.IJWTConfig) {
	m.jwtConfig = jwtConfig
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	for _, rgrp := range m.routers {
		for _, rte := range rgrp.routes {

			match, params := rte.path.Match(req.URL.Path)
			if !match {
				continue
			}

			if !slices.Contains(rte.methods, req.Method) {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}

			clonedReq := req.Clone(context.Background())

			if params != nil {
				clonedReq = clonedReq.WithContext(
					context.WithValue(clonedReq.Context(), 1, params),
				)
			}

			if rgrp.authorization {
				if m.jwtConfig != nil {
					rgrp.middlewares = append(
						rgrp.middlewares,
						md.JWTMiddleware(m.jwtConfig),
					)
				}
			}

			combinedMiddlewares := combineMiddlewares(
				m.globalMiddlewares,
				rgrp.middlewares,
				rte.middlewares,
			)

			if rte.isWebSocket {
				dependencies.WebSocketWrapper(
					m.dig,
					m.logger,
					rte.handler,
					rte.wsUpgrader,
					combinedMiddlewares...,
				)(w, clonedReq)
				return

			}

			dependencies.HTTPWrapper(
				m.dig,
				m.logger,
				rte.handler,
				combinedMiddlewares...,
			)(w, clonedReq)
			return
		}
	}

	http.NotFound(w, req)
}
