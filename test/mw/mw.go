package mw

import (
	"net/http"
)

type TestMiddlewareFirst struct{}

func (mw *TestMiddlewareFirst) Middleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})
	}
}
func (mw *TestMiddlewareFirst) Description() string { return "Test Description" }
