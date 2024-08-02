package api

import "github.com/mawkler/go-web-server/database"

type errResponse struct {
	Error string `json:"error"`
}

type APIConfig struct {
	DB             *database.DB
	jwtSecret      string
	polkaAPIKey    string
	fileserverHits int
}

func NewAPIConfig(database *database.DB, jwtSecret, polkaAPIKey string, fileserverHits int) APIConfig {
	return APIConfig{DB: database, polkaAPIKey: polkaAPIKey, jwtSecret: jwtSecret, fileserverHits: fileserverHits}
}
