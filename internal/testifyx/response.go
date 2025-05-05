package testifyx

import (
	"net/http"
	"testing"
)

type Response struct {
	*http.Response
	BodyBytes []byte
}

type TC struct {
	t testing.TB
}

type HC struct {
	t        testing.TB
	response *Response
}
