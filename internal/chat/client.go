package chat

import (
	"context"
	"log"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type message struct {
	Username string `json:"username"`
	Msg      string `json:"msg"`
}

type client struct {
	username    string
	conn        *websocket.Conn
	currentRoom *room
	receive     chan *message
}

func (c *client) readPump() {
	defer func() {
		c.currentRoom.leave <- c
		c.conn.Close(websocket.StatusInternalError, "read pump internal error")
	}()

	ctx := context.Background()

	for {
		var message message
		err := wsjson.Read(ctx, c.conn, &message)
		if err != nil {
			log.Printf("error reading message from pump: %v", err)
			return
		}
		message.Username = c.username

		c.currentRoom.broadcast <- &message
	}
}

func (c *client) writePump() {
	defer c.conn.Close(websocket.StatusInternalError, "write pump internal error")

	for m := range c.receive {
		err := wsjson.Write(context.Background(), c.conn, m)
		if err != nil {
			log.Printf("error writing message to pump: %v", err)
			return
		}
	}

	log.Println("client receive channel closed")
}
