package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/misiuwielki/Chirpy/internal/auth"
	"github.com/misiuwielki/Chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	prm := parameters{}
	err := decodeJson(r, &prm)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	if prm.Body == "" {
		respondWithError(w, http.StatusBadRequest, "Chirp text is required")
		return
	}
	if len(prm.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}
	prm.Body = profanityCheck(prm.Body)
	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error on header authorization: %v", err))
		return
	}
	uID, err := auth.ValidateJWT(tokenString, cfg.secret)
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   prm.Body,
		UserID: uID,
	})
	if err != nil {
		log.Printf("Error while adding chirp to database: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
	}
	chirpS := sqlToStructChirp(chirp)
	respondWithJSON(w, http.StatusCreated, chirpS)
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		log.Printf("Error while retrieving all chirps: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps")
	}
	jsonSlice := []Chirp{}
	for _, chirp := range chirps {
		chirpS := sqlToStructChirp(chirp)
		jsonSlice = append(jsonSlice, chirpS)
	}
	respondWithJSON(w, http.StatusOK, jsonSlice)

}

func (cfg *apiConfig) handlerGetSingleChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDs := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDs)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID")
		return
	}
	chirp, err := cfg.db.GetSingleChirp(r.Context(), chirpID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Chirp not found")
			return
		}
		log.Printf("Error while retrieving a chirp: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't process finding chirp")
		return
	}
	chirpS := sqlToStructChirp(chirp)
	respondWithJSON(w, http.StatusOK, chirpS)
}

func profanityCheck(chirp string) string {
	words := strings.Split(chirp, " ")
	censored := []string{}
	profane := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	for _, word := range words {
		_, ok := profane[strings.ToLower(word)]
		if ok {
			word = "****"
		}
		censored = append(censored, word)
	}
	correctChirp := strings.Join(censored, " ")
	return correctChirp
}
