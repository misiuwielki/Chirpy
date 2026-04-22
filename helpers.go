package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/misiuwielki/Chirpy/internal/database"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type errorS struct {
		Error string `json:"error"`
	}
	eS := errorS{msg}
	respondWithJSON(w, code, eS)
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
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

func decodeJson(r *http.Request, dest any) error {
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(dest)
}

func sqlToStructChirp(chirp database.Chirp) Chirp {
	return Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
}

func sqlToStructUser(user database.User) User {
	return User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}
}
