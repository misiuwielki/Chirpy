package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
	"github.com/misiuwielki/Chirpy/internal/auth"
)

type Token struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	rToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error on header authorization: %v", err))
		return
	}
	uID, err := cfg.db.GetUserFromRefreshToken(r.Context(), rToken)
	if err != nil {
		log.Printf("Error on verifying refresh token: %v", err)
		respondWithError(w, http.StatusUnauthorized, "Not authorized")
		return
	}
	aToken, err := auth.MakeJWT(uID, cfg.secret, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate token")
		return
	}
	Token := Token{Token: aToken}
	respondWithJSON(w, http.StatusOK, Token)
}

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	rToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error on header authorization: %v", err))
		return
	}
	err = cfg.db.RevokeRefreshToken(r.Context(), rToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database error")
		return
	}
	w.WriteHeader(204)
}
