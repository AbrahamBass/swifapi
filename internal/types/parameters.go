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
	TagQuery   TagType = "Qry"
	TagPath    TagType = "Pth"
	TagHeader  TagType = "Hd"
	TagBody    TagType = "Bdy"
	TagCookie  TagType = "Ck"
	TagForm    TagType = "Ipt"
	TagService TagType = "Svcx"
	TagContext TagType = "Ctx"
)
