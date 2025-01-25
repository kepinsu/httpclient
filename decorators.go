package httpclient

import "net/http"

type Decorator func(Doer) Doer

// Interface for HTTP.Client
type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

// DoerFunc is an adapter to allow the use of ordinary functions
// as HTTP handlers. If f is a function with the appropriate signture
type DoerFunc func(*http.Request) (*http.Response, error)

// The do function allows the clientsFunc to satisafy the Doer interface
func (f DoerFunc) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}

// chain builds a Decorator composed of an inline middleware stack and endpoint
// handler in the order they are passed.
func chain(endpoint Doer, middlewares ...Decorator) Doer {
	// Return ahead of time if there aren't any middlewares for the chain
	if len(middlewares) == 0 {
		return endpoint
	}

	// Wrap the end handler with the middleware chain
	h := middlewares[len(middlewares)-1](endpoint)
	for i := len(middlewares) - 2; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}
