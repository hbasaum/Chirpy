package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func CheckHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) HomePageViewCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	hits := fmt.Sprintf(`
	<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
	`, cfg.fileserverHits.Load())

	w.Write([]byte(hits))
}

func (cfg *apiConfig) ResetViewCount(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)

	w.Write([]byte("Hits: 0"))
}

// chirps
func HandleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type reqVals struct {
		Body string `json:"Body"`
	}

	type validResVals struct {
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)

	req := reqVals{}
	err := decoder.Decode(&req)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
	}

	if len(req.Body) > 140 {
		RespondWithError(w, http.StatusBadRequest, "Chirp is too long", err)
	}

	RespondWithJSON(w, http.StatusOK, validResVals{Valid: true})
}

func main() {
	const port = "8080"
	mux := http.NewServeMux()

	apiConfig := &apiConfig{}

	mux.Handle("/app/", http.StripPrefix("/app/", apiConfig.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	// mux.Handle("/assets", http.FileServer(http.Dir("/assets/logo.png")))

	mux.HandleFunc("GET /api/healthz", CheckHealth)
	mux.HandleFunc("GET /admin/metrics", apiConfig.HomePageViewCount)
	mux.HandleFunc("POST /admin/reset", apiConfig.ResetViewCount)
	mux.HandleFunc("POST /api/validate_chirp", HandleValidateChirp)

	srv := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	srv.ListenAndServe()
}
