package dependencies

import (
	"net/http"
	"reflect"

	c "github.com/AbrahamBass/swiftapi/internal/context"
	"github.com/AbrahamBass/swiftapi/internal/responses"
	"github.com/AbrahamBass/swiftapi/internal/types"
	"github.com/AbrahamBass/swiftapi/internal/ws"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type issue struct {
	Loc  []string        `json:"loc"`
	Msg  string          `json:"msg"`
	Type types.IssueType `json:"type"`
}

func newIssue(location []string, message string, errorType types.IssueType) *issue {
	return &issue{
		Loc:  location,
		Msg:  message,
		Type: errorType,
	}
}

type dependencyResolver struct {
	webscoketManeger *ws.WebsocketManager
	dig              types.IDigContainer
	logger           *zap.Logger
	handler          interface{}
	req              *http.Request
	w                *responses.ResponseWriter
	body             interface{}
}

func newDependencyResolver(dig types.IDigContainer, logger *zap.Logger, handler interface{}, req *http.Request, w *responses.ResponseWriter, webscoketManeger *ws.WebsocketManager) *dependencyResolver {
	return &dependencyResolver{
		dig:              dig,
		logger:           logger,
		handler:          handler,
		req:              req,
		w:                w,
		webscoketManeger: webscoketManeger,
	}
}

func (dr *dependencyResolver) resolve() ([]reflect.Value, []*issue) {
	body, issues := extractBody(dr.req)
	if len(issues) > 0 {
		return nil, issues
	}
	dr.body = body

	depend, err := analyzeDependenciesWithCache(dr.handler)
	if err != nil {
		dr.logger.Fatal("🚨 Failed to analyze handler dependencies",
			zap.Error(err),
		)
	}

	if issues := solveDependant(dr.webscoketManeger, dr.dig, dr.logger, dr.req, dr.w, depend, dr.body); len(issues) > 0 {
		return nil, issues
	}

	return processDependant(depend), nil
}

func (dr *dependencyResolver) invoke(handler interface{}, values []reflect.Value) interface{} {
	handlerValue := reflect.ValueOf(handler)
	results := handlerValue.Call(values)

	if len(results) == 0 {
		return nil
	}

	return results[0].Interface()
}

func HTTPWrapper(
	dig types.IDigContainer,
	logger *zap.Logger,
	handler interface{},
	middlewares ...types.Middleware,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := responses.NewResponseWriter(w)
		ctx := c.NewContext(rw, r)

		fn := func(currentScope types.IRequestScope) {
			resolver := newDependencyResolver(
				dig,
				logger,
				handler,
				currentScope.Request(),
				rw,
				nil,
			)
			deps, issues := resolver.resolve()
			if issues != nil {
				rw.SetStatusCode(http.StatusUnprocessableEntity)
				rw.Send(issues)
				return
			}
			response := resolver.invoke(handler, deps)
			rw.Send(response)
		}

		chain := buildMiddlewareChain(
			fn,
			middlewares,
		)

		chain(ctx)

	}
}

func WebSocketWrapper(
	dig types.IDigContainer,
	logger *zap.Logger,
	handler interface{},
	wsUpgrader *websocket.Upgrader,
	middlewares ...types.Middleware,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := responses.NewResponseWriter(w)
		ctx := c.NewContext(rw, r)

		fn := func(currentScope types.IRequestScope) {

			wsManager := ws.NewWebsocketManager(wsUpgrader, logger)

			resolver := newDependencyResolver(
				dig,
				logger,
				handler,
				currentScope.Request(),
				rw,
				wsManager,
			)

			deps, issues := resolver.resolve()
			if issues != nil {
				rw.SetStatusCode(http.StatusUnprocessableEntity)
				rw.Send(issues)
				return
			}

			err := wsManager.Connect(w, r)
			if err != nil {
				logger.Error("Error connecting WebSocket", zap.Error(err))
				return
			}

			_ = resolver.invoke(handler, deps)
			wsManager.Start()
		}

		chain := buildMiddlewareChain(
			fn,
			middlewares,
		)

		chain(ctx)

	}
}
