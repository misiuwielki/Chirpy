package main

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/misiuwielki/Chirpy/internal/auth"
	"github.com/misiuwielki/Chirpy/internal/database"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func handlerReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerNewUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	prm := parameters{}
	err := decodeJson(r, &prm)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	hashed_password, err := auth.HashPassword(prm.Password)
	if err != nil {
		log.Printf("Error while hashing password: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't create account")
		return
	}
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          prm.Email,
		HashedPassword: hashed_password,
	})
	userS := sqlToStructUser(user)
	respondWithJSON(w, http.StatusCreated, userS)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	prm := parameters{}
	err := decodeJson(r, &prm)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	if prm.Email == "" || prm.Password == "" {
		respondWithError(w, http.StatusBadRequest, "email and password are required")
		return
	}
	user, err := cfg.db.GetUser(r.Context(), prm.Email)
	if err != nil {
		log.Printf("Error on getting user from db: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	passwordCheck, err := auth.CheckPasswordHash(prm.Password, user.HashedPassword)
	if err != nil {
		log.Printf("Error on comparing password: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	if !passwordCheck {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	expirationTime := time.Duration(prm.ExpiresInSeconds) * time.Second
	if prm.ExpiresInSeconds > 3600 || prm.ExpiresInSeconds == 0 {
		expirationTime = 3600 * time.Second
	}
	userS := sqlToStructUser(user)
	token, err := auth.MakeJWT(userS.ID, cfg.secret, expirationTime)
	if err != nil {
		log.Printf("Error on comparing password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate token")
		return
	}
	userS.Token = token
	respondWithJSON(w, 200, userS)

}
