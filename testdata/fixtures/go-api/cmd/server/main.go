package main

import (
	"log"
	"net/http"

	"example.com/go-api/internal/handler"
	"example.com/go-api/internal/service"
)

func main() {
	svc := service.NewUserService()
	h := handler.NewHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", h.GetUser)
	mux.HandleFunc("POST /users", h.CreateUser)

	log.Println("starting server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
