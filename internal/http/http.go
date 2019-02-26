package http

import (
	"encoding/hex"
	"fmt"
	"html"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type staticHanlderStruct struct {
	root http.Dir
}

func staticHanlder(root http.Dir) http.Handler {
	return &staticHanlderStruct{root}
}

func staticHTTPError(err error) (msg string, httpStatus int) {
	if os.IsNotExist(err) {
		return "404 Page not found", http.StatusNotFound
	}
	if os.IsPermission(err) {
		return "403 Forbidden", http.StatusForbidden
	}
	return "500 Internal Server Error", http.StatusInternalServerError
}

func (this *staticHanlderStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "405 Method not allowed", 405)
		return
	}

	url := r.URL.Path
	isStaticFile := false
	for i, c := range url {
		if c != '/' {
			if !strings.HasPrefix(url[i:], ".") && strings.Contains(url[i:], ".") {
				isStaticFile = true
			}
			break
		}
	}

	if !isStaticFile {
		url = "/index.html"
	}

	f, err := this.root.Open(url)
	if err != nil {
		msg, code := staticHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		msg, code := staticHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	if d.IsDir() {
		if !strings.HasSuffix(r.URL.Path, "/") {
			w.Header().Set("Location", r.URL.String()+"/")
			w.WriteHeader(http.StatusMovedPermanently)
		} else {
			http.Error(w, "403 Directory Listing Forbidden", http.StatusForbidden)
		}
	} else {
		http.ServeContent(w, r, d.Name(), d.ModTime(), f)
	}
}

func handleStat(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "405 Method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	PadMutex.RLock()
	GlobalClientsMutex.RLock()
	UserMutex.RLock()
	fmt.Fprintf(w, `<html>
<body>
go version: %s<br/>
goroutines count: %d<br/>
len(CacherChannel): %d<br/>
len(UserMap): %d<br/>
len(PadMap): %d<br/>
len(GlobalClientList): %d<br/>
<div style="font-weight: bold; font-size: 16px">Global users:</div>
<table border="1">
<tr><th>Id</th><th>Ip</th><th>User-Agent</th></tr>
`, html.EscapeString(runtime.Version()), runtime.NumGoroutine(), len(cacherChannel),
		len(UserMap), len(PadMap), GlobalClients.Len())
	UserMutex.RUnlock()
	for clientIter := GlobalClients.Front(); clientIter != nil; clientIter = clientIter.Next() {
		client := clientIter.Value.(*Client)
		fmt.Fprintf(w, "    <tr><td>%p</td><td>%s</td><td>%s</td></tr>\n",
			client, html.EscapeString(client.Ip), html.EscapeString(client.UserAgent))
	}
	fmt.Fprint(w, "</table><br/>\n")
	GlobalClientsMutex.RUnlock()
	padMapCopy := map[string]*Pad{}
	for k, v := range PadMap {
		padMapCopy[k] = v
	}
	PadMutex.RUnlock()
	for _, p := range padMapCopy {
		p.ClientsMutex.RLock()
		p.ChatMutex.RLock()
		p.DeltaMutex.RLock()
		fmt.Fprintf(w, `<div style="font-weight: bold; font-size: 16px">Pad num: %d name: %s</div>
    len(PadCacherChannel): %d<br/>
    len(Clients): %d<br/>
    len(ChatArray): %d<br/>
    len(DeltaArray): %d<br/>
    len(DocumentArray): %d<br/>
    <table border="1">
    <tr><th>Id</th><th>UserId</th><th>Nickname</th><th>SessId</th><th>Color</th><th>len(messages)</th></tr>
		`, p.Id, html.EscapeString(p.Name), len(p.CacherChannel),
			p.Clients.Len(), len(p.ChatArray), len(p.DeltaArray), len(p.DocumentArray))
		p.DeltaMutex.RUnlock()
		p.ChatMutex.RUnlock()
		for clientIter := p.Clients.Front(); clientIter != nil; clientIter = clientIter.Next() {
			client := clientIter.Value.(*Client)
			fmt.Fprintf(w, "    <tr><td>%p</td><td>%d</td><td>%s</td><td>%s</td><td>%06X</td><td>%d</td></tr>\n",
				client, client.UserId, html.EscapeString(client.User.Nickname),
				hex.EncodeToString(client.SessId[:]), client.User.Color, len(client.Messages))

		}
		p.ClientsMutex.RUnlock()
		fmt.Fprint(w, "</table><br/>\n")
	}
	fmt.Fprint(w, "</body>\n</html>\n")
}

func handleClearAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "405 Method not allowed", 405)
		return
	}
	CacherClearAll()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	}, //TODO fix
}

func handleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithError(err).Error("upgrader")
		return
	}
	defer conn.Close()
	ip := ""
	if b, _ := strconv.ParseBool(Config["http"]["use-x-forwarded-for"].(string)); b {
		ip = r.Header.Get("x-forwarded-for")
	} else {
		ip = r.RemoteAddr[:strings.IndexByte(r.RemoteAddr, ':')]
	}
	client := Client{Ip: ip, UserAgent: r.Header.Get("user-agent")}
	client.Process(conn)
}

func Init() {
	http.Handle("/", staticHanlder(http.Dir("frontend/dist")))
	http.HandleFunc("/.clearall", handleClearAll)
	http.HandleFunc("/.stat", handleStat)
	http.HandleFunc("/.ws", handleWs)
	httpListen := Config["http"]["listen"].(string)
	log.Info("Listening on ", httpListen)
	err := http.ListenAndServe(httpListen, nil)
	if err != nil {
		log.WithError(err).Error("ListenAndServe")
	}
}
