package router

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"server/internal/user"
	"server/internal/ws"

	"github.com/rs/cors"
)

var (
	port = flag.Int("port", 8080, "Port to listen on")
)

func StartRouter(userHandler *user.Handler, wsHandler *ws.Handler) {
	flag.Parse()

	mux := http.NewServeMux()

	handler := cors.New(cors.Options{
		AllowOriginFunc: checkOrigin,
		AllowedMethods:  []string{"GET", "POST"},
		AllowedHeaders:  []string{"Content-Type", "Authorization"},
	}).Handler(mux)

	mux.HandleFunc("/login", userHandler.Login)
	mux.HandleFunc("/register", userHandler.Register)
	mux.HandleFunc("/refresh", userHandler.RefreshToken)
	mux.HandleFunc("/logout", userHandler.Logout)
	mux.HandleFunc("/new-room", userHandler.CreateRoom)
	mux.HandleFunc("/rooms", userHandler.GetRooms)

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler.Serve(ws.NewWebSocketClient, w, r)
	})

	addr := fmt.Sprintf("0.0.0.0:%d", *port)

	log.Printf("Starting server on %s", addr)
	err := http.ListenAndServe(addr, handler)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func checkOrigin(origin string) bool {
	switch origin {
	case "http://localhost:5174":
		return true
	case "http://localhost:5175":
		return true
	default:
		return false
	}
}
