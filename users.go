package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mawkler/go-web-server/auth"
	"github.com/mawkler/go-web-server/database"
)

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
	// TODO: move to middleware
	bearerToken := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(bearerToken, "Bearer ")
	token, err := auth.Authorize(tokenString, cfg.jwtSecret)
	if err != nil {
		log.Printf("Invalid jwt: %s", err)
		w.WriteHeader(401)
		return
	}

	idString, err := token.Claims.GetSubject()
	if err != nil {
		log.Printf("token has no subject: %s", err)
		w.WriteHeader(403)
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
		Token string `json:"token"`
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

	// TODO: is this the way?
	expiresIn := 24 * time.Hour
	if req.ExpiresInSeconds != nil {
		expiresIn = time.Duration(*req.ExpiresInSeconds) * time.Second
	}

	jwt, err := auth.CreateJwt(user.ID, cfg.jwtSecret, expiresIn)
	if err != nil {
		log.Printf("Failed to create jwt: %s", err)
		w.WriteHeader(401)
		return
	}

	res := response{
		User:  database.User{Email: user.Email, ID: user.ID},
		Token: jwt,
	}
	writeResponse(res, 200, w)
}
