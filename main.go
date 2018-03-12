package main

////////////////////////////////////////////////////////////////////////////////

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
	"github.com/pkg/term"

	"github.com/sabhiram/go-ogle/hub"
	"github.com/sabhiram/go-ogle/server"
	"github.com/sabhiram/go-ogle/types"
)

////////////////////////////////////////////////////////////////////////////////

var (
	cli = struct {
		isServer  bool
		homedir   string
		configDir string
		args      []string
	}{}
)

////////////////////////////////////////////////////////////////////////////////

func fatalOnError(err error) {
	if err != nil {
		fmt.Printf("Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}

////////////////////////////////////////////////////////////////////////////////

func getPID() (int, error) {
	pidFile := path.Join(cli.configDir, "pid")
	bs, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(bs))
}

func setPID(pid int) error {
	if _, err := os.Stat(cli.configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cli.configDir, 0777); err != nil {
			return err
		}
	}

	pidFile := path.Join(cli.configDir, "pid")
	return ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0777)
}

func serverRunning() (bool, error) {
	// Check for pid file, if it does not exist return false.
	_, err := getPID()
	if err != nil {
		return false, err
	}

	// Otherwise return true.
	return true, nil
}

func serverMain(args []string) {
	h, err := hub.New()
	fatalOnError(err)
	go h.Run()

	server, err := server.New(":18881", h)
	fatalOnError(err)

	server.Start()
}

////////////////////////////////////////////////////////////////////////////////

func spawnServerThread() error {
	cmd := exec.Command("go-ogle", "-server")
	if err := cmd.Start(); err != nil {
		return err
	}

	err := setPID(cmd.Process.Pid)
	if err != nil {
		return err
	}

	fmt.Printf("Allowing server to startup before attempting connection...\n")
	<-time.After(10 * time.Millisecond)
	return nil
}

func getKeyPress(t *term.Term) []byte {
	buf := make([]byte, 3)
	n, err := t.Read(buf)
	if err != nil {
		return nil
	}
	return buf[0:n]
}

func sendMessage(c *websocket.Conn, t string, d interface{}) error {
	sm := types.NewSocketMessage(t, d)
	bs, err := sm.Marshal()
	if err != nil {
		return err
	}

	return c.WriteMessage(websocket.TextMessage, bs)
}

func readWSMessages(c *websocket.Conn, ch chan<- *types.SocketMessage) {
	for {
		sm := types.SocketMessage{}
		fatalOnError(c.ReadJSON(&sm))
		if sm.Type != "" {
			ch <- &sm
		}
	}
}

func readKeysFromStdin(t *term.Term, ch chan<- []byte) {
	for {
		ch <- getKeyPress(t)
	}
}

func childMain(args []string) {
	// Check to see if the server is running
	running, _ := serverRunning()
	if !running {
		err := spawnServerThread()
		fatalOnError(err)
	}

	u := url.URL{Scheme: "ws", Host: "localhost:18881", Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		err = spawnServerThread()
		fatalOnError(err)

		// Try again now that we kicked off a server thread.
		c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	}
	fatalOnError(err)
	defer c.Close()

	q := fmt.Sprintf("https://www.google.com/search?q=%s", strings.Join(args, "+"))
	sendMessage(c, "open_new_tab_with_url", q)

	defer func(c *websocket.Conn) {
		err = sendMessage(c, "clear_selected", "")
		fatalOnError(err)

		err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		fatalOnError(err)
	}(c)

	// Open a terminal so we can poll keys from stdin.
	t, _ := term.Open("/dev/tty")
	term.RawMode(t)
	defer func() {
		t.Restore()
	}()

	keyCh := make(chan []byte)
	wsCmdCh := make(chan *types.SocketMessage)

	go readKeysFromStdin(t, keyCh)
	go readWSMessages(c, wsCmdCh)

	for {
		select {
		case key := <-keyCh:
			switch {
			case bytes.Equal(key, []byte{3}) || bytes.Equal(key, []byte{27}):
				return
			case bytes.Equal(key, []byte{13}):
				err := sendMessage(c, "select_current_result", "")
				fatalOnError(err)
				return
			case bytes.Equal(key, []byte{27, 91, 65}):
				err := sendMessage(c, "highlight_prev_result", "")
				fatalOnError(err)
			case bytes.Equal(key, []byte{27, 91, 66}):
				err := sendMessage(c, "highlight_next_result", "")
				fatalOnError(err)
			default:
				r, _ := utf8.DecodeRune(key)
				fmt.Printf("Unknown key pressed (%c)\n", r)
				return
			}
		case cmd := <-wsCmdCh:
			switch cmd.Type {
			case "browser_has_focus":
				return
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func main() {
	if cli.isServer {
		serverMain(cli.args)
	} else {
		childMain(cli.args)
	}
}

func init() {
	flag.BoolVar(&cli.isServer, "server", false, "run this as a server")
	flag.Parse()

	u, err := user.Current()
	fatalOnError(err)
	cli.homedir = u.HomeDir
	cli.configDir = path.Join(cli.homedir, ".go-ogle")

	cli.args = flag.Args()
}

////////////////////////////////////////////////////////////////////////////////
