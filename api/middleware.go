package api

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/mawkler/go-web-server/auth"
)

type contextKey string

const authorizedJWTKey contextKey = "authorizedJWT"

func (cfg *APIConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func getBearerToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	return strings.TrimPrefix(bearerToken, "Bearer ")
}

func (cfg *APIConfig) MiddlewareAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := getBearerToken(r)
		token, err := auth.Authorize(tokenString, cfg.jwtSecret)
		if err != nil {
			log.Printf("Invalid jwt: %s", err)
			w.WriteHeader(401)
			return
		}

		context := context.WithValue(r.Context(), authorizedJWTKey, token)
		r = r.WithContext(context)

		next.ServeHTTP(w, r)
	})
}
