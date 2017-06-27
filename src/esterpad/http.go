/*
Esterpad online collaborative editor
Copyright (C) 2017 Anon2Anon

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package esterpad

import (
	"encoding/hex"
	"fmt"
	"html"
	"net/http"
	"os"
	"runtime"
	"strings"
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
			if strings.Contains(url[i:], ".") || strings.HasPrefix(url[i:], "static/") {
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

func HttpStat(w http.ResponseWriter, r *http.Request) {
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

func HttpClearAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "405 Method not allowed", 405)
		return
	}
	CacherClearAll()
}

func HttpInit() {
	httpLogger := LogInit("http")
	http.Handle("/", staticHanlder(http.Dir("frontend/dist")))
	http.HandleFunc("/.clearall", HttpClearAll)
	http.HandleFunc("/.stat", HttpStat)
	http.HandleFunc("/.ws", WsHandler)
	httpListen := Config["http"]["listen"].(string)
	httpLogger.Log(LOG_INFO, "Listening on", httpListen)
	err := http.ListenAndServe(httpListen, nil)
	if err != nil {
		httpLogger.Log(LOG_FATAL, "ListenAndServe", err)
	}
}
