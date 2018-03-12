package socket

////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"

	"github.com/sabhiram/go-ogle/types"
)

////////////////////////////////////////////////////////////////////////////////

type Socket struct {
	conn   *websocket.Conn
	sendCh chan *types.SocketMessage
	ab     types.AppBroadcaster
}

func New(c *websocket.Conn, ab types.AppBroadcaster) *Socket {
	return &Socket{
		conn:   c,
		sendCh: make(chan *types.SocketMessage),
		ab:     ab,
	}
}

////////////////////////////////////////////////////////////////////////////////

// Returns true if the command passed in needs to be broadcasted.
func (s *Socket) HandleAppSpecificCommands(sm *types.SocketMessage) bool {
	fmt.Printf("APP SPECIFIC CMD HANDLER: %#v\n", sm)
	switch sm.Type {
	case "register_extension":
		err := s.ab.RegisterExtensionSocket(s.ID())
		if err != nil {
			fmt.Printf("ERROR: %s\n", err.Error())
		}
		return false
	}
	return true
}

func (s *Socket) ID() string {
	return fmt.Sprintf("%p", s)
}

////////////////////////////////////////////////////////////////////////////////

func (s *Socket) Read() {
	defer s.conn.Close()
	s.conn.SetReadLimit(1024)
	s.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		mt, msg, err := s.conn.ReadMessage()
		if err != nil {
			if _, ok := err.(*websocket.CloseError); !ok {
				fmt.Printf("wsHandler :: unable to read ws :: %s\n", err.Error())
			} else {
				fmt.Printf("wsHandler :: connection closed\n")
			}
			break
		}

		switch mt {
		case websocket.TextMessage:
			sm := &types.SocketMessage{}
			sm.Unmarshal(msg)
			if s.HandleAppSpecificCommands(sm) {
				s.ab.BroadcastJSON(sm.Type, sm.Data)
			}
		default:
			fmt.Printf("wsHandler :: unknown message type :: %d\n", mt)
		}

	}
}

func (s *Socket) Write() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		s.conn.Close()
	}()

	for {
		select {
		case sm, ok := <-s.sendCh:
			// s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			// Handle case when the hub / app closes a socket.
			if !ok {
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := s.conn.WriteJSON(sm); err != nil {
				return
			}
		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func (s *Socket) Send(sm *types.SocketMessage) {
	s.sendCh <- sm
}

func (s *Socket) Close() {
	close(s.sendCh)
}

////////////////////////////////////////////////////////////////////////////////
