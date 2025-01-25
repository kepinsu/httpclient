package httpclient

import (
	"bytes"
	"context"
	"io"
	"math"
	"net/http"
	"net/url"
)

// Client struct is used to create a httpclient with client-level settings,
// these settings apply to all the requests raised from the client.
type Client struct {
	// baseURL to send any Request
	baseURL string

	// HTTPclient for all call
	httpClient *http.Client

	// All Decorators
	decorators []Decorator

	// Limit size of response
	limitSize int

	// User agent for any http request
	userAgent string
}

func NewClient(baseURL string, opts ...ClientsOption) (*Client, error) {
	// Check the format of the baseURL
	_, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, err
	}
	// Create the configuration of client
	options := newclientConfig()
	options.LimitSize = math.MaxInt64
	for _, o := range opts {
		o(options)
	}

	// Provide the http.client
	httpclient := &http.Client{
		Transport: options.Transport,
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpclient,
		decorators: options.Decorators,
	}, nil
}

// SetBaseURL method sets the Base URL in the client instance. It will be used with a request
// raised from this client with a relative URL
//
//	// Setting HTTP address
//	client.SetBaseURL("http://myjeeva.com")
//
//	// Setting HTTPS address
//	client.SetBaseURL("https://myjeeva.com")
func (c *Client) SetBaseURL(url string) *Client {
	c.baseURL = url
	return c
}

// Get method does GET HTTP request. It's defined in section 4.3.1 of RFC7231.
func (c *Client) Get(
	ctx context.Context,
	path string,
	result any,
	resultError any,
	opts ...RequestOption,
) (Response, error) {
	return c.createAndDo(ctx, path, http.MethodGet, nil, result, resultError, opts...)
}

// Post method does POST HTTP request. It's defined in section 4.3.3 of RFC7231.
func (c *Client) Post(
	ctx context.Context,
	path string,
	body any,
	result any,
	resultError any,
	opts ...RequestOption,
) (Response, error) {
	return c.createAndDo(ctx, path, http.MethodPost, body, result, resultError, opts...)
}

// Post method does POST HTTP request. It's defined in section 4.3.3 of RFC7231.
func (c *Client) PostForm(
	ctx context.Context,
	path string, data url.Values,
	result any,
	resultError any,
	opts ...RequestOption) (Response, error) {

	if opts == nil {
		opts = []RequestOption{}
	}
	queries := make(map[string]string)
	for key := range data {
		queries[key] = data.Get(key)
	}
	opts = append(opts, WithQueries(queries))
	headers := http.Header{}
	headers.Set(contentTypeHeaderKey, "application/x-www-form-urlencoded")
	opts = append(opts, WithHeaders(headers))

	return c.createAndDo(ctx, path, http.MethodPost, nil, result, resultError, opts...)
}

// Delete method does DELETE HTTP request. It's defined in section 4.3.5 of RFC7231.
func (c *Client) Delete(
	ctx context.Context,
	path string,
	body any,
	result any,
	resultError any,
	opts ...RequestOption,
) (Response, error) {
	return c.createAndDo(ctx, path, http.MethodDelete, body, result, resultError, opts...)
}

// Put method does PUT HTTP request. It's defined in section 4.3.4 of RFC7231.
func (c *Client) Put(
	ctx context.Context,
	path string,
	body any,
	result any,
	resultError any,
	opts ...RequestOption,
) (Response, error) {
	return c.createAndDo(ctx, path, http.MethodPut, body, result, resultError, opts...)
}

func (c *Client) Head(
	ctx context.Context,
	path string,
	resultError any,
	opts ...RequestOption,
) (Response, error) {
	return c.createAndDo(ctx, path, http.MethodHead, nil, nil, resultError, opts...)
}

// Do method returns the http.Request if your need to do your own request.
func (c *Client) Do(r *http.Request) (*http.Response, error) {
	// Apply all Decorators pattern
	do := chain(c.httpClient, c.decorators...)
	return do.Do(r)
}

// createAndDo create the http.request and Do the request
func (c *Client) createAndDo(
	ctx context.Context,
	path string,
	method string,
	body any,
	response any,
	resultError any,
	opts ...RequestOption) (Response, error) {

	config := new(requestConfig)
	for _, o := range opts {
		o(config)
	}

	// Prepare the request
	r, err := c.newRequestWithContext(ctx, path, method, body, config)
	if err != nil {
		return Response{}, err
	}

	// Apply all Decorators pattern
	do := chain(c.httpClient, c.decorators...)
	httpresponse, err := do.Do(r)
	if err != nil {
		return Response{Request: r}, err
	}

	// Decode the body here
	rawBody, err := readAllWithLimit(httpresponse.Body, c.limitSize)
	if err != nil {
		return Response{Request: r, RawResponse: httpresponse}, err
	}
	_ = httpresponse.Body.Close()

	// remplace the response body by nop.closer
	httpresponse.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	// Check the Content-Type here
	if err := parseResponse(response, resultError, httpresponse, rawBody); err != nil {
		return Response{Request: r, RawResponse: httpresponse}, err
	}

	// clean the request config
	config = nil
	return Response{
		Request:     r,
		RawResponse: httpresponse,
	}, nil
}
