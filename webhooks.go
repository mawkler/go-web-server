package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerUpgraded(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Event string `json:"event"`
		Data  struct {
			UserID int `json:"user_id"`
		} `json:"data"`
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeResponse(errResponse{Error: "Invalid JSON body"}, 400, w)
		return
	}

	if req.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	if err := cfg.DB.UpgradeUser(req.Data.UserID); err != nil {
		log.Printf("failed to upgrade user: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(204)
}
