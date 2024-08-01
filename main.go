package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/mawkler/go-web-server/api"
	"github.com/mawkler/go-web-server/database"
)

type errResponse struct {
	Error string `json:"error"`
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	databasePath := "database/database.json"
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *debug {
		os.Remove(databasePath)
	}

	godotenv.Load()
	jwtSecret := os.Getenv("JWT_SECRET")

	db := database.New(databasePath)
	mux := http.NewServeMux()

	cfg := api.NewAPIConfig(jwtSecret, 0, db)
	fileServer := http.FileServer(http.Dir("."))
	appHandler := http.StripPrefix("/app", fileServer)

	// File server
	mux.Handle("/app/*", cfg.MiddlewareMetricsInc(appHandler))

	// Health and metrics
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /admin/metrics", cfg.HandlerMetrics)
	mux.HandleFunc("GET /api/reset", cfg.HandlerReset)

	// Authentication
	mux.HandleFunc("POST /api/login", cfg.HandlerLogin)
	mux.Handle("POST /api/refresh", cfg.MiddlewareAuthorization(http.HandlerFunc(cfg.HandlerRefresh)))
	mux.Handle("POST /api/revoke", cfg.MiddlewareAuthorization(http.HandlerFunc(cfg.HandlerRevoke)))

	// Chirps
	mux.HandleFunc("/api/validate_chirp", cfg.HandlerValidateChirp)
	mux.Handle("POST /api/chirps", cfg.MiddlewareAuthorization(http.HandlerFunc(cfg.HandlerCreateChirp)))
	mux.Handle("DELETE /api/chirps/{id}", cfg.MiddlewareAuthorization(http.HandlerFunc(cfg.HandlerDeleteChirp)))
	mux.HandleFunc("GET /api/chirps", cfg.HandlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{id}", cfg.HandlerGetChirp)

	// Users
	mux.HandleFunc("POST /api/users", cfg.HandlerCreateUser)
	mux.HandleFunc("GET /api/users", cfg.HandlerGetUsers)
	mux.HandleFunc("GET /api/users/{id}", cfg.HandlerGetUser)
	mux.Handle("PUT /api/users", cfg.MiddlewareAuthorization(http.HandlerFunc(cfg.HandlerUpdateUser)))

	// Webhooks
	mux.HandleFunc("POST /api/polka/webhooks", cfg.HandlerUpgraded)

	corsMux := middlewareCors(mux)
	port := "8080"
	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: corsMux,
	}

	fmt.Printf("Listening to port %s\n", port)
	server.ListenAndServe()
}
