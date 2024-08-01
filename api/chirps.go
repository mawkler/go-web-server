package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func cleanMessage(msg string) string {
	profanities := []string{"kerfuffle", "sharbert", "fornax"}
	newWords := []string{}

	for _, w := range strings.Split(msg, " ") {
		newWord := w
		for _, p := range profanities {
			if strings.ToLower(w) == p {
				newWord = "****"
			}
		}
		newWords = append(newWords, newWord)
	}

	return strings.Join(newWords, " ")
}

func (cfg *APIConfig) HandlerValidateChirp(w http.ResponseWriter, r *http.Request) {
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

func getSubject(token *jwt.Token, w http.ResponseWriter) (int, error) {
	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		w.WriteHeader(500)
		return 0, fmt.Errorf("failed to get subject from token: %s", err)
	}

	userID, err := strconv.Atoi(userIDString)
	if err != nil {
		w.WriteHeader(403)
		return 0, fmt.Errorf("token subject is non-numeric")
	}

	return userID, nil
}

func (cfg *APIConfig) HandlerCreateChirp(w http.ResponseWriter, r *http.Request) {
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

func (cfg *APIConfig) HandlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
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

func (cfg *APIConfig) HandlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.DB.GetChirps()
	if err != nil {
		w.WriteHeader(500)
	}
	writeResponse(chirps, 200, w)
}

func (cfg *APIConfig) HandlerGetChirp(w http.ResponseWriter, r *http.Request) {
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
