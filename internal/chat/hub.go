package chat

import (
	"errors"
	"log"

	"github.com/google/uuid"
	"nhooyr.io/websocket"
)

type hub struct {
	rooms    map[uuid.UUID]*room
	register chan *client
	addRoom  chan *room
	bot      *bot
}

func NewHub(bot *bot) *hub {
	return &hub{
		rooms:    make(map[uuid.UUID]*room),
		register: make(chan *client),
		addRoom:  make(chan *room),
		bot:      bot,
	}
}

func (h *hub) Run() {
	log.Println("hub running...")
	for {
		select {
		case c := <-h.register:
			h.rooms[c.currentRoom.ID].join <- c
		case r := <-h.addRoom:
			h.rooms[r.ID] = r
		}

	}
}

func (h *hub) createRoom(title string) (*room, error) {

	room, err := newRoom(title, h.bot)
	if err != nil {
		return nil, err
	}

	h.bot.registerRoomId <- room.ID

	go room.run()

	h.addRoom <- room

	return room, nil
}

func (h *hub) getRoom(roomId uuid.UUID) (*room, error) {
	room, ok := h.rooms[roomId]
	if !ok {
		return nil, errors.New("room doesn't exist")
	}

	return room, nil
}

func (h *hub) join(conn *websocket.Conn, username string, roomId uuid.UUID) error {

	room, err := h.getRoom(roomId)
	if err != nil {
		return err
	}

	newClient := &client{
		username:    username,
		conn:        conn,
		currentRoom: room,
		receive:     make(chan *message),
	}

	h.register <- newClient

	go newClient.readPump()
	go newClient.writePump()

	return nil
}
