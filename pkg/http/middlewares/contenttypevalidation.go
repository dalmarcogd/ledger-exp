package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"go.uber.org/zap"
)

var defaultContentTypes = []string{"application/json", "text/plain"}

// NewDefaultContentTypeValidator returns a middleware that validates if the content-type header is one of the default
// content types pre-defined and if the request body is according to the content-type header.
func NewDefaultContentTypeValidator() Middleware {
	return validateContentType(defaultContentTypes)
}

// NewContentTypeValidator returns a middleware that validates if the content-type header is one of the given
// content types, a comma separated string, and if the request body is according to the content-type header.
func NewContentTypeValidator(
	acceptedContentTypes string,
) (Middleware, error) {
	act := strings.Split(strings.ReplaceAll(acceptedContentTypes, " ", ""), ",")
	cts := make([]string, 0, len(act))

	for _, ct := range act {
		mt, _, err := mime.ParseMediaType(ct)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Type %s", ct)
		}

		cts = append(cts, mt)
	}

	return validateContentType(cts), nil
}

func validateContentType(
	acceptedContentTypes []string,
) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ctx := request.Context()

			if request.Body == http.NoBody {
				handler.ServeHTTP(writer, request)
				return
			}

			ct := request.Header.Get("Content-Type")

			if ct == "" {
				zapctx.L(ctx).Error("missing_content_type")
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			mt, _, err := mime.ParseMediaType(ct)
			if err != nil {
				zapctx.L(ctx).Error("invalid_content_type", zap.Error(err))
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			if !acceptedContentType(mt, acceptedContentTypes) {
				zapctx.L(ctx).Error("unaccepted_content_type", zap.String("content-type", mt))
				writer.WriteHeader(http.StatusUnsupportedMediaType)
				return
			}

			if !validContentBody(request, mt) {
				zapctx.L(ctx).Error("invalid_content_body")
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			handler.ServeHTTP(writer, request)
		})
	}
}

func acceptedContentType(mt string, acceptedContentTypes []string) bool {
	for _, act := range acceptedContentTypes {
		if mt == act {
			return true
		}
	}

	return false
}

func validContentBody(request *http.Request, mimeType string) bool {
	rawBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return false
	}

	var result bool
	switch mimeType {
	case "text/plain":
		result = utf8.Valid(rawBody)
	case "application/json":
		var jm json.RawMessage
		result = json.Unmarshal(rawBody, &jm) == nil
	default:
		result = false
	}

	request.Body = ioutil.NopCloser(bytes.NewBuffer(rawBody))

	return result
}
