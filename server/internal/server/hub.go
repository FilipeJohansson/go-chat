package server

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"net/http"
	"sort"

	"time"

	"server/internal/server/db"
	"server/internal/server/objects"
	"server/pkg/packets"

	_ "modernc.org/sqlite"
)

//go:embed db/config/schema.sql
var schemaGenSql string

type DbTx struct {
	Ctx     context.Context
	Queries *db.Queries
}

func (h *Hub) NewDbTx() *DbTx {
	return &DbTx{
		Ctx:     context.Background(),
		Queries: db.New(h.dbPool),
	}
}

// A structure for a state machine to process the client's messages
type ClientStateHandler interface {
	Name() string

	// Inject the client into the state handler
	SetClient(client ClientInterfacer)

	OnEnter()
	HandleMessage(senderId uint64, message packets.Msg)

	// Cleanup the state handle and perform any alst actions
	OnExit()
}

type ClientInterfacer interface {
	Id() uint64
	ProcessMessage(senderId uint64, message packets.Msg)

	// Sets the client's ID and anything else that needs to be initialized
	Initialize(id uint64)

	SetState(state ClientStateHandler)

	// Puts data from this client into the write pump
	SocketSend(message packets.Msg)

	// Puts data from another client into the write pump
	SocketSendAs(message packets.Msg, senderId uint64)

	// Foward message to another client for processing
	PassToPeer(message packets.Msg, peerId uint64)

	// Forward message to all other clients for processing
	Broadcast(message packets.Msg)

	// Pump data from the connected socket directly to the client
	ReadPump()

	// Pump data from the client directly to the connected socket
	WritePump()

	// A reference to the database transaction context for this client
	DbTx() *DbTx

	// Close the client's connections and cleanup
	Close(reason string)
}

type StoragedMessage struct {
	Timestamp time.Time
	Msg       *packets.Packet_Chat
	SenderId  uint64
}

// The hub is the central point of communication between all connected clients
type Hub struct {
	Clients *objects.SharedCollection[ClientInterfacer]

	// Last messages sent from clients, so it can be sent to new clients
	LastMessages *objects.SharedCollection[StoragedMessage]

	// Packets in this channel will be processed by all connected clients except the sender
	BroadcastChan chan *packets.Packet

	// Clients in this channel will be registered to the hub
	RegisterChan chan ClientInterfacer

	// Clients in this channel will be unregistered from the hub
	UnregisterChan chan ClientInterfacer

	dbPool *sql.DB
}

func NewHub() *Hub {
	dbPool, err := sql.Open("sqlite", "db.sqlite")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	return &Hub{
		Clients:        objects.NewSharedCollection[ClientInterfacer](),
		LastMessages:   objects.NewSharedCollection[StoragedMessage](),
		BroadcastChan:  make(chan *packets.Packet, 256),
		RegisterChan:   make(chan ClientInterfacer),
		UnregisterChan: make(chan ClientInterfacer),
		dbPool:         dbPool,
	}
}

func (h *Hub) Run() {
	log.Println("Initializing database")
	if _, err := h.dbPool.ExecContext(context.Background(), schemaGenSql); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	log.Println("Awaiting client registrations")
	for {
		select {
		case client := <-h.RegisterChan:
			h.RemoveOldMessages(h.LastMessages)
			client.Initialize(h.Clients.Add(client))
		case client := <-h.UnregisterChan:
			client.Broadcast(packets.NewUnregister(client.Id()))
			h.Clients.Remove(client.Id())
		case packet := <-h.BroadcastChan:
			h.Clients.ForEach(func(clientId uint64, client ClientInterfacer) {
				if clientId != packet.SenderId {
					client.ProcessMessage(packet.SenderId, packet.Msg)
				}
			})
		}
	}
}

func (h *Hub) Serve(
	getNewClient func(*Hub, http.ResponseWriter, *http.Request) (ClientInterfacer, error),
	writter http.ResponseWriter,
	request *http.Request,
) {
	log.Println("New client connected from", request.RemoteAddr)
	client, err := getNewClient(h, writter, request)
	if err != nil {
		log.Printf("Error obtaining client for new connection: %v\n", err)
		return
	}

	h.RegisterChan <- client

	go client.WritePump()
	go client.ReadPump()
}

func (h *Hub) OrderLastMessages(lastMessages *objects.SharedCollection[StoragedMessage]) []StoragedMessage {
	messages := make([]StoragedMessage, 0, lastMessages.Len())
	lastMessages.ForEach(func(id uint64, sm StoragedMessage) {
		messages = append(messages, sm)
	})

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})

	return messages
}

func (h *Hub) RemoveOldMessages(lastMessages *objects.SharedCollection[StoragedMessage]) {
	lastMessages.ForEach(func(id uint64, sm StoragedMessage) {
		if time.Since(sm.Timestamp) >= 5*time.Minute {
			lastMessages.Remove(id)
		}
	})
}
