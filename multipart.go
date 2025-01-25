package httpclient

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
)

// MultipartBody struct represents the multipart body in HTTP request
type MultipartBody struct {
	// The Header of the Boundary like multipart/form-data
	Boundary string
	List     []MultipartField
}

// MultipartField struct represents the custom data part for a multipart request
type MultipartField struct {
	Param       string
	FileName    string
	ContentType string
	ContentID   string
	io.Reader
}

func NewMultipartBody() *MultipartBody {
	return &MultipartBody{
		List:     make([]MultipartField, 0),
		Boundary: "multipart/form-data",
	}
}

// SetMultipartFields method sets multiple data fields using [io.Reader] for multipart upload.
//
// For Example:
//
//	 m := NewMultipartBody()
//	 m.SetMultipartFields(
//			httpclient.MultipartField{
//				Param:       "uploadManifest1",
//				FileName:    "upload-file-1.json",
//				ContentType: "application/json",
//				Reader:      strings.NewReader(`{"input": {"name": "Uploaded document 1", "_filename" : ["file1.txt"]}}`),
//			},
//			httpclient.MultipartField{
//				Param:       "uploadManifest2",
//				ContentID:   "2",
//				ContentType: "application/json",
//				Reader:      strings.NewReader(`{"input": {"name": "random file"}}`),
//			})
//
// If you have a `slice` of fields already, then call-
//
//	m.SetMultipartFields.SetMultipartFields(fields...)
func (m *MultipartBody) SetMultipartFields(fields ...MultipartField) {
	m.List = append(m.List, fields...)
}

func addMultipartFormField(w *multipart.Writer, mf MultipartField) error {
	partWriter, err := w.CreatePart(createMultipartHeader(mf.Param,
		mf.FileName, mf.ContentID, mf.ContentType))
	if err != nil {
		return err
	}
	_, err = io.Copy(partWriter, mf.Reader)
	return err
}

func createMultipartHeader(param, fileName, contentID, contentType string) textproto.MIMEHeader {
	hdr := make(textproto.MIMEHeader)

	var contentDispositionValue string
	if fileName == "" {
		contentDispositionValue = fmt.Sprintf(`form-data; name="%s"`, param)
	} else {
		contentDispositionValue = fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			param, escapeQuotes(fileName))
	}
	hdr.Set("Content-Disposition", contentDispositionValue)

	if contentType != "" {
		hdr.Set(contentTypeHeaderKey, contentType)
	}
	if contentID != "" {
		hdr.Set("Content-ID", contentID)
	}
	return hdr
}
