package responses

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/AbrahamBass/swiftapi/internal/types"
)

type StreamingResponse struct {
	Stream <-chan []byte
}

type TemplateResponse struct {
	Template *template.Template
	Data     interface{}
}

type IActionResult struct {
	statusCode int
	content    interface{}
	headers    map[string]string
	cookies    []*http.Cookie
	mediaType  types.MediaType
}

func Response(statusCode int, content interface{}) *IActionResult {
	return &IActionResult{
		statusCode: statusCode,
		content:    content,
		headers:    make(map[string]string),
		cookies:    make([]*http.Cookie, 0),
	}
}

func Problem(detail string) *IActionResult {
	body := map[string]interface{}{
		"detail": detail,
		"status": http.StatusInternalServerError,
	}
	return InternalServerError(body).MtType(types.ApplicationProblemJSON)
}

func ValidationProblem(errors map[string]string) *IActionResult {
	body := map[string]interface{}{
		"errors": errors,
		"status": http.StatusUnprocessableEntity,
	}
	return UnprocessableEntity(body).MtType(types.ApplicationProblemJSON)
}

func CreatedAt(url string, content interface{}) *IActionResult {
	return Response(http.StatusCreated, content).SetHeader("Location", url)
}

func PartialContent(content []byte) *IActionResult {
	return Response(http.StatusPartialContent, content)
}

func Streamer[T any](generator func(yield func(T) error) error) <-chan []byte {
	stream := make(chan []byte)

	go func() {
		defer close(stream)

		send := func(data T) error {
			var bytes []byte
			switch v := any(data).(type) {
			case []byte:
				bytes = v
			case string:
				bytes = []byte(v)
			default:
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					return err
				}
				bytes = jsonBytes
			}

			stream <- bytes
			return nil
		}

		generator(send)
	}()

	return stream
}

func Streaming(mt types.MediaType, stream <-chan []byte) *IActionResult {

	content := StreamingResponse{
		Stream: stream,
	}

	return Response(http.StatusOK, content).
		MtType(mt)
}

func Template(dir, file string, data interface{}) *IActionResult {
	path := filepath.Join(dir, file)
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return Response(http.StatusInternalServerError, "Template not found")
	}

	content := TemplateResponse{
		Template: tmpl,
		Data:     data,
	}

	return Response(http.StatusOK, content)
}

func Html(content interface{}) *IActionResult {
	return Response(http.StatusOK, content).MtType(types.TextHTML)
}

func Throw(statusCode int, content interface{}) *IActionResult {
	return Response(statusCode, content)
}

func Ok(content interface{}) *IActionResult {
	return Response(http.StatusOK, content)
}

func Created(content interface{}) *IActionResult {
	return Response(http.StatusCreated, content)
}

func Accepted(content interface{}) *IActionResult {
	return Response(http.StatusAccepted, content)
}

func NoContent() *IActionResult {
	return Response(http.StatusNoContent, nil)
}

func MovedPermanently(url string) *IActionResult {
	return Response(http.StatusMovedPermanently, nil).SetHeader("Location", url)
}

func Found(url string) *IActionResult {
	return Response(http.StatusFound, nil).SetHeader("Location", url)
}

func NotModified() *IActionResult {
	return Response(http.StatusNotModified, nil)
}

func NotFound(content interface{}) *IActionResult {
	return Response(http.StatusNotFound, content)
}

func BadRequest(content interface{}) *IActionResult {
	return Response(http.StatusBadRequest, content)
}

func Unauthorized(content interface{}) *IActionResult {
	return Response(http.StatusUnauthorized, content)
}

func Forbidden(content interface{}) *IActionResult {
	return Response(http.StatusForbidden, content)
}

func Conflict(content interface{}) *IActionResult {
	return Response(http.StatusConflict, content)
}

func UnprocessableEntity(errors interface{}) *IActionResult {
	return Response(http.StatusUnprocessableEntity, errors)
}

func InternalServerError(content interface{}) *IActionResult {
	return Response(http.StatusInternalServerError, content)
}

func NotImplemented(content interface{}) *IActionResult {
	return Response(http.StatusNotImplemented, content)
}

func ServiceUnavailable(content interface{}) *IActionResult {
	return Response(http.StatusServiceUnavailable, content)
}

func File(content []byte, filename string) *IActionResult {
	return Ok(content).
		MtType(types.OctetStream).
		SetHeader("Content-Disposition", "attachment; filename="+filename)
}

func (a *IActionResult) SetHeader(key, value string) *IActionResult {
	a.headers[key] = value
	return a
}

func (a *IActionResult) SetCrumb(cookie *http.Cookie) *IActionResult {
	a.cookies = append(a.cookies, cookie)
	return a
}

func (a *IActionResult) MtType(mt types.MediaType) *IActionResult {
	switch mt {
	case types.ApplicationJSON, types.TextPlain, types.TextHTML, types.ApplicationXML,
		types.OctetStream, types.ApplicationForm, types.MultipartForm:
		a.mediaType = mt
	default:
		panic("invalid media type")
	}
	return a
}
