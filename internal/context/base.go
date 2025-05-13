package context

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/url"

	"github.com/AbrahamBass/swiftapi/internal/responses"
	"github.com/AbrahamBass/swiftapi/internal/types"
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

func (c *Middleware) Request() *http.Request {
	return c.req
}

func (c *Middleware) Response() http.ResponseWriter {
	return c.res.W
}

func (c *Middleware) Protocol() string {
	return c.req.Method
}

func (c *Middleware) Location() *url.URL {
	return c.req.URL
}

func (c *Middleware) Pathway() string {
	return c.req.URL.Path
}

func (c *Middleware) ClientIP() string {
	return c.req.RemoteAddr
}

func (c *Middleware) Referral() string {
	return c.req.Referer()
}

func (c *Middleware) QueryVal(key string) (string, bool) {
	if value := c.req.URL.Query().Get(key); value != "" {
		return value, true
	}
	return "", false
}

func (c *Middleware) UriParam(key string) (string, bool) {
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

func (c *Middleware) MetaVal(key string) (string, bool) {
	if value := c.req.Header.Get(key); value != "" {
		return value, true
	}
	return "", false
}

func (c *Middleware) CrumbVal(name string) (*http.Cookie, bool) {
	if value, err := c.req.Cookie(name); err == nil {
		return value, true
	}
	return nil, false
}

func (c *Middleware) Respond(status int, v interface{}) {
	c.SetStatus(status)
	c.res.Send(v)
}

func (c *Middleware) Throw(status int, err interface{}) {
	c.res.MediaType = types.TextPlain
	c.res.StatusCode = status
	c.res.Send(err)
}

func (c *Middleware) SetStatus(status int) {
	c.res.StatusCode = status
}

func (c *Middleware) MediaType(mt types.MediaType) {
	c.res.MediaType = mt
}

func (c *Middleware) SetHeader(key, value string) {
	c.res.W.Header().Set(key, value)
}

func (c *Middleware) SetCrumb(cookie *http.Cookie) {
	http.SetCookie(c.res.W, cookie)
}

func (c *Middleware) SetBaggage(key string, value any) {
	ctx := context.WithValue(c.req.Context(), key, value)
	c.req = c.req.WithContext(ctx)
}

func (c *Middleware) SecureChannel() *tls.ConnectionState {
	return c.req.TLS
}

func (c *Middleware) Hostname() string {
	return c.req.Host
}

func (c *Middleware) RedirectTo(code int, url string) {
	http.Redirect(c.res.W, c.req, url, code)
}
