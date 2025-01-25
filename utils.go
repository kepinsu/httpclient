package httpclient

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
)

var (
	jsonCheck = regexp.MustCompile(`(?i:(application|text)/(.*json.*)(;|$))`)
	xmlCheck  = regexp.MustCompile(`(?i:(application|text)/(.*xml.*)(;|$))`)

	quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")
)

// https://github.com/golang/go/issues/51115
// [io.LimitedReader] can only return [io.EOF]
func readAllWithLimit(r io.Reader, maxSize int) ([]byte, error) {
	if maxSize <= 0 {
		return io.ReadAll(r)
	}

	var buf [512]byte // make buf stack allocated
	result := make([]byte, 0, 512)
	total := 0
	for {
		n, err := r.Read(buf[:])
		total += n
		if total > maxSize {
			return nil, ErrResponseBodyTooLarge
		}

		if err != nil {
			if err == io.EOF {
				result = append(result, buf[:n]...)
				break
			}
			return nil, err
		}

		result = append(result, buf[:n]...)
	}

	return result, nil
}

func encodeMultipart(body *MultipartBody) (string, io.Reader, error) {
	buf := bytes.NewBuffer(nil)
	w := multipart.NewWriter(buf)

	for _, mf := range body.List {
		if err := addMultipartFormField(w, mf); err != nil {
			return "", buf, err
		}
	}
	return w.Boundary(), buf, w.Close()
}

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func parseResponse(
	response any,
	resultError any,
	httpResponse *http.Response,
	payload []byte,
) error {
	// If the remote send No Content
	if httpResponse.StatusCode == http.StatusNoContent {
		return nil
	}

	contentType := httpResponse.Header.Get(contentTypeHeaderKey)
	// if the server accept
	if httpResponse.StatusCode >= http.StatusOK && httpResponse.StatusCode <= 299 {
		// Check the content-Type of the response
		if jsonCheck.MatchString(contentType) {
			return json.Unmarshal(payload, response)
		} else if xmlCheck.MatchString(contentType) {
			return xml.Unmarshal(payload, response)
		}
		return json.Unmarshal(payload, resultError)
	}

	// if the server refuse
	if httpResponse.StatusCode >= http.StatusBadRequest {
		// Check the content-Type of the response
		if jsonCheck.MatchString(contentType) {
			return json.Unmarshal(payload, resultError)
		} else if xmlCheck.MatchString(contentType) {
			return xml.Unmarshal(payload, resultError)
		}
		// If error payload is text probably default http server error
		if strings.Contains(contentType, "text/plain") {
			return fmt.Errorf("the server return %s text", payload)
		}
	}
	return nil
}
