package main

import (
	_ "embed"
	"log"
	"server/internal/db"
	"server/internal/user"
	"server/internal/ws"
	"server/router"

	_ "modernc.org/sqlite"
)

func main() {
	dbPool, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("Error creating database: %v", err)
	}

	hub := ws.NewHub()
	wsRepository := ws.NewRepository(dbPool)
	wsService := ws.NewService(wsRepository)
	wsHandler := ws.NewHandler(hub, wsService)

	userRepository := user.NewRepository(dbPool)
	userService := user.NewService(userRepository, hub)
	userHandler := user.NewHandler(userService)

	go hub.Run()

	router.StartRouter(userHandler, wsHandler)
}
