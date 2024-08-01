package api

import (
	"fmt"
	"net/http"
)

func (cfg *APIConfig) HandlerMetrics(w http.ResponseWriter, _ *http.Request) {
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

func (cfg *APIConfig) HandlerReset(_ http.ResponseWriter, _ *http.Request) {
	cfg.fileserverHits = 0
}
