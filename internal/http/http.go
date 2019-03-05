// Package http handles serving frontend files and upgrading WebSocket connections
package http

import (
	"net/http"
	"os"
	"path"
	"strings"

	ep "github.com/anon2anon/esterpad/internal/types"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Config for HTTP server
type Config struct {
	Listen           string
	StaticPath       string `yaml:"staticPath"`
	UseXForwardedFor bool   `yaml:"useXForwardedFor"`
}

type httpServer struct {
	conf Config
	env  ep.Env
}

func (h *httpServer) newFileHandler() http.Handler {
	return http.FileServer(&indexFallbackFS{http.Dir(h.conf.StaticPath)})
}

type indexFallbackFS struct {
	assets http.FileSystem
}

// Open tries to open requested file, if it doesn't exist returns index.html
func (i *indexFallbackFS) Open(name string) (http.File, error) {
	ret, err := i.assets.Open(name)
	if !os.IsNotExist(err) || path.Ext(name) != "" {
		return ret, err
	}
	return i.assets.Open("index.html")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	}, //TODO fix
}

func (h *httpServer) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("failed to upgrade to WS")
		return
	}
	defer conn.Close()
	ip := ""
	if h.conf.UseXForwardedFor {
		ip = r.Header.Get("x-forwarded-for")
	} else {
		ip = r.RemoteAddr[:strings.IndexByte(r.RemoteAddr, ':')]
	}
	log.WithField("ip", ip).Info("upgraded client to WS")
	// client := Client{Ip: ip, UserAgent: r.Header.Get("user-agent")}
	// client.Process(conn)
}

// Serve starts serving files and websockets
// Returns error from http.ListenAndServe
func Serve(conf Config, env ep.Env) error {
	server := httpServer{conf, env}
	http.Handle("/", server.newFileHandler())
	http.HandleFunc("/.ws", server.wsHandler)
	log.Info("listening on ", conf.Listen)
	return http.ListenAndServe(conf.Listen, nil)
}
