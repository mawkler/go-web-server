package main

import (
	"encoding/json"
	"net/http"
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

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Body string `json:"body"`
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResponse(errResponse{Error: "Invalid JSON body"}, 400, w)
		return
	}

	chirp, err := cfg.DB.CreateChirp(req.Body)
	if err != nil {
		w.WriteHeader(500)
	}
	writeResponse(chirp, 201, w)
}
