package main

////////////////////////////////////////////////////////////////////////////////

import (
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

	"github.com/gorilla/websocket"

	"github.com/sabhiram/go-ogle/hub"
	"github.com/sabhiram/go-ogle/server"
	"github.com/sabhiram/go-ogle/types"
)

////////////////////////////////////////////////////////////////////////////////

var (
	cli = struct {
		// Common args
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

func spawnServerThread() error {
	cmd := exec.Command("go-ogle", "-server")
	if err := cmd.Start(); err != nil {
		return err
	}
	return setPID(cmd.Process.Pid)
}

func serverRunning() (bool, error) {
	// Check for pid file, if it does not exist return false.
	pid, err := getPID()
	if err != nil {
		return false, err
	}

	// If the pid file exists, check if the pid is running, if
	// not, remove the pid file and return false.
	p, err := os.FindProcess(pid)
	if p == nil || err != nil {
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

func childMain(args []string) {
	// Check to see if the server is running
	running, _ := serverRunning()
	if !running {
		err := spawnServerThread()
		fatalOnError(err)
	}

	u := url.URL{Scheme: "ws", Host: "localhost:18881", Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	fatalOnError(err)
	defer c.Close()

	q := fmt.Sprintf("https://www.google.com/search?q=%s", strings.Join(args, "+"))
	m := types.NewSocketMessage("open_new_tab_with_url", q)
	bs, err := m.Marshal()
	fatalOnError(err)

	err = c.WriteMessage(websocket.TextMessage, bs)
	fatalOnError(err)

	// TESTING: wait 3 sec, choose next result, wait 3 sec select it.
	{
		<-time.After(3 * time.Second)
		m.Type = "next_result"
		bs, err = m.Marshal()
		fatalOnError(err)

		err = c.WriteMessage(websocket.TextMessage, bs)
		fatalOnError(err)

		<-time.After(3 * time.Second)
		m.Type = "select_current_result"
		bs, err = m.Marshal()
		fatalOnError(err)

		err = c.WriteMessage(websocket.TextMessage, bs)
		fatalOnError(err)

		<-time.After(3 * time.Second)
	}

	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	fatalOnError(err)
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
