package swiftapi

import (
	"net/http"
	"strings"

	"github.com/AbrahamBass/swifapi/internal/types"

	"github.com/gorilla/websocket"
)

type APIWebsocketRoute struct {
	path        pathMatcher
	methods     []string
	handler     interface{}
	middlewares []types.Middleware
	wsUpgrader  websocket.Upgrader
}

func newAPIWebsocketRoute(path pathMatcher, methods []string, handler interface{}) *APIWebsocketRoute {
	return &APIWebsocketRoute{
		path:        path,
		methods:     methods,
		handler:     handler,
		middlewares: []types.Middleware{},
		wsUpgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (a *APIWebsocketRoute) Use(middlewares ...types.Middleware) {
	a.middlewares = append(a.middlewares, middlewares...)
}

func (wsr *APIWebsocketRoute) PathMatcher() pathMatcher {
	return wsr.path
}

func (wsr *APIWebsocketRoute) Methods() []string {
	return wsr.methods
}

func (wsr *APIWebsocketRoute) Handler() interface{} {
	return wsr.handler
}

func (wsr *APIWebsocketRoute) Middlewares() []types.Middleware {
	return wsr.middlewares
}

func (wsr *APIWebsocketRoute) WSUpgrader() websocket.Upgrader {
	return wsr.wsUpgrader
}

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

func (a *APIRoute) Use(middlewares ...types.Middleware) {
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

func (a *APIRouter) Use(middlewares ...types.Middleware) {
	a.middlewares = append(a.middlewares, middlewares...)
}

func (a *APIRouter) Authorization(authorization bool) {
	a.authorization = authorization
}

func (a *APIRouter) Get(
	path string,
	handler interface{},
) types.IAPIRoute {
	return a.AddRoute(
		path,
		handler,
		http.MethodGet,
	)
}

func (a *APIRouter) Post(
	path string,
	handler interface{},
) types.IAPIRoute {
	return a.AddRoute(
		path,
		handler,
		http.MethodPost,
	)
}

func (a *APIRouter) Put(
	path string,
	handler interface{},
) types.IAPIRoute {
	return a.AddRoute(
		path,
		handler,
		http.MethodPut,
	)
}

func (a *APIRouter) Patch(
	path string,
	handler interface{},
) types.IAPIRoute {
	return a.AddRoute(
		path,
		handler,
		http.MethodPatch,
	)
}

func (a *APIRouter) Delete(
	path string,
	handler interface{},
) types.IAPIRoute {
	return a.AddRoute(
		path,
		handler,
		http.MethodDelete,
	)
}

func (a *APIRouter) Head(
	path string,
	handler interface{},
) types.IAPIRoute {
	return a.AddRoute(
		path,
		handler,
		http.MethodHead,
	)
}

func (a *APIRouter) Options(
	path string,
	handler interface{},
) types.IAPIRoute {
	return a.AddRoute(
		path,
		handler,
		http.MethodOptions,
	)
}

func (a *APIRouter) Websocket(
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

	if a.prefix != "" {
		fullPath.WriteString(cleanPath(a.prefix))
	}

	if a.version != "" {
		fullPath.WriteString(cleanPath(a.version))
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
