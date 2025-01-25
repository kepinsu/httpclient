package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var ErrResponseBodyTooLarge = errors.New("httpclient: response body too large")

// Constants for some headers
const (
	userAgentHeaderKey   = "User-Agent"
	contentTypeHeaderKey = "Content-Type"
)

// NewRequest method returns the http.Request if your need to create your own request.
func (c *Client) NewRequest(
	path string,
	method string,
	body any,
	opts ...RequestOption) (*http.Request, error) {
	return c.NewRequestWithContext(
		context.Background(), path, method, body, opts...)
}

// NewRequestWithContext method returns the http.Request if your need to create your own request.
func (c *Client) NewRequestWithContext(ctx context.Context,
	path string,
	method string,
	body any,
	opts ...RequestOption) (*http.Request, error) {

	config := newRequestConfig()
	for _, o := range opts {
		o(config)
	}
	return c.newRequestWithContext(ctx, path, method, body, config)
}

// NewRequestWithContext method returns the http.Request if your need to create your own request.
func (c *Client) newRequestWithContext(ctx context.Context,
	path string,
	method string,
	body any,
	config *requestConfig) (*http.Request, error) {

	// Set the URL
	uri := c.baseURL + path

	var reader io.Reader
	// Add Body here
	switch body := body.(type) {
	case []byte:
		buffer := &bytes.Buffer{}
		buffer.Write(body)
		reader = buffer
	case nil:
	case string:
		reader = bytes.NewBufferString(body)
	case *MultipartBody:
		var (
			err      error
			boundary string
		)
		boundary, reader, err = encodeMultipart(body)
		if err != nil {
			return nil, err
		}
		if config.headers == nil {
			config.headers = http.Header{}
		}
		config.headers.Add(contentTypeHeaderKey, fmt.Sprintf("%s; boundary=%s", body.Boundary, boundary))
	default:
		if config.isJson {
			payload, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			buffer := &bytes.Buffer{}
			buffer.Write(payload)
			reader = buffer
		} else if config.isXml {
			payload, err := xml.MarshalIndent(body, "", " ")
			if err != nil {
				return nil, err
			}
			buffer := &bytes.Buffer{}
			buffer.Write(payload)
			reader = buffer
		}
	}

	// Create the http.request
	r, err := http.NewRequestWithContext(ctx, method, uri, reader)
	if err != nil {
		return nil, err
	}

	// content-type by default
	r.Header.Set(contentTypeHeaderKey, "application/text")
	if config.isJson {
		r.Header.Set(contentTypeHeaderKey, "application/json")
	} else if config.isXml {
		r.Header.Set(contentTypeHeaderKey, "application/xml")
	}

	// Add headers
	for key := range config.headers {
		r.Header.Set(key, config.headers.Get(key))
	}
	r.Header.Set(userAgentHeaderKey, c.userAgent)

	// Add queries
	var queries url.Values
	for p, v := range config.queries {
		queries.Add(p, v)
	}
	r.URL.RawQuery = queries.Encode()

	return r, nil
}

// RequestOption is to create convenient request options like wait custom fields for http.request
type RequestOption func(*requestConfig)

// requestConfig is to create http.request options
type requestConfig struct {
	isJson  bool
	isXml   bool
	headers http.Header
	queries map[string]string
}

func newRequestConfig() *requestConfig {
	return &requestConfig{
		isJson: true,
	}
}

// WithIsJson is to indicate this request/response is in json payload
func WithIsJson() RequestOption {
	return func(rc *requestConfig) {
		rc.isJson = true
		rc.isXml = false
	}
}

// WithIsXml is to indicate this request/response is in xml payload
func WithIsXml() RequestOption {
	return func(rc *requestConfig) {
		rc.isXml = true
		rc.isJson = false
	}
}

// WithHeaders is when you need to add specific headers in your request
func WithHeaders(headers http.Header) RequestOption {
	return func(rc *requestConfig) {
		rc.headers = headers
	}
}

// WithQueries is when you need to add specific query in your request
func WithQueries(
	queries map[string]string) RequestOption {
	return func(rc *requestConfig) {
		rc.queries = queries
	}
}
