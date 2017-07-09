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

package esterpad_tester

import (
	. "esterpad_utils"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MagicWordInfo struct {
	StartTime   time.Time
	clientCount int
}

func PrintUsage() {
	fmt.Println("Usage:", os.Args[0], "[server address] [number of total clients] [number of writers] [pad name]")
	fmt.Println("  server address - Server address in host:port format (ex. localhost:9000)")
	fmt.Println("  number of total clients - number of clients that read text from pad")
	fmt.Println("  number of writers - number of clients that write text to pad")
	fmt.Println("  pad name - optional parameter, name of pad which clients will connect")
	os.Exit(1)
}

var Random *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func GenRandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	ret := make([]byte, n)
	for i := range ret {
		ret[i] = charset[Random.Intn(len(charset))]
	}
	return string(ret)
}

func MagicWordReceiver(clients int, magicWordChannel chan string) {
	m := map[string]*MagicWordInfo{}
	testCount := 0
	maxDelta := time.Duration(0)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	for {
		select {
		case <-signalChan:
			fmt.Println("Testing stopped,", testCount, "tests performed, max responce time", maxDelta)
			return
		case word, ok := <-magicWordChannel:
			if !ok {
				return
			}
			info, exist := m[word]
			if !exist {
				startTime := time.Now()
				info = &MagicWordInfo{startTime, 1}
				m[word] = info
				fmt.Println("New test started, word", word, ", start time", startTime)
			} else {
				info.clientCount++
			}
			if info.clientCount >= clients+1 {
				endTime := time.Now()
				delta := endTime.Sub(info.StartTime)
				if delta > maxDelta {
					maxDelta = delta
				}
				fmt.Println("Test success, word", word, ", end time ", endTime, ", delta", endTime.Sub(info.StartTime))
				delete(m, word)
				testCount++
			}
		}
	}
}

func Main() {
	if len(os.Args) < 4 {
		PrintUsage()
	}
	serverUrl := os.Args[1]
	if strings.HasPrefix(serverUrl, "http:") {
		serverUrl = serverUrl[5:]
	}
	for len(serverUrl) > 0 && serverUrl[0] == '/' {
		serverUrl = serverUrl[1:]
	}
	clients, err := strconv.Atoi(os.Args[2])
	if err != nil || clients <= 0 {
		fmt.Println("Invalid number of total clients", os.Args[2])
		os.Exit(1)
	}
	writers, err := strconv.Atoi(os.Args[3])
	if err != nil || writers <= 0 || writers > clients {
		fmt.Println("Invalid number of writers", os.Args[3])
		os.Exit(1)
	}
	padName := ""
	if len(os.Args) < 5 {
		padName = GenRandomString(16)
	} else {
		padName = os.Args[4]
	}
	magicWordChannel := make(chan string, clients)
	for i := 0; i < clients; i++ {
		c := Client{
			id: i, writers: writers, magicWordChannel: magicWordChannel,
			padName: padName, mutex: &sync.Mutex{}, opsMap: map[uint32][]*Op{}}
		c.Connect(serverUrl)
	}
	MagicWordReceiver(clients, magicWordChannel)
}
