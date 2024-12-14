package user

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jose-lico/go-plate/database"
	"github.com/jose-lico/go-plate/middleware"

	"gorm.io/gorm"
)

type contextKey string

const IsValidated contextKey = "isValidated"
const User contextKey = "user"

func ValidateUserMiddleware(store UserStore, redis database.RedisStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), IsValidated, false)
			isAuthenticated, _ := r.Context().Value(middleware.IsAuthenticated).(bool)

			if isAuthenticated {
				session := r.Context().Value(middleware.SessionInfo).(middleware.Session)

				id := session.UserID

				u, err := store.GetUserByID(id)

				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						redis.Del(r.Context(), r.Context().Value(middleware.Token).(string))

						http.SetCookie(w, &http.Cookie{
							Name:     "session",
							Value:    "",
							Path:     "/",
							Expires:  time.Unix(0, 0),
							MaxAge:   -1,
							HttpOnly: true,
							Secure:   true,
							SameSite: http.SameSiteStrictMode,
							Domain:   "",
						})

						w.WriteHeader(http.StatusUnauthorized)
					} else {
						w.WriteHeader(http.StatusInternalServerError)
					}
				} else {
					ctx = context.WithValue(ctx, IsValidated, true)
					ctx = context.WithValue(ctx, User, u)
				}
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
