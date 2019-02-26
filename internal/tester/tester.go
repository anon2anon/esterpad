package tester

import (
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	pb "github.com/anon2anon/esterpad/internal/proto"
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
)

type MagicWordInfo struct {
	StartTime   time.Time
	clientCount int
}

func PrintUsage() {
	log.Info("Usage:", os.Args[0], "[server address] [number of total clients] [number of writers] [pad name]")
	log.Info("  server address - Server address in host:port format (ex. localhost:9000)")
	log.Info("  number of total clients - number of clients that read text from pad")
	log.Info("  number of writers - number of clients that write text to pad")
	log.Info("  pad name - optional parameter, name of pad which clients will connect")
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
			log.WithFields(log.Fields{
				"testCount": testCount,
				"maxDelta":  maxDelta,
			}).Info("Testing stopped")
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
				log.WithFields(log.Fields{
					"word":      word,
					"startTime": startTime,
				}).Info("New test started")
			} else {
				info.clientCount++
			}
			if info.clientCount >= clients+1 {
				endTime := time.Now()
				delta := endTime.Sub(info.StartTime)
				if delta > maxDelta {
					maxDelta = delta
				}
				log.WithFields(log.Fields{
					"word":    word,
					"endTime": endTime,
					"delta":   endTime.Sub(info.StartTime),
				}).Info("Test success")
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
	log.AddHook(filename.NewHook())
	serverUrl := os.Args[1]
	if strings.HasPrefix(serverUrl, "http:") {
		serverUrl = serverUrl[5:]
	}
	for len(serverUrl) > 0 && serverUrl[0] == '/' {
		serverUrl = serverUrl[1:]
	}
	clients, err := strconv.Atoi(os.Args[2])
	if err != nil || clients <= 0 {
		log.Error("Invalid number of total clients: ", os.Args[2])
		os.Exit(1)
	}
	writers, err := strconv.Atoi(os.Args[3])
	if err != nil || writers <= 0 || writers > clients {
		log.Error("Invalid number of writers: ", os.Args[3])
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
			id:               i,
			writers:          writers,
			magicWordChannel: magicWordChannel,
			padName:          padName,
			mutex:            &sync.Mutex{},
			opsMap:           map[uint32][]*pb.Op{},
			logger:           log.WithField("client", i),
		}
		c.Connect(serverUrl)
	}
	MagicWordReceiver(clients, magicWordChannel)
}
