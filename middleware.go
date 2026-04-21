package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/misiuwielki/Chirpy/internal/auth"
)

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
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	cfg.fileserverHits.Swap(0)
	err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		log.Printf("Error on resetting users database: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) middlewareAuthenticate(w http.ResponseWriter, r *http.Request) (uID uuid.UUID) {
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error on header authorization: %v", err))
		return uuid.Nil
	}
	uID, err = auth.ValidateJWT(tokenString, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return uuid.Nil
	}
	return uID
}
