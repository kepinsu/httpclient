package httpclient

import (
	"net/http"
)

// The response from the HTTP Request
type Response struct {
	// Raw request send by the client
	Request *http.Request
	// Raw response receive by the client
	RawResponse *http.Response
}
