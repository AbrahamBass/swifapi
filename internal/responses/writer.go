package responses

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/AbrahamBass/swifapi/internal/types"
)

type ResponseWriter struct {
	W             http.ResponseWriter
	StatusCode    int
	MediaType     types.MediaType
	headerWritten bool
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		W: w,
	}
}

func (rw *ResponseWriter) SetStatusCode(statusCode int) {
	rw.StatusCode = statusCode
}

func (rw *ResponseWriter) Send(v interface{}) {

	if actionResult, ok := v.(*IActionResult); ok {
		for key, value := range actionResult.headers {
			rw.W.Header().Set(key, value)
		}

		for _, cookie := range actionResult.cookies {
			http.SetCookie(rw.W, cookie)
		}

		if actionResult.mediaType != "" && rw.W.Header().Get("Content-Type") == "" {
			rw.W.Header().Set("Content-Type", string(actionResult.mediaType))
		}

		statusCode := actionResult.statusCode
		if statusCode == 0 {
			statusCode = http.StatusOK
		}
		rw.StatusCode = statusCode

		rw.Send(actionResult.content)
		return
	}

	if rw.StatusCode == 0 {
		rw.StatusCode = http.StatusOK
	}

	if rw.MediaType == "" {
		switch v.(type) {
		case string, error:
			rw.MediaType = types.TextPlain
		case []byte:
			rw.MediaType = types.OctetStream
		case TemplateResponse:
			rw.MediaType = types.TextHTML
		default:
			rw.MediaType = types.ApplicationJSON
		}
	}

	if rw.W.Header().Get("Content-Type") == "" && rw.MediaType != "" {
		rw.W.Header().Set("Content-Type", string(rw.MediaType))
	}

	switch value := v.(type) {
	case nil:
		rw.writeHeader(http.StatusNoContent)
	case string:
		rw.write([]byte(value))
	case []byte:
		rw.write(value)
	case error:
		rw.handleError(value)
	case TemplateResponse:
		rw.handleTemplate(value)
	case StreamingResponse:
		rw.handleStream(value)
	default:
		rw.handleDefault(value)
	}
}

func (rw *ResponseWriter) writeHeader(statusCode int) {
	if !rw.headerWritten {
		rw.W.WriteHeader(statusCode)
		rw.StatusCode = statusCode
		rw.headerWritten = true
	}
}

func (rw *ResponseWriter) write(data []byte) {
	rw.writeHeader(rw.StatusCode)
	if _, err := rw.W.Write(data); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func (rw *ResponseWriter) handleError(err error) {
	rw.writeHeader(http.StatusInternalServerError)
	rw.write([]byte(err.Error()))
}

func (rw *ResponseWriter) handleTemplate(t TemplateResponse) {
	rw.writeHeader(rw.StatusCode)
	if err := t.Template.Execute(rw.W, t.Data); err != nil {
		rw.handleError(fmt.Errorf("internal server error"))
	}
}

func (rw *ResponseWriter) handleStream(s StreamingResponse) {
	rw.W.Header().Set("Transfer-Encoding", "chunked")
	rw.writeHeader(http.StatusOK)

	for chunk := range s.Stream {
		if _, err := rw.W.Write(chunk); err != nil {
			break
		}
		if flusher, ok := rw.W.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}

func (rw *ResponseWriter) handleDefault(v interface{}) {
	if rw.MediaType == types.ApplicationJSON {
		rw.writeHeader(rw.StatusCode)
		if err := json.NewEncoder(rw.W).Encode(v); err != nil {
			log.Printf("JSON encoding error: %v", err)
		}
	} else {
		rw.write([]byte(fmt.Sprintf("%v", v)))
	}
}

type CustomResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Size       int
}

func NewCustomResponseWriter(w http.ResponseWriter) *CustomResponseWriter {
	return &CustomResponseWriter{
		ResponseWriter: w,
		StatusCode:     -1,
	}
}

func (w *CustomResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (cwr *CustomResponseWriter) Write(b []byte) (int, error) {
	size, err := cwr.ResponseWriter.Write(b)
	cwr.Size += size
	return size, err
}

func (cwr *CustomResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := cwr.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("underlying ResponseWriter does not implement http.Hijacker")
	}
	return hijacker.Hijack()
}

func (cwr *CustomResponseWriter) Flush() {
	if fl, ok := cwr.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (cwr *CustomResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := cwr.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (cwr *CustomResponseWriter) Written() bool {
	// Determina si ya se ha escrito alguna respuesta
	return cwr.StatusCode != 0 || cwr.Size > 0
}
