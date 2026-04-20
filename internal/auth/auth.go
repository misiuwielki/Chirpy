package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}
	return hashedPassword, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	return match, err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	t := time.Now()
	userIDString := userID.String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(t),
		ExpiresAt: jwt.NewNumericDate(t.Add(expiresIn)),
		Subject:   userIDString,
	})
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenSecret), nil
		})
	if err != nil {
		return uuid.Nil, err
	}
	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}
	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, err
	}
	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	tokenHeader := headers.Get("Authorization")
	if tokenHeader == "" {
		return "", fmt.Errorf("Authorization header not found")
	}
	validHeader := strings.HasPrefix(tokenHeader, "Bearer")
	if !validHeader {
		return "", fmt.Errorf("Invalid authorization header")
	}
	tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return "", fmt.Errorf("Token not found")
	}

	return tokenString, nil
}
