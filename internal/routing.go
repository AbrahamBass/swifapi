package swiftapi

import (
	"net/http"
	"strings"

	"github.com/AbrahamBass/swiftapi/internal/types"

	"github.com/gorilla/websocket"
)

type APIRoute struct {
	path        pathMatcher
	methods     []string
	handler     interface{}
	middlewares []types.Middleware
	isWebSocket bool
	wsUpgrader  *websocket.Upgrader
}

func newAPIRoute(
	path pathMatcher,
	methods []string,
	handler interface{},
	isWebSocket bool,
	wsUpgrader *websocket.Upgrader,
) *APIRoute {
	return &APIRoute{
		path:        path,
		methods:     methods,
		handler:     handler,
		middlewares: []types.Middleware{},
		isWebSocket: isWebSocket,
		wsUpgrader:  wsUpgrader,
	}
}

func (a *APIRoute) Wrap(middlewares ...types.Middleware) {
	a.middlewares = append(a.middlewares, middlewares...)
}

type APIRouter struct {
	prefix        string
	version       string
	authorization bool
	routes        []*APIRoute
	middlewares   []types.Middleware
}

func newAPIRouter() *APIRouter {
	return &APIRouter{
		authorization: false,
		routes:        []*APIRoute{},
		middlewares:   []types.Middleware{},
	}
}

func (a *APIRouter) Prefix(prefix string) {
	a.prefix = prefix
}

func (a *APIRouter) Version(version string) {
	a.version = version
}

func (a *APIRouter) Wrap(middlewares ...types.Middleware) {
	a.middlewares = append(a.middlewares, middlewares...)
}

func (a *APIRouter) Secure(authorization bool) {
	a.authorization = authorization
}

func (a *APIRouter) Handle(
	method string,
	path string,
	handler interface{},
) types.IAPIRoute {
	return a.AddRoute(
		path,
		handler,
		method,
	)
}

func (a *APIRouter) Stream(
	path string,
	handler interface{},
	origin func(r *http.Request) bool,
) types.IAPIRoute {
	return a.AddWebsocketRoute(
		path,
		handler,
		origin,
		http.MethodGet,
	)
}

func (a *APIRouter) buildRoute(path string) pathMatcher {
	var fullPath strings.Builder

	if a.version != "" {
		fullPath.WriteString(cleanPath(a.version))
	}

	if a.prefix != "" {
		fullPath.WriteString(cleanPath(a.prefix))
	}

	fullPath.WriteString(cleanPath(path))

	compiledPath := compilePattern(fullPath.String())
	return compiledPath
}

func (a *APIRouter) AddRoute(
	path string,
	handler interface{},
	methods ...string,
) types.IAPIRoute {
	compiled := a.buildRoute(path)

	route := newAPIRoute(
		compiled,
		methods,
		handler,
		false,
		nil,
	)

	a.routes = append(a.routes, route)
	return route
}

func (a *APIRouter) AddWebsocketRoute(
	path string,
	handler interface{},
	origin func(r *http.Request) bool,
	methods ...string,
) types.IAPIRoute {
	compiled := a.buildRoute(path)

	if origin == nil {
		origin = func(r *http.Request) bool { return true }
	}

	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     origin,
	}

	route := newAPIRoute(
		compiled,
		methods,
		handler,
		true,
		upgrader,
	)

	a.routes = append(a.routes, route)
	return route
}
