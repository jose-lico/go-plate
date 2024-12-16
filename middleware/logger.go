package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

const (
	magenta = "\033[35m"
	blue    = "\033[34m"
	green   = "\033[32m"
	cyan    = "\033[36"
	yellow  = "\033[33m"
	red     = "\033[31m"

	reset = "\033[0m"
)

func ZapLoggerMiddlwareDev(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				logger.Info(fmt.Sprintf("%s %s %s %s", methodStyling(r.Method), statusStyling(ww.Status()), r.URL, r.Proto),
					zap.Int("BytesWritten", ww.BytesWritten()),
					zap.Duration("Duration", time.Since(t1)),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

func ZapLoggerMiddlwareProd(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				logger.Info("Served",
					zap.String("Method", r.Method),
					zap.String("URL", r.URL.Path),
					zap.String("Proto", r.Proto),
					zap.Int("Status", ww.Status()),
					zap.Int("BytesWritten", ww.BytesWritten()),
					zap.Duration("Duration", time.Since(t1)),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

func methodStyling(method string) string {
	if len(method) >= 7 {
		return magenta + method + reset
	}
	return magenta + method + strings.Repeat(" ", 7-len(method)) + reset

}

func statusStyling(status int) string {
	var color string

	switch {
	case status < 200:
		color = blue
	case status < 300:
		color = green
	case status < 400:
		color = cyan
	case status < 500:
		color = yellow
	default:
		color = red
	}
	return fmt.Sprintf("%s%d%s", color, status, reset)
}
