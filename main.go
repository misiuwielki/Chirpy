package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/misiuwielki/Chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error while connecting to database %v", err)
	}
	cfg := apiConfig{
		db:       database.New(db),
		platform: os.Getenv("PLATFORM"),
		secret:   os.Getenv("SECRET"),
	}
	serveMux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serveMux.Handle("/app/", cfg.middlewareMetricsInc(handler))
	serveMux.HandleFunc("GET /api/healthz", handlerReady)
	serveMux.HandleFunc("GET /admin/metrics", cfg.middlewareMetricsRead)
	serveMux.HandleFunc("POST /admin/reset", cfg.middlewareMetricsReset)
	serveMux.HandleFunc("POST /api/chirps", cfg.handlerPostChirp)
	serveMux.HandleFunc("GET /api/chirps", cfg.handlerGetAllChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerGetSingleChirp)
	serveMux.HandleFunc("POST /api/users", cfg.handlerNewUser)
	serveMux.HandleFunc("POST /api/login", cfg.handlerLogin)
	serveMux.HandleFunc("POST /api/refresh", cfg.handlerRefresh)
	serveMux.HandleFunc("POST /api/revoke", cfg.handlerRevokeRefreshToken)
	server := http.Server{Addr: ":8080", Handler: serveMux}
	server.ListenAndServe()
}
