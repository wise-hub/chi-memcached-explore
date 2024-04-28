package main

import (
	"cgi-memcached-explore/internal"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/api/login", internal.LoginEndpoint)
	router.With(internal.AuthMiddleware).Get("/api/resource", internal.ResourceEndpoint)

	fmt.Println("Starting server on :8080")
	http.ListenAndServe(":8080", router)

}
