package testifyx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/AbrahamBass/swiftapi/internal/types"
)

type TestClient struct {
	handler http.Handler
	tb      testing.TB
}

func NewTestClient(app types.IApplication, tb testing.TB) *TestClient {
	return &TestClient{handler: app.Mux(), tb: tb}
}

type RequestBuilder struct {
	method   string
	endpoint string
	queries  url.Values
	headers  http.Header
	body     io.Reader
	client   *TestClient
}

func NewRequestBuilder(method, endpoint string, c *TestClient) *RequestBuilder {
	return &RequestBuilder{
		method:   method,
		endpoint: endpoint,
		queries:  make(url.Values),
		headers:  make(http.Header),
		client:   c,
	}
}

func (rb *RequestBuilder) WithQuery(key string, value interface{}) *RequestBuilder {
	rb.queries.Add(key, fmt.Sprintf("%v", value))
	return rb
}

func (rb *RequestBuilder) WithHeader(key, value string) *RequestBuilder {
	rb.headers.Add(key, value)
	return rb
}

func (rb *RequestBuilder) WithCookie(c *http.Cookie) *RequestBuilder {
	rb.headers.Add("Cookie", c.String())
	return rb
}

func (rb *RequestBuilder) WithBody(body io.Reader) *RequestBuilder {
	rb.body = body
	return rb
}

func (rb *RequestBuilder) WithFormData(data url.Values) *RequestBuilder {
	rb.body = strings.NewReader(data.Encode())
	rb.headers.Set("Content-Type", "application/x-www-form-urlencoded")
	return rb
}

func (rb *RequestBuilder) WithMultipartForm(boundary string, body io.Reader) *RequestBuilder {
	rb.body = body
	rb.headers.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	return rb
}

func (rb *RequestBuilder) Do() *HC {
	url := rb.endpoint
	if len(rb.queries) > 0 {
		url += "?" + rb.queries.Encode()
	}

	req, err := http.NewRequest(rb.method, url, rb.body)
	if err != nil {
		rb.client.tb.Fatalf("❌ Error creating request: %v", err)
	}

	req.Header = rb.headers
	rr := httptest.NewRecorder()
	rb.client.handler.ServeHTTP(rr, req)

	return &HC{
		t: rb.client.tb,
		response: &Response{
			Response:  rr.Result(),
			BodyBytes: rr.Body.Bytes(),
		},
	}
}

func (c *TestClient) request(method, endpoint string) *RequestBuilder {
	return NewRequestBuilder(method, endpoint, c)
}

func (c *TestClient) Get(path string) *RequestBuilder {
	return c.request(http.MethodGet, path)
}

func (c *TestClient) Post(path string, body io.Reader) *RequestBuilder {
	return c.request(http.MethodPost, path).WithBody(body)
}

func (c *TestClient) Put(path string, body io.Reader) *RequestBuilder {
	return c.request(http.MethodPut, path).WithBody(body)
}

func (c *TestClient) Patch(path string, body io.Reader) *RequestBuilder {
	return c.request(http.MethodPatch, path).WithBody(body)
}

func (c *TestClient) Delete(path string) *RequestBuilder {
	return c.request(http.MethodDelete, path)
}

func (c *TestClient) Head(path string) *RequestBuilder {
	return c.request(http.MethodHead, path)
}

func (c *TestClient) Options(path string) *RequestBuilder {
	return c.request(http.MethodOptions, path)
}

func (c *TestClient) PostJSON(path string, body interface{}) *RequestBuilder {
	return c.sendJSON(http.MethodPost, path, body)
}

func (c *TestClient) PutJSON(path string, body interface{}) *RequestBuilder {
	return c.sendJSON(http.MethodPut, path, body)
}

func (c *TestClient) PatchJSON(path string, body interface{}) *RequestBuilder {
	return c.sendJSON(http.MethodPatch, path, body)
}

func (c *TestClient) sendJSON(method, path string, body interface{}) *RequestBuilder {
	if body == nil {
		c.tb.Fatal("JSON body no puede ser nil")
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		c.tb.Fatalf("Error serializing JSON: %v", err)
	}

	return c.request(method, path).
		WithBody(bytes.NewReader(bodyBytes)).
		WithHeader("Content-Type", "application/json")
}
