package main

import (
	"database/sql"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
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
	uID := cfg.middlewareAuthenticate(w, r)
	if uID == uuid.Nil {
		return
	}
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
	authorID := r.URL.Query().Get("author_id")
	sortType := r.URL.Query().Get("sort")
	chirps := []database.Chirp{}
	var err error
	if authorID == "" {
		chirps, err = cfg.db.GetAllChirps(r.Context())
		if err != nil {
			log.Printf("Error while retrieving all chirps: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps")
		}
	} else {
		uID, err := uuid.Parse(authorID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid user ID")
			return
		}
		chirps, err = cfg.db.GetChirpsByAuthor(r.Context(), uID)
	}
	jsonSlice := []Chirp{}
	for _, chirp := range chirps {
		chirpS := sqlToStructChirp(chirp)
		jsonSlice = append(jsonSlice, chirpS)
	}
	if sortType == "desc" {
		sort.Slice(jsonSlice, func(i, j int) bool { return jsonSlice[i].CreatedAt.After(jsonSlice[j].CreatedAt) })
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	uID := cfg.middlewareAuthenticate(w, r)
	if uID == uuid.Nil {
		return
	}
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
	if chirp.UserID != uID {
		respondWithError(w, http.StatusForbidden, "Cannot delete a chirp of other user")
		return
	}
	err = cfg.db.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID:     chirpID,
		UserID: uID,
	})
	if err != nil {
		log.Printf("Error while deleting a chirp: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't process deleting chirp")
		return
	}
	w.WriteHeader(204)
}

// helper functions

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
