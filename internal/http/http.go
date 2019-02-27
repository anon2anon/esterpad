package http

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Listen           string
	StaticPath       string `yaml:"staticPath"`
	UseXForwardedFor bool   `yaml:"useXForwardedFor"`
}

func newFileHandler(staticPath string) http.Handler {
	return http.FileServer(&indexFallbackFS{http.Dir(staticPath)})
}

type indexFallbackFS struct {
	assets http.FileSystem
}

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

func newWsHandler(conf Config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.WithError(err).Error("Failed to upgrade to WS")
			return
		}
		defer conn.Close()
		ip := ""
		if conf.UseXForwardedFor {
			ip = r.Header.Get("x-forwarded-for")
		} else {
			ip = r.RemoteAddr[:strings.IndexByte(r.RemoteAddr, ':')]
		}
		log.WithField("ip", ip).Info("upgraded client to WS")
		// client := Client{Ip: ip, UserAgent: r.Header.Get("user-agent")}
		// client.Process(conn)
	}
}

func Serve(conf Config) error {
	http.Handle("/", newFileHandler(conf.StaticPath))
	http.HandleFunc("/.ws", newWsHandler(conf))
	log.Info("Listening on ", conf.Listen)
	return http.ListenAndServe(conf.Listen, nil)
}
