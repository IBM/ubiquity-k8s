package interfaces

import (
	"bytes"
)

type HttpClient interface {
	DoHttpRequest(path string, masks string, filters string, requestType string, requestBody *bytes.Buffer) ([]byte, error)
}
