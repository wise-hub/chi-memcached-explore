package main

import (
	"cgi-memcached-explore/internal"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	fmt.Println(os.Getenv("TEST_ENV"))

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/api/login", internal.LoginEndpoint)
	router.With(internal.AuthMiddleware).Get("/api/resource", internal.ResourceEndpoint)

	fmt.Println("Starting server on :8080")
	http.ListenAndServe(":8080", router)

}
