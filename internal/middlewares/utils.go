package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

func isMethodAllowed(method string, allowedMethods []string) bool {
	for _, m := range allowedMethods {
		if strings.ToUpper(m) == strings.ToUpper(method) {
			return true
		}
	}
	return false
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func deepDecode(s string) string {
	const maxDecodeDepth = 5
	return deepDecodeRecursive(s, 0)
}

func deepDecodeRecursive(s string, depth int) string {
	if depth >= 5 {
		return s
	}

	decoded, err := url.QueryUnescape(s)
	if err != nil || decoded == s {
		return decoded
	}

	return deepDecodeRecursive(decoded, depth+1)
}

func sanitizeQueryParams(r *http.Request, p *bluemonday.Policy) {
	query := r.URL.Query()
	for param, values := range query {
		for i, v := range values {
			decoded := deepDecode(v)
			sanitized := p.Sanitize(decoded)
			query[param][i] = url.QueryEscape(sanitized) // Re-encode seguro
		}
	}
	r.URL.RawQuery = query.Encode()
}

func sanitizeCookies(r *http.Request, p *bluemonday.Policy) {
	var sanitizedCookies []*http.Cookie
	for _, cookie := range r.Cookies() {
		decodedValue := deepDecode(cookie.Value)
		sanitized := p.Sanitize(decodedValue)

		newCookie := *cookie
		newCookie.Value = url.QueryEscape(sanitized) // Preservar encoding válido
		sanitizedCookies = append(sanitizedCookies, &newCookie)
	}

	r.Header.Del("Cookie")
	for _, cookie := range sanitizedCookies {
		r.AddCookie(cookie)
	}
}

func sanitizeFragment(r *http.Request, p *bluemonday.Policy) {
	if frag := r.URL.Fragment; frag != "" {
		decoded := deepDecode(frag)
		sanitized := p.Sanitize(decoded)
		r.URL.Fragment = url.QueryEscape(sanitized)
	}
}

func sanitizeHeaders(r *http.Request, p *bluemonday.Policy) {
	newHeaders := make(http.Header)
	for key, values := range r.Header {
		sanitizedValues := make([]string, len(values))
		for i, v := range values {
			decoded := deepDecode(v)
			sanitizedValues[i] = p.Sanitize(decoded)
		}
		newHeaders[key] = sanitizedValues
	}
	r.Header = newHeaders
}

func sanitizeRequestBody(r *http.Request, p *bluemonday.Policy) {
	// Preservar cuerpo original
	buf := &bytes.Buffer{}
	tee := io.TeeReader(r.Body, buf)
	originalBody, _ := io.ReadAll(tee)
	r.Body.Close()

	contentType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))

	switch {
	case strings.HasPrefix(contentType, "application/json"):
		sanitizeJSONBody(originalBody, p, r)

	case strings.HasPrefix(contentType, "text/html"),
		strings.HasPrefix(contentType, "application/xml"):
		sanitized := p.SanitizeBytes(originalBody)
		r.Body = io.NopCloser(bytes.NewReader(sanitized))

	case strings.HasPrefix(contentType, "image/"),
		strings.HasPrefix(contentType, "video/"),
		strings.HasPrefix(contentType, "application/octet-stream"):
		r.Body = io.NopCloser(bytes.NewReader(originalBody))

	default:
		sanitized := p.Sanitize(string(originalBody))
		r.Body = io.NopCloser(strings.NewReader(sanitized))
	}
}

func sanitizeFormData(r *http.Request, p *bluemonday.Policy) {
	contentType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))

	// Caso URL-encoded
	if contentType == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err == nil {
			for key, values := range r.PostForm {
				for i, v := range values {
					decoded := deepDecode(v)
					sanitized := p.Sanitize(decoded)
					r.PostForm[key][i] = url.QueryEscape(sanitized)
				}
			}
			r.Body = io.NopCloser(strings.NewReader(r.PostForm.Encode()))
		}
	}

	// Caso multipart
	if strings.HasPrefix(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err == nil {
			for key, values := range r.MultipartForm.Value {
				for i, v := range values {
					decoded := deepDecode(v)
					sanitized := p.Sanitize(decoded)
					r.MultipartForm.Value[key][i] = sanitized
				}
			}
		}
	}
}

func sanitizeJSONBody(body []byte, p *bluemonday.Policy, r *http.Request) {
	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return
	}

	sanitizedData := sanitizeJSON(data, p)
	newBody, _ := json.Marshal(sanitizedData)
	r.Body = io.NopCloser(bytes.NewReader(newBody))
}

func sanitizeJSON(data interface{}, p *bluemonday.Policy) interface{} {
	switch v := data.(type) {
	case string:
		decoded := deepDecode(v)
		return p.Sanitize(decoded)
	case map[string]interface{}:
		for key, val := range v {
			v[key] = sanitizeJSON(val, p)
		}
		return v
	case []interface{}:
		for i, item := range v {
			v[i] = sanitizeJSON(item, p)
		}
		return v
	default:
		return v
	}
}
