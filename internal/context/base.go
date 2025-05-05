package context

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/AbrahamBass/swifapi/internal/responses"
	"github.com/AbrahamBass/swifapi/internal/types"
)

type Middleware struct {
	req *http.Request
	res *responses.ResponseWriter
}

func NewContext(w *responses.ResponseWriter, r *http.Request) *Middleware {
	return &Middleware{
		req: r,
		res: w,
	}
}

func (c *Middleware) Req() *http.Request {
	return c.req
}

func (c *Middleware) Res() http.ResponseWriter {
	return c.res.W
}

func (c *Middleware) Method() string {
	return c.req.Method
}

func (c *Middleware) URL() *url.URL {
	return c.req.URL
}

func (c *Middleware) Path() string {
	return c.req.URL.Path
}

func (c *Middleware) RemoteAddr() string {
	return c.req.RemoteAddr
}

func (c *Middleware) Referer() string {
	return c.req.Referer()
}

func (c *Middleware) Qry(key string) (string, bool) {
	if value := c.req.URL.Query().Get(key); value != "" {
		return value, true
	}
	return "", false
}

func (c *Middleware) Prm(key string) (string, bool) {
	if rv := c.req.Context().Value(1); rv != nil {
		mapParams, ok := rv.(map[string]string)
		if !ok {
			return "", false
		}
		if value, ok := mapParams[key]; ok {
			return value, true
		}
	}
	return "", false
}

func (c *Middleware) HdVal(key string) (string, bool) {
	if value := c.req.Header.Get(key); value != "" {
		return value, true
	}
	return "", false
}

func (c *Middleware) CkVal(name string) (*http.Cookie, bool) {
	if value, err := c.req.Cookie(name); err == nil {
		return value, true
	}
	return nil, false
}

func (c *Middleware) Response(status int, v interface{}) {
	c.SetStatus(status)
	c.res.Send(v)
}

func (c *Middleware) Exception(status int, err interface{}) {
	c.res.MediaType = types.TextPlain
	c.res.StatusCode = status
	c.res.Send(err)
}

func (c *Middleware) SetStatus(status int) {
	c.res.StatusCode = status
}

func (c *Middleware) MtType(mt types.MediaType) {
	c.res.MediaType = mt
}

func (c *Middleware) Set(key, value string) {
	c.res.W.Header().Set(key, value)
}

func (c *Middleware) SetCk(cookie *http.Cookie) {
	http.SetCookie(c.res.W, cookie)
}

func (c *Middleware) SetCtx(key string, value any) {
	ctx := context.WithValue(c.req.Context(), key, value)
	c.req = c.req.WithContext(ctx)
}

func (c *Middleware) TLS() *tls.ConnectionState {
	return c.req.TLS
}

func (c *Middleware) Host() string {
	return c.req.Host
}

func (c *Middleware) Redirect(code int, url string) {
	http.Redirect(c.res.W, c.req, url, code)
}
