package states

import (
	"context"
	"fmt"
	"log"
	"server/internal/server"
	"server/internal/server/db"
	"server/pkg/packets"
)

type Connected struct {
	client  server.ClientInterfacer
	logger  *log.Logger
	queries *db.Queries
	dbCtx   context.Context
}

func (c *Connected) Name() string {
	return "Connected"
}

func (c *Connected) SetClient(client server.ClientInterfacer) {
	c.client = client
	logginPrefix := fmt.Sprintf("Client %d [%s]: ", client.Id(), c.Name())
	c.logger = log.New(log.Writer(), logginPrefix, log.LstdFlags)
	c.queries = client.DbTx().Queries
	c.dbCtx = client.DbTx().Ctx
}

func (c *Connected) OnEnter() {
	c.client.SocketSend(packets.NewId(c.client.Id()))
}

func (c *Connected) HandleMessage(senderId uint64, message packets.Msg) {
	if senderId == c.client.Id() {
		// This message was sent by our own client, so broadcast it to everyone else
		c.client.Broadcast(message)
	} else {
		// Another client interfacer passed this onto us, or it was broadcast from the hub,
		// so forward it to our own client
		c.client.SocketSendAs(message, senderId)
	}
}

func (c *Connected) OnExit() {

}
