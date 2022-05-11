package middlewares

import "net/http"

type wrapResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
	bodySize    int
}

func (w *wrapResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	// Check after in case there's error handling in the wrapped ResponseWriter.
	if w.wroteHeader {
		return
	}
	w.statusCode = code
	w.wroteHeader = true
}

func (w *wrapResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *wrapResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	w.bodySize += len(b)
	return w.ResponseWriter.Write(b)
}
