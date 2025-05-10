package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"server/internal/server"
	"server/internal/server/clients"

	"github.com/rs/cors"
)

var (
	port = flag.Int("port", 8080, "Port to listen on")
)

func main() {
	flag.Parse()

	mux := http.NewServeMux()

	// Define the chat hub
	hub := server.NewHub()

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewHttpClient, w, r)
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewHttpClient, w, r)
	})
	mux.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewHttpClient, w, r)
	})
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewHttpClient, w, r)
	})
	mux.HandleFunc("/rooms", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewHttpClient, w, r)
	})
	mux.HandleFunc("/new-room", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewHttpClient, w, r)
	})
	// Define the handler for WebSocket connections
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewWebSocketClient, w, r)
	})

	handler := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // Allows all origins
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "Authorization"},
	}).Handler(mux)

	go hub.Run()
	addr := fmt.Sprintf("0.0.0.0:%d", *port)

	log.Printf("Starting server on %s", addr)
	err := http.ListenAndServe(addr, handler)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
