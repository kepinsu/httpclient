package httpclient

import (
	"fmt"
	"net/http"
	"runtime"
)

// ClientsOption is to create convenient client options like wait custom RoundTripper, custom http.client
type ClientsOption func(*clientConfig)

// clientConfig is to create httpclient options
type clientConfig struct {
	Transport  http.RoundTripper
	Decorators []Decorator
	LimitSize  int
	UserAgent  string
}

func newclientConfig() *clientConfig {
	return &clientConfig{
		// Avoid to use the default Transport, it's better to use a clone
		Transport: http.DefaultTransport.(*http.Transport).Clone(),
		UserAgent: fmt.Sprintf("httpclient, v1.0.0, go %s", runtime.Version()),
	}
}

// WithTransport is to add custom Transport for http call
func WithTransport(t http.RoundTripper) ClientsOption {
	return func(cc *clientConfig) {
		if t != nil {
			cc.Transport = t
		}
	}
}

// WithUserAgent is to add user agent for any http request
func WithUserAgent(u string) ClientsOption {
	return func(cc *clientConfig) {
		cc.UserAgent = u
	}
}

// WithDecorators is to add custom Decorators for http call
// Is more like http middlewares
func WithDecorator(decorators []Decorator) ClientsOption {
	return func(cc *clientConfig) {
		cc.Decorators = decorators
	}
}

// WithDecorators is to add custom Decorators for http call
// Is more like http middlewares
func WithSizeLimit(limitSize int) ClientsOption {
	return func(cc *clientConfig) {
		cc.LimitSize = limitSize
	}
}
