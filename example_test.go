package httpclient_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/kepinsu/httpclient"
)

func Example_get() {
	// Create a http client
	client, err := httpclient.NewClient("http://example.com")
	if err != nil {
		fmt.Println("fail to setup the client", err)
		return
	}
	client.Get(context.Background(), "", nil, nil)
}

func Example_post_Multipart() {
	// Create a http client
	client, err := httpclient.NewClient("http://example.com")
	if err != nil {
		fmt.Println("fail to setup the client", err)
		return
	}
	// prepare the multipart body
	m := httpclient.NewMultipartBody()
	m.SetMultipartFields(
		httpclient.MultipartField{
			Param:       "uploadManifest1",
			FileName:    "upload-file-1.json",
			ContentType: "application/json",
			Reader:      strings.NewReader(`{"input": {"name": "Uploaded document 1", "_filename" : ["file1.txt"]}}`),
		},
		httpclient.MultipartField{
			Param:       "uploadManifest2",
			ContentType: "application/json",
			ContentID:   "up",
			Reader:      strings.NewReader(`{"input": {"name": "random file"}}`),
		})

	// Want the response in JSON decode
	client.Post(context.Background(), "", m, nil, nil, httpclient.WithIsJson())
}

func Example_delete_WithDecorators() {

	// Random Decorator like Logging
	var logger httpclient.Decorator = func(d httpclient.Doer) httpclient.Doer {
		return httpclient.DoerFunc(func(r *http.Request) (*http.Response, error) {
			return d.Do(r)
		})
	}
	// Create a http client
	client, err := httpclient.NewClient("http://example.com",
		httpclient.WithDecorator(logger))
	if err != nil {
		fmt.Println("fail to setup the client", err)
		return
	}

	// Want the response in JSON decode
	client.Delete(context.Background(), "", nil, nil, nil, httpclient.WithIsJson())

}
