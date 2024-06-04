package chat

import (
	"financial-chat-api/util/auth"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/tomiok/webh"
	"nhooyr.io/websocket"
)

type handler struct {
	hub *hub
}

func NewHandler(hub *hub) *handler {
	return &handler{
		hub: hub,
	}
}

func (h *handler) HandleCreateRoom(w http.ResponseWriter, r *http.Request) error {
	var req struct {
		Title string `json:"title"`
	}
	webh.DJson(r.Body, &req)

	room, err := h.hub.createRoom(req.Title)
	if err != nil {
		return webh.ErrHTTP{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	err = webh.EJson(w, room)
	if err != nil {
		return webh.ErrHTTP{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	return nil
}

func (h *handler) HandleJoinRoom(w http.ResponseWriter, r *http.Request) error {
	roomId := r.URL.Query().Get("roomId")
	if roomId == "" {
		return webh.ErrHTTP{Code: http.StatusBadRequest, Message: "roomId required"}
	}
	roomUuid, err := uuid.Parse(roomId)
	if err != nil {
		return webh.ErrHTTP{Code: http.StatusBadRequest, Message: "roomId must be uuid"}
	}

	_, err = h.hub.getRoom(roomUuid)
	if err != nil {
		return webh.ErrHTTP{Code: http.StatusBadRequest, Message: "room doesn't exist"}
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, //not recomended
		Subprotocols:       []string{"chat"}})
	if err != nil {
		return webh.ErrHTTP{Code: http.StatusInternalServerError, Message: "cant create websocket"}
	}

	authPayload := r.Context().Value(auth.AuthorizationPayloadCtxKey).(*auth.Payload)

	err = h.hub.join(conn, authPayload.Username, roomUuid)
	if err != nil {
		log.Println(err)
		conn.Close(websocket.StatusInternalError, err.Error())
	}

	return nil
}
