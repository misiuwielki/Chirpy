package auth

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	uID := uuid.New()
	expiresInC := 1 * time.Hour
	creationSecret := "creation-secret"
	cases := []struct {
		expiresIn   time.Duration
		tokenSecret string
		wantErr     bool
	}{
		{
			expiresIn:   expiresInC,
			tokenSecret: creationSecret,
			wantErr:     false,
		},
		{
			expiresIn:   expiresInC,
			tokenSecret: "wrong-secret",
			wantErr:     true,
		},
		{
			expiresIn:   -expiresInC,
			tokenSecret: creationSecret,
			wantErr:     true,
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			token, err := MakeJWT(uID, creationSecret, c.expiresIn)
			if err != nil {
				t.Fatalf("MakeJWT unexpectedly failed: %v", err)
			}
			returnedId, err := ValidateJWT(token, c.tokenSecret)
			if (err != nil) != c.wantErr {
				t.Errorf("Validate JWT err = %v, wantErr = %v", err, c.wantErr)
			}
			if (returnedId == uID) != !c.wantErr {
				t.Errorf("UUID doesn't match - wanted = %v, got = %v", err, c.wantErr)
			}
		})
	}
}

func GetBearerTokenTest(t *testing.T) {
	cases := []struct {
		header  http.Header
		token   string
		wantErr bool
	}{
		{
			header: http.Header{
				"Authorization": []string{"Bearer 123456polki"},
			},
			token:   "123456polki",
			wantErr: false,
		},
		{
			header:  http.Header{},
			token:   "",
			wantErr: true,
		},
		{
			header: http.Header{
				"Authorization": []string{"Bear 123456polki"},
			},
			token:   "",
			wantErr: true,
		},
		{
			header: http.Header{
				"Authorization": []string{"Bearer     123456polki   "},
			},
			token:   "123456polki",
			wantErr: false,
		},
		{
			header: http.Header{
				"Authorization": []string{"123456polki"},
			},
			token:   "",
			wantErr: true,
		},
		{
			header: http.Header{
				"Authorization": []string{"Bearer     rareBear123   "},
			},
			token:   "rareBear123",
			wantErr: false,
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			tokenString, err := GetBearerToken(c.header)
			if (err != nil) != c.wantErr {
				t.Errorf("GetBearerToken err = %v, wantErr = %v", err, c.wantErr)
			}
			if tokenString != c.token {
				t.Errorf("Wrong token value, expected %v, got %v", c.token, tokenString)
			}
		})
	}
}
