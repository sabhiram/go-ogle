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
	jb     types.JSONBroadcaster
}

func New(c *websocket.Conn, jb types.JSONBroadcaster) *Socket {
	return &Socket{
		conn:   c,
		sendCh: make(chan *types.SocketMessage),
		jb:     jb,
	}
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
			sm := types.SocketMessage{}
			sm.Unmarshal(msg)
			s.jb.BroadcastJSON(sm.Type, sm.Data)
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
