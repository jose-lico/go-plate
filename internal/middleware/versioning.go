package middleware

import (
	"context"
	"net/http"
)

type version string

const Version version = "version"

func VersionURLMiddleware(version string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), Version, version)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func VersionHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := r.Header.Get("X-API-Version")
		if version == "" {
			version = "v1"
		}
		ctx := context.WithValue(r.Context(), Version, version)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
