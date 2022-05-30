//go:build unit

package middlewares

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type errReader int

func (errReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("random error")
}

func (r errReader) Close() error {
	return nil
}

func TestNewDefaultContentTypeValidator(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHandler := NewMockHandler(ctrl)
	mockResponseWriter := NewMockResponseWriter(ctrl)

	middleware := NewDefaultContentTypeValidator()
	handlerFunc := middleware(mockHandler)

	t.Run("Handle request with valid content-type and body", func(t *testing.T) {
		request := &http.Request{
			Header: map[string][]string{},
			Body:   ioutil.NopCloser(bytes.NewBuffer([]byte(`{"test": 123}`))),
		}
		request.Header.Set("Content-Type", "application/json")

		mockHandler.
			EXPECT().
			ServeHTTP(mockResponseWriter, request).
			Times(1)

		handlerFunc.ServeHTTP(mockResponseWriter, request)
	})

	t.Run("Handle request with empty body", func(t *testing.T) {
		request := &http.Request{
			Header: map[string][]string{},
			Body:   http.NoBody,
		}

		mockHandler.
			EXPECT().
			ServeHTTP(mockResponseWriter, request).
			Times(1)

		handlerFunc.ServeHTTP(mockResponseWriter, request)
	})

	t.Run("Respond bad request due empty content-type", func(t *testing.T) {
		request := &http.Request{
			Header: map[string][]string{},
			Body:   ioutil.NopCloser(bytes.NewBuffer([]byte(`{"test": 123}`))),
		}
		request.Header.Set("Content-Type", "")

		mockResponseWriter.
			EXPECT().
			WriteHeader(http.StatusBadRequest).
			Times(1)

		handlerFunc.ServeHTTP(mockResponseWriter, request)
	})

	t.Run("Respond bad request due invalid content-type", func(t *testing.T) {
		request := &http.Request{
			Header: map[string][]string{},
			Body:   ioutil.NopCloser(bytes.NewBuffer([]byte(`{"test": 123}`))),
		}
		request.Header.Set("Content-Type", "application/")

		mockResponseWriter.
			EXPECT().
			WriteHeader(http.StatusBadRequest).
			Times(1)

		handlerFunc.ServeHTTP(mockResponseWriter, request)
	})

	t.Run("Respond unsupported media type due unaccepted content-type", func(t *testing.T) {
		request := &http.Request{
			Header: map[string][]string{},
			Body:   ioutil.NopCloser(bytes.NewBuffer([]byte(`{"test": 123}`))),
		}
		request.Header.Set("Content-Type", "text/css")

		mockResponseWriter.
			EXPECT().
			WriteHeader(http.StatusUnsupportedMediaType).
			Times(1)

		handlerFunc.ServeHTTP(mockResponseWriter, request)
	})

	t.Run("Respond bad request due invalid body", func(t *testing.T) {
		request := &http.Request{
			Header: map[string][]string{},
			Body:   errReader(0),
		}
		request.Header.Set("Content-Type", "application/json")

		mockResponseWriter.
			EXPECT().
			WriteHeader(http.StatusBadRequest).
			Times(1)

		handlerFunc.ServeHTTP(mockResponseWriter, request)
	})

	t.Run("Respond bad request due invalid body when content-type is text/plain", func(t *testing.T) {
		request := &http.Request{
			Header: map[string][]string{},
			Body:   ioutil.NopCloser(bytes.NewBuffer([]byte{0xff, 0xfe, 0xfd})),
		}
		request.Header.Set("Content-Type", "text/plain")

		mockResponseWriter.
			EXPECT().
			WriteHeader(http.StatusBadRequest).
			Times(1)

		handlerFunc.ServeHTTP(mockResponseWriter, request)
	})

	t.Run("Respond bad request due invalid body when content-type is application/json", func(t *testing.T) {
		request := &http.Request{
			Header: map[string][]string{},
			Body:   ioutil.NopCloser(bytes.NewBuffer([]byte(`.{"test": 123}`))),
		}
		request.Header.Set("Content-Type", "application/json")

		mockResponseWriter.
			EXPECT().
			WriteHeader(http.StatusBadRequest).
			Times(1)

		handlerFunc.ServeHTTP(mockResponseWriter, request)
	})

	t.Run("Respond bad request due invalid body with unsupported mime type", func(t *testing.T) {
		handlerFunc := validateContentType([]string{"text/css"})(mockHandler)

		request := &http.Request{
			Header: map[string][]string{},
			Body:   ioutil.NopCloser(bytes.NewBuffer([]byte(`{"test": 123}`))),
		}
		request.Header.Set("Content-Type", "text/css")

		mockResponseWriter.
			EXPECT().
			WriteHeader(http.StatusBadRequest).
			Times(1)

		handlerFunc.ServeHTTP(mockResponseWriter, request)
	})
}

func TestNewContentTypeValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                      string
		inputAcceptedContentTypes string
		wantMiddlewareNil         bool
		wantErr                   error
	}{
		{
			"Valid content-type list",
			strings.Join(defaultContentTypes, ","),
			false,
			nil,
		},
		{
			"Error due empty content-type list",
			"",
			true,
			errors.New("invalid Content-Type "),
		},
		{
			"Error due invalid content-type",
			"text/",
			true,
			errors.New("invalid Content-Type text/"),
		},
		{
			"Error due one invalid content-type",
			"application/json,text/",
			true,
			errors.New("invalid Content-Type text/"),
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			middleware, err := NewContentTypeValidator(tt.inputAcceptedContentTypes)
			if tt.wantMiddlewareNil {
				assert.Nil(t, middleware)
			} else {
				assert.NotNil(t, middleware)
			}

			assert.Equal(t, tt.wantErr, err)
		})
	}
}
