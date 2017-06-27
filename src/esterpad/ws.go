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
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"strings"
)

var (
	wsLogger = LogInit("ws")
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		}, //TODO fix
	}
)

func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		wsLogger.Log(LOG_ERROR, "upgrader.Upgrade", err)
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
