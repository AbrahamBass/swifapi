package swiftapi

import (
	i "github.com/AbrahamBass/swiftapi/internal"
	"github.com/AbrahamBass/swiftapi/internal/responses"
	"github.com/AbrahamBass/swiftapi/internal/tasks"
	"github.com/AbrahamBass/swiftapi/internal/testifyx"
	"github.com/AbrahamBass/swiftapi/internal/types"
	"github.com/AbrahamBass/swiftapi/internal/ws"
)

type Application = types.IApplication

var Bootstrap = i.NewApplication

var Describe = testifyx.Describe
var Benchmark = testifyx.Benchmark
var NewTestClient = testifyx.NewTestClient

type (
	TestSuite      = testifyx.TestSuite
	TestSuiteBench = testifyx.TestSuiteBench
	TC             = testifyx.TC
	HC             = testifyx.HC
)

var (
	Response = responses.Response

	Ok                  = responses.Ok
	Created             = responses.Created
	Accepted            = responses.Accepted
	NoContent           = responses.NoContent
	MovedPermanently    = responses.MovedPermanently
	Found               = responses.Found
	NotModified         = responses.NotModified
	NotFound            = responses.NotFound
	BadRequest          = responses.BadRequest
	Unauthorized        = responses.Unauthorized
	Forbidden           = responses.Forbidden
	Conflict            = responses.Conflict
	UnprocessableEntity = responses.UnprocessableEntity
	InternalServerError = responses.InternalServerError
	NotImplemented      = responses.NotImplemented
	ServiceUnavailable  = responses.ServiceUnavailable
	File                = responses.File
	Html                = responses.Html
	Template            = responses.Template
	Throw               = responses.Throw
	Problem             = responses.Problem
	ValidationProblem   = responses.ValidationProblem
	CreatedAt           = responses.CreatedAt
	PartialContent      = responses.PartialContent
	Streaming           = responses.Streaming
)

func Streamer[T any](generator func(yield func(T) error) error) <-chan []byte {
	return responses.Streamer[T](generator)
}

type IActionResult = responses.IActionResult

type (
	APIRoute  = types.IAPIRoute
	APIRouter = types.IAPIRouter
)

type (
	Silk[T any] struct {
		Value T
	}

	Signal[T any] struct {
		Value T
	}

	Pathway[T any] struct {
		Value T
	}

	Query[T any] struct {
		Value T
	}

	Crumb[T any] struct {
		Value T
	}

	Body[T any] struct {
		Value T
	}

	Dependency[T any] struct {
		Value T
	}

	Scope[T any] struct {
		Value T
	}
)

type UploadFile = types.UploadFile

type BackgroundTaskManager = tasks.BackgroundTaskManager

type WebsocketManager = ws.WebsocketManager

type RequestScope = types.IRequestScope

type MediaType = types.MediaType

const (
	ApplicationProblemJSON MediaType = types.ApplicationProblemJSON
	ApplicationJSON        MediaType = types.ApplicationJSON
	TextPlain              MediaType = types.TextPlain
	TextHTML               MediaType = types.TextHTML
	ApplicationXML         MediaType = types.ApplicationXML
	OctetStream            MediaType = types.OctetStream
	ApplicationForm        MediaType = types.ApplicationForm
	MultipartForm          MediaType = types.MultipartForm
)
