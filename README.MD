<p align="center">
<h1 align="center">httpcleint</h1>
<p align="center">Simple HTTP and REST client library for Go without dependency</p>
</p>


## Features

  * GET, POST, PUT, DELETE, HEAD. Can be extended in the future
  * Simple and chainable methods for settings and request
  * [Request] Body can be `string`, `[]byte`, `struct`, `map`, `slice` and `io.Reader` too
  * Can add any *middlewares* you want in the httpclient

## Installation

httpclient supports Go Modules. Run go mod tidy in your project's directory to collect the required packages automatically.

```bash
go get github.com/kepinsu/httpclient
```

Or with imports:

```bash
// Import httpclient into your code and refer it as `httpclient`.
import "github.com/kepinsu/httpclient"
```

## Usage

### Simple POST

```go
// Create the client
c ,err := httpclient.NewClient("")
if err !=nil{
  // handle your error
}

// My result structure
type Result struct{
  ...
}

// My error result structure if your api return a error
type ErrorResult struct{
  ...
}

var result Result
var errorresult ErrorResult


httpresponse , err :=c.Post(context.Todo(), "/endpoint", body , &result, &errorresult, 
  // Say this request is in json
  httpclient.WithIsJSON())
// handle the error or the result
...
```

### Simple POST with Multipart body

```go
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
```

### Simple Decorator

```go
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
```