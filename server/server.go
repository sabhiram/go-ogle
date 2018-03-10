package server

////////////////////////////////////////////////////////////////////////////////

import (
	"fmt"
	"net/http"
	"net/http/pprof"

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
			fmt.Printf("CheckOrigin: %s\n", r.URL.String())
			return true
		},
	}
)

////////////////////////////////////////////////////////////////////////////////

// Server handles all websocket, HTTP API and file requests.
type Server struct {
	*http.Server

	hub *hub.Hub // websocket hub
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

func (s *Server) todoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("TODO handler hit!\n")
		w.Write([]byte("TODO"))
	}
}

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

	// Debugging
	if cEnableDebugProfiling {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	}

	mux.Handle("/ws", s.wsHandler())

	s.Handler = mux
	return nil
}

////////////////////////////////////////////////////////////////////////////////

func (s *Server) Start() {
	fmt.Printf("Kicking off webserver at: %s\n", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		fmt.Printf("error :: webserver died :: %s\n", err.Error())
	}
}

////////////////////////////////////////////////////////////////////////////////
