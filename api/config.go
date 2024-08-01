package api

import "github.com/mawkler/go-web-server/database"

type errResponse struct {
	Error string `json:"error"`
}

type APIConfig struct {
	DB             *database.DB
	jwtSecret      string
	fileserverHits int
}

func NewAPIConfig(jwtSecret string, fileserverHits int, database *database.DB) APIConfig {
	return APIConfig{DB: database, jwtSecret: jwtSecret, fileserverHits: fileserverHits}
}
