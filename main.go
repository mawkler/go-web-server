package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
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

func writeResponse[T any](response T, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/json")
	resp, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		w.Write(nil)
	}
	w.Write(resp)
}

func (cfg *apiConfig) handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Body string `json:"body"`
	}

	type errResponse struct {
		Error string `json:"error"`
	}

	type okResponse struct {
		Valid bool `json:"valid"`
	}

	req := request{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(400)
		writeResponse(errResponse{Error: "Invalid JSON body"}, w)
		return
	}

	if len(req.Body) > 140 {
		w.WriteHeader(400)
		writeResponse(errResponse{Error: "Chirp is too long"}, w)
		return
	}

	writeResponse(okResponse{Valid: true}, w)
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
	mux := http.NewServeMux()

	apiCfg := apiConfig{fileserverHits: 0}
	fileServer := http.FileServer(http.Dir("."))
	appHandler := http.StripPrefix("/app", fileServer)
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(appHandler))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("/api/reset", apiCfg.handlerReset)
	mux.HandleFunc("/api/validate_chirp", apiCfg.handlerValidateChirp)

	corsMux := middlewareCors(mux)
	port := "8080"
	server := http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: corsMux,
	}

	fmt.Printf("Listening to port %s", port)
	server.ListenAndServe()
}
