package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-plate/internal/database"
	"go-plate/internal/utils"

	"github.com/redis/go-redis/v9"
)

type contextKey string

const IsAuthenticated contextKey = "isAuthenticated"
const SessionInfo contextKey = "sessionInfo"
const Token contextKey = "token"

type Session struct {
	UserID       int       `json:"user_id"`
	Expiration   time.Time `json:"expiration"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
}

func AuthMiddleware(rdb database.RedisStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), IsAuthenticated, false)

			cookie, err := r.Cookie("session")
			if err == nil && cookie.Value != "" {
				sessionToken := cookie.Value

				sessionJSON, err := rdb.Get(r.Context(), "session:"+sessionToken)
				if err == nil {
					var session Session
					err := json.Unmarshal([]byte(sessionJSON), &session)

					if err != nil {
						utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to unmarshal session"))
						return
					}

					ctx = context.WithValue(ctx, Token, "session:"+sessionToken)
					ctx = context.WithValue(ctx, IsAuthenticated, true)
					ctx = context.WithValue(ctx, SessionInfo, session)
				} else if err == redis.Nil {
					http.SetCookie(w, &http.Cookie{
						Name:     "session",
						Value:    "",
						Path:     "/",
						Expires:  time.Unix(0, 0),
						MaxAge:   -1,
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteStrictMode,
						Domain:   ".compared.pt",
					})
				} else {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to read session from cache"))
					return
				}
			} else {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to read cookie"))
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AuthMiddlewareBlocking(rdb database.RedisStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			cookie, err := r.Cookie("session")
			if err == nil && cookie.Value != "" {
				sessionToken := cookie.Value

				sessionJSON, err := rdb.Get(r.Context(), "session:"+sessionToken)
				if err == nil {
					var session Session
					err := json.Unmarshal([]byte(sessionJSON), &session)

					if err != nil {
						utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to unmarshal session"))
						return
					}

					ctx = context.WithValue(ctx, Token, "session:"+sessionToken)
					ctx = context.WithValue(ctx, IsAuthenticated, true)
					ctx = context.WithValue(ctx, SessionInfo, session)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				} else if err == redis.Nil {
					http.SetCookie(w, &http.Cookie{
						Name:     "session",
						Value:    "",
						Path:     "/",
						Expires:  time.Unix(0, 0),
						MaxAge:   -1,
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteStrictMode,
						Domain:   ".compared.pt",
					})
				} else {
					utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to read session from cache"))
					return
				}
			}

			http.Error(w, "No session token", http.StatusUnauthorized)
		})
	}
}
