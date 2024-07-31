package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Body string `json:"body"`
	}

	type okResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResponse(errResponse{Error: "Invalid JSON body"}, 400, w)
	} else if len(req.Body) > 140 {
		writeResponse(errResponse{Error: "Chirp is too long"}, 400, w)
	} else {
		writeResponse(okResponse{CleanedBody: cleanMessage(req.Body)}, 200, w)
	}
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.DB.GetChirps()
	if err != nil {
		w.WriteHeader(500)
	}
	writeResponse(chirps, 200, w)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeResponse("Query parameter `id` is non-numeric", 400, w)
		return
	}

	chirp, err := cfg.DB.GetChirp(id)
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

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Body string `json:"body"`
	}

	token, ok := r.Context().Value(authorizedJWTKey).(*jwt.Token)
	if !ok || token == nil {
		log.Printf("context does not contain authorized access token")
		w.WriteHeader(500)
		return
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResponse(errResponse{Error: "Invalid JSON body"}, 400, w)
		return
	}

	userID, err := getSubject(token, w)
	if err != nil {
		log.Print(err)
		return
	}

	chirp, err := cfg.DB.CreateChirp(req.Body, userID)
	if err != nil {
		w.WriteHeader(500)
	}
	writeResponse(chirp, 201, w)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeResponse("Query parameter `id` is non-numeric", 400, w)
		return
	}

	chirp, err := cfg.DB.GetChirp(chirpID)
	if err != nil {
		log.Printf("tried to delete chirp %d, but something went wrong when retrieving it: %s", chirpID, err)
		w.WriteHeader(500)
		return
	}

	token, ok := r.Context().Value(authorizedJWTKey).(*jwt.Token)
	if !ok || token == nil {
		log.Printf("context does not contain authorized access token")
		w.WriteHeader(500)
		return
	}

	userID, err := getSubject(token, w)
	if err != nil {
		log.Print(err)
		return
	}

	if chirp.AuthorID != userID {
		w.WriteHeader(403)
		return
	}

	if err := cfg.DB.DeleteChirp(chirpID); err != nil {
		log.Printf("could not delete chirp %d: %s", chirpID, err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(204)
}
