package httpclient

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

// ProblemDetails from RFC 9457
type ProblemDetails struct {
	Name     xml.Name `json:"-" xml:"problemDetails"`
	Type     string   `json:"type,omitempty" xml:"type,omitempty"`
	Title    string   `json:"title,omitempty" xml:"title,omitempty"`
	Status   int32    `json:"status,omitempty" xml:"status,omitempty"`
	Details  string   `json:"details,omitempty" xml:"details,omitempty"`
	Cause    string   `json:"cause,omitempty" xml:"cause,omitempty"`
	Instance string   `json:"instance,omitempty" xml:"instance,omitempty"`
}

// Policies body for tests
type Policies struct {
	Name xml.Name `json:"-" xml:"policies"`
	Type string   `json:"type,omitempty" xml:"type,omitempty"`
}

// Random Decorator like Logging
func WithLogger() Decorator {
	return func(d Doer) Doer {
		return DoerFunc(func(r *http.Request) (*http.Response, error) {
			return d.Do(r)
		})
	}
}

func TestNewClient(t *testing.T) {
	type args struct {
		baseURL string
		opts    []ClientsOption
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "nok case - missing base url",
			wantErr: true,
		},
		{
			name: "ok case - valid client",
			args: args{
				baseURL: "http://example.com",
				opts: []ClientsOption{
					// Max of the size of the response 100 bytes
					WithSizeLimit(100),
					// user agent test
					WithUserAgent("test"),
					WithTransport(&http.Transport{}),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.args.baseURL, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestClient_Get(t *testing.T) {

	path := "/result"
	// Prepare fake http request here
	mux := http.NewServeMux()
	// Prepare the handler
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		// return any a problemDetails here
		w.Header().Add(contentTypeHeaderKey, "application/problem+json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(ProblemDetails{
			Status:   http.StatusNotFound,
			Details:  "not found",
			Instance: r.URL.Path,
		})
	})
	s := httptest.NewServer(mux)

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	type fields struct {
		baseURL    string
		httpClient *http.Client
		decorators []Decorator
		limitSize  int
		userAgent  string
	}
	type args struct {
		ctx         context.Context
		path        string
		result      any
		resultError any
		opts        []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Response
		wantErr bool
	}{
		{
			name: "nok case - missing context.context",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
				limitSize:  500,
			},
			args: args{
				ctx: nil,
			},
			wantErr: true,
		},
		{
			name: "nok case - unreachable server",
			fields: fields{
				baseURL:    "http://10.6.6.5:80",
				httpClient: &http.Client{},
				limitSize:  500,
			},
			args: args{
				ctx:  ctx,
				path: "/unknown",
			},
			wantErr: true,
		},
		// Error body unknown format
		{
			name: "nok case - unknown endpoint",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
				limitSize:  500,
			},
			args: args{
				ctx:  context.Background(),
				path: "/unknown",
			},
			wantErr: true,
		},
		{
			name: "ok case - the endpoint return not found",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
				limitSize:  500,
			},
			args: args{
				ctx:         context.Background(),
				path:        path,
				resultError: &ProblemDetails{},
				opts: []RequestOption{
					WithIsJson(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL:    tt.fields.baseURL,
				httpClient: tt.fields.httpClient,
				decorators: tt.fields.decorators,
				limitSize:  tt.fields.limitSize,
				userAgent:  tt.fields.userAgent,
			}
			_, err := c.Get(tt.args.ctx, tt.args.path, tt.args.result, tt.args.resultError, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestClient_Post(t *testing.T) {

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	path := "/policies"

	// Prepare fake http request here
	mux := http.NewServeMux()
	// Prepare the handler
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		var policies Policies
		if err := xml.NewDecoder(r.Body).Decode(&policies); err != nil {
			// return any a problemDetails here
			w.Header().Add(contentTypeHeaderKey, "application/problem+xml")
			w.WriteHeader(http.StatusBadRequest)
			_ = xml.NewEncoder(w).Encode(ProblemDetails{
				Status:   http.StatusBadRequest,
				Details:  err.Error(),
				Instance: r.URL.Path,
			})
			return
		}
		// If the policies is a good format return 201
		w.Header().Set("Location", fmt.Sprintf("/%s", policies.Type))
		w.Header().Add(contentTypeHeaderKey, "application/xml")
		w.WriteHeader(http.StatusCreated)
		_ = xml.NewEncoder(w).Encode(policies)

	})
	s := httptest.NewServer(mux)

	type fields struct {
		baseURL    string
		httpClient *http.Client
		decorators []Decorator
		limitSize  int
		userAgent  string
	}
	type args struct {
		ctx         context.Context
		path        string
		body        any
		result      any
		resultError any
		opts        []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Response
		wantErr bool
	}{
		{
			name: "nok case - missing context.context",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
				limitSize:  500,
			},
			args: args{
				ctx: nil,
			},
			wantErr: true,
		},
		{
			name: "nok case - unreachable server",
			fields: fields{
				baseURL:    "http://10.6.6.5:80",
				httpClient: &http.Client{},
				limitSize:  500,
			},
			args: args{
				ctx:  ctx,
				path: "/unknown",
			},
			wantErr: true,
		},
		// Error body unknown format
		{
			name: "nok case - unknown endpoint",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
				limitSize:  500,
			},
			args: args{
				ctx:  context.Background(),
				path: "/unknown",
			},
			wantErr: true,
		},
		{
			name: "ok case - the endpoint return bad request because we send json payload",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
			},
			args: args{
				ctx:         context.Background(),
				path:        path,
				resultError: &ProblemDetails{},
				body: Policies{
					Type: "failure",
				},
				opts: []RequestOption{
					WithIsJson(),
				},
			},
		},
		{
			name: "ok case - the endpoint return 201 status",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
			},
			args: args{
				ctx:         context.Background(),
				path:        path,
				resultError: &ProblemDetails{},
				body: Policies{
					Type: "failure",
				},
				result: &Policies{},
				opts: []RequestOption{
					WithIsXml(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL:    tt.fields.baseURL,
				httpClient: tt.fields.httpClient,
				decorators: tt.fields.decorators,
				limitSize:  tt.fields.limitSize,
				userAgent:  tt.fields.userAgent,
			}
			_, err := c.Post(tt.args.ctx, tt.args.path, tt.args.body, tt.args.result, tt.args.resultError, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Post() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

		})
	}
}

func TestClient_PostForm(t *testing.T) {

	path := "/result"

	// Prepare fake http request here
	mux := http.NewServeMux()
	// Prepare the handler
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {

	})
	s := httptest.NewServer(mux)
	type fields struct {
		baseURL    string
		httpClient *http.Client
		decorators []Decorator
		limitSize  int
		userAgent  string
	}
	type args struct {
		ctx         context.Context
		path        string
		data        url.Values
		result      any
		resultError any
		opts        []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Response
		wantErr bool
	}{
		{
			name: "nok case - missing context.context",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
				limitSize:  500,
			},
			args: args{
				ctx: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL:    tt.fields.baseURL,
				httpClient: tt.fields.httpClient,
				decorators: tt.fields.decorators,
				limitSize:  tt.fields.limitSize,
				userAgent:  tt.fields.userAgent,
			}
			got, err := c.PostForm(tt.args.ctx, tt.args.path, tt.args.data, tt.args.result, tt.args.resultError, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.PostForm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.PostForm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Delete(t *testing.T) {
	path := "/result"

	// Prepare fake http request here
	mux := http.NewServeMux()
	// Prepare the handler
	mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	s := httptest.NewServer(mux)

	type fields struct {
		baseURL    string
		httpClient *http.Client
		decorators []Decorator
		limitSize  int
		userAgent  string
	}
	type args struct {
		ctx         context.Context
		path        string
		body        any
		result      any
		resultError any
		opts        []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Response
		wantErr bool
	}{
		{
			name: "nok case - missing context.context",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
				limitSize:  500,
			},
			args: args{
				ctx: nil,
			},
			wantErr: true,
		},
		{
			name: "ok case - the endpoint return 204 status",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
			},
			args: args{
				ctx:         context.Background(),
				path:        path,
				resultError: &ProblemDetails{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL:    tt.fields.baseURL,
				httpClient: tt.fields.httpClient,
				decorators: tt.fields.decorators,
				limitSize:  tt.fields.limitSize,
				userAgent:  tt.fields.userAgent,
			}
			_, err := c.Delete(tt.args.ctx, tt.args.path, tt.args.body, tt.args.result, tt.args.resultError, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestClient_Put(t *testing.T) {
	path := "/result"

	// Prepare fake http request here
	mux := http.NewServeMux()
	// Prepare the handler
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		// Check if we don't receive multipart request
		reader, err := r.MultipartReader()
		if err != nil {
			// return any a problemDetails here
			w.Header().Add(contentTypeHeaderKey, "application/problem+json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ProblemDetails{
				Status:   http.StatusBadRequest,
				Instance: r.URL.Path,
			})
			return
		}
		nextPart, err := reader.NextPart()
		if err != nil {
			// return any a problemDetails here
			w.Header().Add(contentTypeHeaderKey, "application/problem+json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ProblemDetails{
				Status:   http.StatusBadRequest,
				Instance: r.URL.Path,
			})
			return
		}
		if nextPart == nil {
			// return any a problemDetails here
			w.Header().Add(contentTypeHeaderKey, "application/problem+json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ProblemDetails{
				Status:   http.StatusBadRequest,
				Instance: r.URL.Path,
			})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	s := httptest.NewServer(mux)

	// Prepare multipart Body
	m := NewMultipartBody()
	m.SetMultipartFields(
		MultipartField{
			Param:       "uploadManifest1",
			FileName:    "upload-file-1.json",
			ContentType: "application/json",
			Reader:      strings.NewReader(`{"input": {"name": "Uploaded document 1", "_filename" : ["file1.txt"]}}`),
		},
		MultipartField{
			Param:       "uploadManifest2",
			ContentType: "application/json",
			ContentID:   "up",
			Reader:      strings.NewReader(`{"input": {"name": "random file"}}`),
		})

	type fields struct {
		baseURL    string
		httpClient *http.Client
		decorators []Decorator
		limitSize  int
		userAgent  string
	}
	type args struct {
		ctx         context.Context
		path        string
		body        any
		result      any
		resultError any
		opts        []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Response
		wantErr bool
	}{
		{
			name: "nok case - missing context.context",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
				limitSize:  500,
			},
			args: args{
				ctx: nil,
			},
			wantErr: true,
		},
		{
			name: "nok case - multipart request without content, the server refuse but not problemDetails in args",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
			},
			args: args{
				ctx:  context.Background(),
				path: path,
				body: NewMultipartBody(),
				// Want the response in application/json
				opts: []RequestOption{WithIsJson()},
			},
			wantErr: true,
		},
		{
			name: "ok case - multipart request without content, the server refuse and send problemDetails",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
			},
			args: args{
				ctx:         context.Background(),
				path:        path,
				body:        NewMultipartBody(),
				resultError: &ProblemDetails{},
				// Want the response in application/json
				opts: []RequestOption{WithIsJson()},
			},
		},
		{
			name: "ok case - multipart request with content, the server accept",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
				decorators: []Decorator{WithLogger()},
			},
			args: args{
				ctx:         context.Background(),
				path:        path,
				body:        m,
				resultError: &ProblemDetails{},
				// Want the response in application/json
				opts: []RequestOption{WithIsJson()},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL:    tt.fields.baseURL,
				httpClient: tt.fields.httpClient,
				decorators: tt.fields.decorators,
				limitSize:  tt.fields.limitSize,
				userAgent:  tt.fields.userAgent,
			}
			_, err := c.Put(tt.args.ctx, tt.args.path, tt.args.body, tt.args.result, tt.args.resultError, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Put() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestClient_Head(t *testing.T) {

	path := "/result"

	// Prepare fake http request here
	mux := http.NewServeMux()
	// Prepare the handler
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {

	})
	s := httptest.NewServer(mux)

	type fields struct {
		baseURL    string
		httpClient *http.Client
		decorators []Decorator
		limitSize  int
		userAgent  string
	}
	type args struct {
		ctx         context.Context
		path        string
		resultError any
		opts        []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Response
		wantErr bool
	}{
		{
			name: "nok case - missing context.context",
			fields: fields{
				baseURL:    s.URL,
				httpClient: &http.Client{},
				limitSize:  500,
			},
			args: args{
				ctx: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL:    tt.fields.baseURL,
				httpClient: tt.fields.httpClient,
				decorators: tt.fields.decorators,
				limitSize:  tt.fields.limitSize,
				userAgent:  tt.fields.userAgent,
			}
			got, err := c.Head(tt.args.ctx, tt.args.path, tt.args.resultError, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Head() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Head() = %v, want %v", got, tt.want)
			}
		})
	}
}
