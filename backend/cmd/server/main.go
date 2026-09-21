package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"url-shortener/internal/handler"
	"url-shortener/internal/middleware"
	"url-shortener/internal/service"
	"url-shortener/internal/store"

	"github.com/gorilla/mux"
)

func main() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres123")
	dbname := getEnv("DB_NAME", "urlshortener")
	serverPort := getEnv("SERVER_PORT", "8080")

	db, err := store.NewPostgresStore(host, port, user, password, dbname)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	urlService := service.NewURLService(db)
	urlHandler := handler.NewURLHandler(urlService)

	router := mux.NewRouter()
	router.Use(middleware.CORS)

	router.HandleFunc("/api/v1/shorten", urlHandler.ShortenURL).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/v1/stats/{code}", urlHandler.GetStats).Methods("GET", "OPTIONS")
	router.HandleFunc("/{code}", urlHandler.RedirectURL).Methods("GET")

	fmt.Printf("Server starting on port %s\n", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, router))
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
