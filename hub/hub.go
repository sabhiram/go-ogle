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
	sockets  map[*socket.Socket]struct{}
	extSock  *socket.Socket
	extQueue []*types.SocketMessage

	broadcastCh  chan *types.SocketMessage
	registerCh   chan *socket.Socket
	unregisterCh chan *socket.Socket
}

// New returns a new websocket hub.
func New() (*Hub, error) {
	return &Hub{
		sockets:  map[*socket.Socket]struct{}{},
		extSock:  nil,
		extQueue: []*types.SocketMessage{},

		broadcastCh:  make(chan *types.SocketMessage),
		registerCh:   make(chan *socket.Socket),
		unregisterCh: make(chan *socket.Socket),
	}, nil
}

////////////////////////////////////////////////////////////////////////////////

// RegisterSocket registers the socket `s` to the hub.
func (h *Hub) RegisterSocket(s *socket.Socket) {
	h.registerCh <- s
}

// UnregisterSocket unregisters the socket `s` from the hub.
func (h *Hub) UnregisterSocket(s *socket.Socket) {
	// If the editor socket is unregistering, we need to clear the `extSock`
	// member so that we can build a queue on broadcast while we wait for the
	// browser to connect / re-connect.
	if s == h.extSock {
		h.extSock = nil
	}
	h.unregisterCh <- s
}

func (h *Hub) RegisterExtensionSocket(sid string) error {
	for s, _ := range h.sockets {
		if s.ID() == sid {
			h.extSock = s
			break
		}
	}

	if h.extSock == nil {
		return fmt.Errorf("ext socket not found")
	}
	for _, sm := range h.extQueue {
		h.extSock.Send(sm)
	}

	h.extQueue = []*types.SocketMessage{}
	return nil
}

// BroadcastJSON sends a packet of the format:
// {"Type": "`type`", "Data": interface{}}.
func (h *Hub) BroadcastJSON(t string, d interface{}) error {
	sm := types.NewSocketMessage(t, d)
	if h.extSock == nil {
		h.extQueue = append(h.extQueue, sm)
	}
	h.broadcastCh <- sm
	return nil
}

////////////////////////////////////////////////////////////////////////////////

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
		case sm := <-h.broadcastCh:
			for socket := range h.sockets {
				socket.Send(sm)
			}

		}
	}
}

////////////////////////////////////////////////////////////////////////////////
