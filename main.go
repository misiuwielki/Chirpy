package main

import (
	"encoding/json"
	"fmt"
	"log"
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

func (cfg *apiConfig) middlewareMetricsRead(w http.ResponseWriter, r *http.Request) {
	hits := cfg.fileserverHits.Load()
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	rsp := fmt.Sprintf("<html>\n  <body>\n    <h1>Welcome, Chirpy Admin</h1>\n    <p>Chirpy has been visited %d times!</p>\n  </body>\n</html>", hits)
	w.Write([]byte(rsp))
}

func (cfg *apiConfig) middlewareMetricsReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Swap(0)
	w.WriteHeader(http.StatusOK)
}

func handlerReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handlerPost(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	prm := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&prm)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	if len(prm.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}
	type validCh struct {
		Valid bool `json:"valid"`
	}
	vS := validCh{Valid: true}
	respondWithJSON(w, http.StatusOK, vS)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorS struct {
		Error string `json:"error"`
	}
	eS := errorS{msg}
	respondWithJSON(w, code, eS)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error while marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func main() {
	cfg := apiConfig{}
	serveMux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", cfg.middlewareMetricsInc(handler))
	serveMux.HandleFunc("GET /api/healthz", handlerReady)
	serveMux.HandleFunc("GET /admin/metrics", cfg.middlewareMetricsRead)
	serveMux.HandleFunc("POST /admin/reset", cfg.middlewareMetricsReset)
	serveMux.HandleFunc("POST /api/validate_chirp", handlerPost)
	server := http.Server{Addr: ":8080", Handler: serveMux}
	server.ListenAndServe()
}
