package types

type MediaType string

const (
	ApplicationProblemJSON           = "application/problem+json"
	ApplicationJSON        MediaType = "application/json"
	TextPlain              MediaType = "text/plain"
	TextHTML               MediaType = "text/html"
	ApplicationXML         MediaType = "application/xml"
	OctetStream            MediaType = "application/octet-stream"
	ApplicationForm        MediaType = "application/x-www-form-urlencoded"
	MultipartForm          MediaType = "multipart/form-data"
)
