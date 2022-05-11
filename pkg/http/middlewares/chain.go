package middlewares

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

// Chain method chains the middlewares to execute before handler
// ref: https://gist.github.com/husobee/fd23681261a39699ee37.
func Chain(h http.Handler, m ...Middleware) http.Handler {
	if len(m) == 0 {
		return h
	}
	return m[0](Chain(h, m[1:]...))
}
