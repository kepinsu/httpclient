package httpclient

import (
	"encoding/xml"
	"net/http"
	"testing"
)

func TestClient_NewRequest(t *testing.T) {
	type fields struct {
		baseURL   string
		userAgent string
	}
	type args struct {
		path   string
		method string
		body   any
		opts   []RequestOption
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "nok case - unknown method",
			args: args{
				method: "unknown method",
			},
			wantErr: true,
		},
		{
			name: "nok case - post method with xml unsupported type",
			args: args{
				method: http.MethodDelete,
				path:   "/delete",
				body: struct {
					ID string `xml:"id"`
				}{
					ID: "to be delete",
				},
				opts: []RequestOption{
					WithIsXml(),
				},
			},
			wantErr: true,
		},
		{
			name: "ok case - delete method",
			args: args{
				method: http.MethodDelete,
				path:   "/delete",
				body: struct {
					ID string `json:"id"`
				}{
					ID: "to be delete",
				},
				opts: []RequestOption{
					WithIsJson(),
				},
			},
		},
		{
			name: "ok case - post method with xml data",
			args: args{
				method: http.MethodPost,
				path:   "/delete",
				body: struct {
					XMLName xml.Name `xml:"body"`
					Name    string
				}{
					Name: "to be delete",
				},
				opts: []RequestOption{
					WithIsXml(),
				},
			},
		},
		{
			name: "ok case - post method with string",
			args: args{
				method: http.MethodPost,
				path:   "/delete",
				body:   "fdsjkfhsd",
				opts: []RequestOption{
					WithHeaders(http.Header{
						"nay": []string{"fr"},
					}),
				},
			},
		},
		{
			name: "ok case - put method with string",
			args: args{
				method: http.MethodPut,
				path:   "/delete",
				body:   []byte("fdsjkfhsd"),
				opts:   []RequestOption{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				baseURL:   tt.fields.baseURL,
				userAgent: tt.fields.userAgent,
			}
			_, err := c.NewRequest(tt.args.path, tt.args.method, tt.args.body, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.NewRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
