package types

import "mime/multipart"

type UploadFile *multipart.FileHeader

type Location string

const (
	ParamLocationPath   = "path"
	ParamLocationQuery  = "query"
	ParamLocationHeader = "header"
	ParamLocationCookie = "cookie"
)

type TagType string

const (
	TagQuery   TagType = "Query"
	TagPath    TagType = "Pathway"
	TagHeader  TagType = "Signal"
	TagBody    TagType = "Body"
	TagCookie  TagType = "Crumb"
	TagForm    TagType = "Silk"
	TagService TagType = "Dependency"
	TagContext TagType = "Scope"
)
