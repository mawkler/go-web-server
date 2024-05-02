package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/mawkler/go-web-server/auth"
	"github.com/mawkler/go-web-server/database"
)

type contextKey string

const authorizedJWTKey contextKey = "authorizedJWT"

func getBearerToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	return strings.TrimPrefix(bearerToken, "Bearer ")
}

func (cfg *apiConfig) middlewareAuthorization(next http.Handler) http.Handler {
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

func (cfg *apiConfig) handlerGetUsers(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.DB.GetUsers()
	if err != nil {
		w.WriteHeader(500)
	}
	writeResponse(chirps, 200, w)
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResponse(errResponse{Error: "Invalid JSON body"}, 400, w)
		return
	}

	user, err := cfg.DB.CreateUser(req.Email, req.Password)
	if err != nil {
		log.Printf("Failed to create user %s: %s", req.Email, err)
		w.WriteHeader(500)
		return
	}
	writeResponse(user, 201, w)
}

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	token, ok := r.Context().Value(authorizedJWTKey).(*jwt.Token)
	if !ok {
		log.Printf("context does not contain authorized access token")
		w.WriteHeader(500)
		return
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		log.Printf("failed to get issuer from access token: %s", err)
		w.WriteHeader(401)
		return
	}

	if issuer != "chirpy-access" {
		log.Printf("jwt is not an access token. Issuer: %s", issuer)
		w.WriteHeader(401)
		return
	}

	idString, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("token has no subject: %s", err)
		w.WriteHeader(401)
		return
	}

	id, err := strconv.Atoi(idString)
	if err != nil {
		log.Print("JWT subject (user ID) is non-numeric")
		w.WriteHeader(403)
		return
	}

	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	req := request{}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResponse(errResponse{Error: "Invalid JSON body"}, 400, w)
		return
	}

	user, err := cfg.DB.UpdateUser(id, req.Email, req.Password)
	if err != nil {
		log.Printf("Failed to update user: %s", err)
		w.WriteHeader(500)
		return
	}
	if user == nil {
		writeResponse("User not found", 404, w)
	} else {
		writeResponse(user, 200, w)
	}
}

func (cfg *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeResponse("Query parameter `id` is non-numeric", 400, w)
		return
	}

	chirp, err := cfg.DB.GetUser(id)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	if chirp == nil {
		w.WriteHeader(404)
		return
	}

	writeResponse(chirp, 200, w)

}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type request struct {
		ExpiresInSeconds *int   `json:"expires_in_seconds"`
		Email            string `json:"email"`
		Password         string `json:"password"`
	}

	type response struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		database.User
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResponse(errResponse{Error: "Invalid JSON body"}, 400, w)
		return
	}

	user, err := cfg.DB.Login(req.Email, req.Password)
	if err != nil {
		log.Printf("Unauthenticated: %s", err)
		w.WriteHeader(401)
		return

	}

	if user == nil {
		log.Printf("Unauthenticated, user %s does not exist", req.Email)
		w.WriteHeader(401)
		return
	}

	accessToken, err := auth.CreateAccessToken(user.ID, cfg.jwtSecret, req.ExpiresInSeconds)
	if err != nil {
		log.Printf("Failed to create access token: %s", err)
		w.WriteHeader(401)
		return
	}

	refreshToken, err := auth.CreateRefreshToken(user.ID, cfg.jwtSecret)
	if err != nil {
		log.Printf("Failed to create refresh token: %s", err)
		w.WriteHeader(401)
		return
	}

	res := response{
		User:         database.User{Email: user.Email, ID: user.ID},
		Token:        accessToken,
		RefreshToken: refreshToken,
	}
	writeResponse(res, 200, w)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, ok := r.Context().Value(authorizedJWTKey).(*jwt.Token)
	if !ok {
		log.Printf("context does not contain authorized access token")
		w.WriteHeader(500)
		return
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		log.Printf("failed to get issuer from refresh token: %s", err)
		w.WriteHeader(401)
		return
	}

	if issuer != "chirpy-refresh" {
		log.Printf("jwt is not a refresh token")
		w.WriteHeader(401)
		return
	}

	// TODO
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	token, ok := r.Context().Value(authorizedJWTKey).(*jwt.Token)
	if !ok {
		log.Printf("context does not contain authorized access token")
		w.WriteHeader(500)
		return
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		log.Printf("failed to get issuer from refresh token: %s", err)
		w.WriteHeader(401)
		return
	}

	if issuer != "chirpy-refresh" {
		log.Printf("jwt is not a refresh token")
		w.WriteHeader(401)
		return
	}

	revocations, err := cfg.DB.GetRevocations()
	if err != nil {
		log.Printf("failed to get revocations: %s", err)
		w.WriteHeader(500)
		return
	}

	bearerTokenString := getBearerToken(r)
	for _, revocation := range revocations {
		if revocation == bearerTokenString {
			writeResponse("refresh token has been revoked", 401, w)
			return
		}
	}

	userID, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("token has no subject: %s", err)
		w.WriteHeader(401)
		return
	}

	id, err := strconv.Atoi(userID)
	if err != nil {
		log.Print("JWT subject (user ID) is non-numeric")
		w.WriteHeader(401)
		return
	}

	expiresInSeconds := 60 * 60
	refreshToken, err := auth.CreateAccessToken(id, cfg.jwtSecret, &expiresInSeconds)
	if err != nil {
		log.Printf("Failed to create access token: %s", err)
		w.WriteHeader(401)
		return
	}

	writeResponse(response{Token: refreshToken}, 200, w)
}
