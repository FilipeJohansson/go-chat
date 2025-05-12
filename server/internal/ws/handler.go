package ws

import (
	"log"
	"net/http"
	"server/internal/client"
)

type Handler struct {
	hub     *Hub
	service Service
}

func NewHandler(h *Hub, s Service) *Handler {
	return &Handler{
		hub:     h,
		service: s,
	}
}

func (h *Handler) Serve(
	getNewClient func(*Hub, Service, http.ResponseWriter, *http.Request) (client.ClientInterfacer, error),
	writer http.ResponseWriter,
	request *http.Request,
) {
	log.Println("New client connected from", request.RemoteAddr)
	client, err := getNewClient(h.hub, h.service, writer, request)
	if err != nil {
		log.Printf("Error obtaining client for new connection: %v\n", err)
		return
	}

	if client == nil {
		return
	}

	h.hub.RegisterChan <- client

	go client.WritePump()
	go client.ReadPump()
}
