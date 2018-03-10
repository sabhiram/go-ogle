package hub

////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"

	"github.com/sabhiram/go-ogle/server/socket"
	"github.com/sabhiram/go-ogle/types"
)

////////////////////////////////////////////////////////////////////////////////

// Hub represents a websocket hub.
type Hub struct {
	sockets map[*socket.Socket]struct{}

	broadcastCh  chan []byte
	registerCh   chan *socket.Socket
	unregisterCh chan *socket.Socket
}

// New returns a new websocket hub.
func New() (*Hub, error) {
	return &Hub{
		sockets: map[*socket.Socket]struct{}{},

		broadcastCh:  make(chan []byte),
		registerCh:   make(chan *socket.Socket),
		unregisterCh: make(chan *socket.Socket),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// RegisterSocket registers the socket `s` to the hub.
func (h *Hub) RegisterSocket(s *socket.Socket) {
	fmt.Printf("Registering socket\n")
	h.registerCh <- s
}

// UnregisterSocket unregisters the socket `s` from the hub.
func (h *Hub) UnregisterSocket(s *socket.Socket) {
	h.unregisterCh <- s
}

// Send a message to the specified socket.
func (h *Hub) Send(s *socket.Socket, msg []byte) {
	s.Send(msg)
}

func (h *Hub) SendJSON(s *socket.Socket, t string, d interface{}) error {
	sm := types.NewSocketMessage(t, d)
	bs, err := sm.Marshal()
	if err != nil {
		return err
	}
	h.Send(s, bs)
	return nil
}

// Broadcast a message to all subscribed sockets.
func (h *Hub) Broadcast(msg []byte) {
	fmt.Printf("Broadcast message: %s\n", msg)
	h.broadcastCh <- msg
}

// BroadcastJSON sends a packet of the format:
// {"Type": "`type`", "Data": interface{}}.
func (h *Hub) BroadcastJSON(t string, d interface{}) error {
	sm := types.NewSocketMessage(t, d)
	bs, err := sm.Marshal()
	if err != nil {
		return err
	}
	h.Broadcast(bs)
	return nil
}

// Run kicks off an endless loop that processes the websocket hub.  This is
// responsible for registering and unregistering sockets as we ll as
// broadcasting messages for all connected clients.
func (h *Hub) Run() {
	for {
		select {

		// Register.
		case socket := <-h.registerCh:
			h.sockets[socket] = struct{}{}

		// Unregister.
		case socket := <-h.unregisterCh:
			if _, ok := h.sockets[socket]; ok {
				delete(h.sockets, socket)
				socket.Close()
			}

		// Broadcast.
		case msg := <-h.broadcastCh:
			for socket := range h.sockets {
				socket.Send(msg)
			}

		}
	}
}

////////////////////////////////////////////////////////////////////////////////
