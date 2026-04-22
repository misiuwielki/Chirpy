package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/misiuwielki/Chirpy/internal/auth"
	"github.com/misiuwielki/Chirpy/internal/database"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
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
	hashedPassword, err := auth.HashPassword(prm.Password)
	if err != nil {
		log.Printf("Error while hashing password: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't create account")
		return
	}
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          prm.Email,
		HashedPassword: hashedPassword,
	})
	userS := sqlToStructUser(user)
	respondWithJSON(w, http.StatusCreated, userS)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	prm := parameters{}
	err := decodeJson(r, &prm)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	if prm.Email == "" || prm.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
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
	userS := sqlToStructUser(user)
	aToken, err := auth.MakeJWT(userS.ID, cfg.secret, 1*time.Hour)
	if err != nil {
		log.Printf("Error on comparing password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate token")
		return
	}
	rToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error on generating refreh token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate token")
		return
	}
	cfg.db.AddRefreshToken(r.Context(), database.AddRefreshTokenParams{
		Token:  rToken,
		UserID: userS.ID,
	})
	userS.Token = aToken
	userS.RefreshToken = rToken
	respondWithJSON(w, 200, userS)
}

func (cfg *apiConfig) handlerAlterUser(w http.ResponseWriter, r *http.Request) {
	uID := cfg.middlewareAuthenticate(w, r)
	if uID == uuid.Nil {
		return
	}
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	prm := parameters{}
	err := decodeJson(r, &prm)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	if prm.Email == "" || prm.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}
	hashedPassword, err := auth.HashPassword(prm.Password)
	if err != nil {
		log.Printf("Error while hashing password: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't update account")
		return
	}
	upUser, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             uID,
		Email:          prm.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("Error while updating user in database: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't update account")
		return
	}
	User := sqlToStructUser(upUser)
	respondWithJSON(w, http.StatusOK, User)
}

func (cfg *apiConfig) handlerPolkaUpgrade(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Error on authorisation: %v", err))
		return
	}
	if apiKey != cfg.polka {
		respondWithError(w, http.StatusUnauthorized, "Invalid API Key")
		return
	}
	type paramters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}
	prm := paramters{}
	err = decodeJson(r, &prm)
	if err != nil {
		log.Printf("Error decoding parameters from Polka: %s", err)
		respondWithError(w, http.StatusBadRequest, "Request invalid")
		return
	}
	if prm.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	_, err = cfg.db.UpgradeUser(r.Context(), prm.Data.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "User not found")
			return
		}
		log.Printf("Error while upgrading user: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Couldn't process upgrading user")
		return
	}
	w.WriteHeader(204)
}
