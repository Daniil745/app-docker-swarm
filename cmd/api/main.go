package main

import (
	"REST_api_appl/internal/cache"
	"REST_api_appl/internal/database"
	"REST_api_appl/internal/handlers"
	"REST_api_appl/storage/config"
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	log.Println("Server is started")

	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Successfully connect to database")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	redisClient, err := config.NewClient(ctx, config.Config{
		Addr:        "film_redis:6379",
		Password:    "",
		DB:          0,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal("error connection to Redis:", err)
	}
	defer redisClient.Close()
	log.Println("Success connection to Redis")

	filmStore := database.NewFilmStore(db.DB)
	filmsCache := cache.NewFilmsCache(redisClient)
	handler := handlers.NewHandlers(filmStore, filmsCache)

	mux := http.NewServeMux()

	mux.Handle("GET /metrics", promhttp.Handler())

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /films", handler.GetAllFilms)
	mux.HandleFunc("POST /films/create", handler.CreateTask)

	mux.HandleFunc("GET /films/{id}", handler.GetFilms)
	mux.HandleFunc("PUT /films/{id}", handler.UpdateTask)
	mux.HandleFunc("DELETE /films/{id}", handler.DeleteTask)

	loggedMux := loggingMiddleware(mux)

	serverAddr := ":" + serverPort
	log.Printf("Server starting on %s", serverAddr)

	err = http.ListenAndServe(serverAddr, loggedMux)
	if err != nil {
		log.Fatal(err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
