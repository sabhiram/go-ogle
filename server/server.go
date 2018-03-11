package server

////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/sabhiram/go-ogle/hub"
	"github.com/sabhiram/go-ogle/server/socket"
)

////////////////////////////////////////////////////////////////////////////////

const (
	cEnableDebugProfiling = true
)

////////////////////////////////////////////////////////////////////////////////

var (
	wsUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

////////////////////////////////////////////////////////////////////////////////

// Server handles all websocket, HTTP API and file requests.
type Server struct {
	*http.Server
	hub *hub.Hub
}

// New returns an instance of Server.
func New(addr string, h *hub.Hub) (*Server, error) {
	s := &Server{
		Server: &http.Server{
			Addr: addr,
		},

		hub: h,
	}

	return s, s.setupRoutes()
}

////////////////////////////////////////////////////////////////////////////////

func (s *Server) wsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("wsHandler :: error :: %s\n", err.Error())
			return
		}

		sock := socket.New(c, s.hub)

		s.hub.RegisterSocket(sock)
		defer func() {
			s.hub.UnregisterSocket(sock)
		}()

		go sock.Read()
		sock.Write()
	}
}

func (s *Server) setupRoutes() error {
	mux := http.NewServeMux()
	mux.Handle("/ws", s.wsHandler())

	s.Handler = mux

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (s *Server) Start() {
	if err := s.ListenAndServe(); err != nil {
		fmt.Printf("error :: webserver died :: %s\n", err.Error())
	}
}

////////////////////////////////////////////////////////////////////////////////
