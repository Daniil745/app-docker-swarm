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
)

func main() {
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	log.Println("Сервер успешно запущен")

	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Успешное подключение к БД")

	// Redis для кэширования
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	redisClient, err := config.NewClient(ctx, config.Config{
		Addr:        "film_redis:6379",
		Password:    "",
		DB:          0,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal("Ошибка подключения к Redis:", err)
	}
	defer redisClient.Close()
	log.Println("Успешное подключение к Redis")

	filmStore := database.NewFilmStore(db.DB)
	filmsCache := cache.NewFilmsCache(redisClient)
	handler := handlers.NewHandlers(filmStore, filmsCache)

	mux := http.NewServeMux()

	mux.HandleFunc("/films", methodHandler(handler.GetAllFilms, "GET"))
	mux.HandleFunc("/films/create", methodHandler(handler.CreateTask, "POST"))

	mux.HandleFunc("/films/", filmIDHandler(handler))

	loggedMux := loggingMiddleware(mux)

	serverAddr := ":" + serverPort

	err = http.ListenAndServe(serverAddr, loggedMux)

	if err != nil {
		log.Fatal(err)
	}

}

func methodHandler(handlerFunc http.HandlerFunc, allowedMethod string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowedMethod {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		handlerFunc(w, r)
	}
}

func filmIDHandler(handler *handlers.HandlersTask) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetFilms(w, r)
		case http.MethodPut:
			handler.UpdateTask(w, r)
		case http.MethodDelete:
			handler.DeleteTask(w, r)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
