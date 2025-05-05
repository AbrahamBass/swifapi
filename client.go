package swifapi

import (
	i "github.com/AbrahamBass/swifapi/internal"
	"github.com/AbrahamBass/swifapi/internal/responses"
	"github.com/AbrahamBass/swifapi/internal/tasks"
	"github.com/AbrahamBass/swifapi/internal/testifyx"
	"github.com/AbrahamBass/swifapi/internal/types"
	"github.com/AbrahamBass/swifapi/internal/ws"
)

type Application = types.IApplication

var NewApplication = i.NewApplication

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
	Exception           = responses.Exception
	Problem             = responses.Problem
	ValidationProblem   = responses.ValidationProblem
	CreatedAt           = responses.CreatedAt
	PartialContent      = responses.PartialContent
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
	Ipt[T any] struct {
		Value T
	}

	Hd[T any] struct {
		Value T
	}

	Pth[T any] struct {
		Value T
	}

	Qry[T any] struct {
		Value T
	}

	Ck[T any] struct {
		Value T
	}

	Bdy[T any] struct {
		Value T
	}

	Svcx[T any] struct {
		Value T
	}

	Ctx[T any] struct {
		Value T
	}
)

type UploadFile = types.UploadFile

type BackgroundTaskManager = tasks.BackgroundTaskManager

type WebsocketManager = ws.WebsocketManager

type Middleware = types.IMiddlewareContext

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
