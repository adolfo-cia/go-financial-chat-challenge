package chat

import (
	"log"

	"github.com/google/uuid"
)

type room struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	clients   map[*client]struct{}
	join      chan *client
	leave     chan *client
	broadcast chan *message
	bot       *bot
}

func newRoom(title string, bot *bot) (*room, error) {
	uuid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &room{
		ID:        uuid,
		Title:     title,
		clients:   make(map[*client]struct{}),
		join:      make(chan *client),
		leave:     make(chan *client),
		broadcast: make(chan *message),
		bot:       bot,
	}, nil
}

func (r *room) run() {
	log.Printf("room %s running...\n", r.ID)
	for {
		select {
		case client := <-r.join:
			r.clients[client] = struct{}{}
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.receive)
		case msg := <-r.broadcast:
			r.bot.sendCh <- &botMessage{RoomId: r.ID.String(), Msg: msg.Msg}
			for c := range r.clients {
				c.receive <- msg
			}
		case msg := <-r.bot.roomReceiveCh[r.ID]:
			for c := range r.clients {
				c.receive <- &message{Username: "BOT", Msg: msg.Msg}
			}
		}
	}
}
