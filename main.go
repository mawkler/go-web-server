package main

import (
	"encoding/json"
	"fmt"
	"github.com/mawkler/go-web-server/database"
	"log"
	"net/http"
	"strings"
)

type apiConfig struct {
	DB             *database.DB
	fileserverHits int
}

type errResponse struct {
	Error string `json:"error"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintf(w, `
		  <html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		  </html>
		`, cfg.fileserverHits,
	)
}

func (cfg *apiConfig) handlerReset(_ http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits = 0
}

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

func main() {
	db := database.New("database/database.json")
	mux := http.NewServeMux()

	cfg := apiConfig{fileserverHits: 0, DB: db}
	fileServer := http.FileServer(http.Dir("."))
	appHandler := http.StripPrefix("/app", fileServer)
	mux.Handle("/app/*", cfg.middlewareMetricsInc(appHandler))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("/api/reset", cfg.handlerReset)
	mux.HandleFunc("/api/validate_chirp", cfg.handlerValidateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{id}", cfg.handlerGetChirp)
	mux.HandleFunc("POST /api/chirps", cfg.handlerCreateChirp)

	corsMux := middlewareCors(mux)
	port := "8080"
	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: corsMux,
	}

	fmt.Printf("Listening to port %s", port)
	server.ListenAndServe()
}
