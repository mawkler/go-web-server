package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

func writeResponse[T any](response T, code int, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/json")
	resp, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
	} else {
		w.WriteHeader(code)
		w.Write(resp)
	}
}

func (cfg *APIConfig) HandlerCreateUser(w http.ResponseWriter, r *http.Request) {
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

	user, err := cfg.DB.CreateUser(req.Email, req.Password, false)
	if err != nil {
		log.Printf("Failed to create user %s: %s", req.Email, err)
		w.WriteHeader(500)
		return
	}
	writeResponse(user, 201, w)
}

func (cfg *APIConfig) HandlerGetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeResponse("Query parameter `id` is non-numeric", 400, w)
		return
	}

	user, err := cfg.DB.GetUser(id)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	if user == nil {
		w.WriteHeader(404)
		return
	}

	writeResponse(user, 200, w)
}

func (cfg *APIConfig) HandlerGetUsers(w http.ResponseWriter, _ *http.Request) {
	chirps, err := cfg.DB.GetUsers()
	if err != nil {
		w.WriteHeader(500)
	}
	writeResponse(chirps, 200, w)
}

func (cfg *APIConfig) HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
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
