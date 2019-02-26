package tester

import (
	"os"
	"sync"
	"time"

	pb "github.com/anon2anon/esterpad/internal/proto"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	id               int
	writers          int
	magicWordChannel chan string
	padName          string
	text             []rune
	opsMap           map[uint32][]*pb.Op
	revision         uint32
	mutex            *sync.Mutex
	logger           *log.Entry
}

func (c *Client) RenderDelta(ops []*pb.Op) {
	pos := 0
	for _, op := range ops {
		switch op := op.Op.(type) {
		case *pb.Op_Insert:
			insertText := []rune(op.Insert.Text)
			newText := make([]rune, len(c.text)+len(insertText))
			copy(newText, c.text[:pos])
			copy(newText[pos:], insertText)
			copy(newText[pos+len(insertText):], c.text[pos:])
			c.text = newText
			pos += len(insertText)
		case *pb.Op_Delete:
			delPos := pos + int(op.Delete.Len)
			if delPos >= len(c.text) {
				c.text = c.text[:pos]
			} else {
				newText := make([]rune, len(c.text)-int(op.Delete.Len))
				copy(newText, c.text[:pos])
				copy(newText[pos:], c.text[delPos:])
				c.text = newText
			}
		case *pb.Op_Retain:
			pos += int(op.Retain.Len)
			if pos > len(c.text) {
				pos = len(c.text)
			}
		}
	}
}

func (c *Client) FindMagicWord() {
	text := c.text
	end := len(text) - 4 - 16
	for i := 0; i < end; i++ {
		if string(text[i:i+5]) == "FLAG_" {
			c.magicWordChannel <- string(text[i+5 : i+21])
			text[i+4] = '*'
		}
	}
}

func (c *Client) GenerateFlag() *pb.CDelta {
	ops := []*pb.Op{}
	c.mutex.Lock()
	end := len(c.text)
	revision := c.revision
	c.mutex.Unlock()
	left := Random.Intn(end/22+1) * 22
	if left > 0 {
		ops = append(ops, &pb.Op{Op: &pb.Op_Retain{&pb.OpRetain{Len: uint32(left)}}})
	}
	flag := GenRandomString(16)
	ops = append(ops, &pb.Op{Op: &pb.Op_Insert{&pb.OpInsert{Text: "FLAG_" + flag + "\n"}}})
	if end > left {
		ops = append(ops, &pb.Op{Op: &pb.Op_Retain{&pb.OpRetain{Len: uint32(end - left)}}})
	}
	c.magicWordChannel <- flag
	return &pb.CDelta{Revision: revision, Ops: ops}
}

func (c *Client) GenerateDelta() *pb.CDelta {
	ops := []*pb.Op{}
	c.mutex.Lock()
	end := len(c.text)
	revision := c.revision
	c.mutex.Unlock()
	typ := 0
	if end < 22 {
		typ = 1
	} else if end < 100*22 {
		typ = Random.Intn(2)
	}
	if typ == 0 {
		left := Random.Intn(end/22) * 22
		right := Random.Intn(end/22+1) * 22
		if left > right {
			t := left
			left = right
			right = t
		}
		if left > 0 {
			ops = append(ops, &pb.Op{Op: &pb.Op_Retain{&pb.OpRetain{Len: uint32(left)}}})
		}
		ops = append(ops, &pb.Op{Op: &pb.Op_Delete{&pb.OpDelete{Len: uint32(right - left)}}})
		if end > right {
			ops = append(ops, &pb.Op{Op: &pb.Op_Retain{&pb.OpRetain{Len: uint32(end - right)}}})
		}
	} else {
		left := Random.Intn(end/22+1) * 22
		num := Random.Intn(3) + 1
		if left > 0 {
			ops = append(ops, &pb.Op{Op: &pb.Op_Retain{&pb.OpRetain{Len: uint32(left)}}})
		}
		for i := 0; i < num; i++ {
			ops = append(ops, &pb.Op{Op: &pb.Op_Insert{&pb.OpInsert{Text: GenRandomString(21) + "\n"}}})
		}
		if end > left {
			ops = append(ops, &pb.Op{Op: &pb.Op_Retain{&pb.OpRetain{Len: uint32(end - left)}}})
		}
	}
	return &pb.CDelta{Revision: revision, Ops: ops}
}

func (c *Client) Write(wsConn *websocket.Conn) {
	time.Sleep(time.Duration(c.writers) * time.Second / 10)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for _ = range ticker.C {
		ops := (*pb.CDelta)(nil)
		if int(c.revision)%(3*c.writers) == 3*c.id {
			ops = c.GenerateFlag()
		} else {
			ops = c.GenerateDelta()
		}
		smessage := &pb.CMessage{CMessage: &pb.CMessage_Delta{ops}}
		dataBytes, err := proto.Marshal(&pb.CMessages{Cm: []*pb.CMessage{smessage}})
		if err != nil {
			c.logger.WithError(err).Error("marshal")
			return
		}
		if err := wsConn.WriteMessage(websocket.BinaryMessage, dataBytes); err != nil {
			c.logger.WithError(err).Error("ws write")
			return
		}
	}
}

func (c *Client) Process(wsConn *websocket.Conn) {
	message1 := pb.CSession{SessId: ""}
	message2 := pb.CEnterPad{Name: c.padName}
	smessage1 := &pb.CMessage{CMessage: &pb.CMessage_Session{&message1}}
	smessage2 := &pb.CMessage{CMessage: &pb.CMessage_EnterPad{&message2}}
	welcomeDataBytes, err := proto.Marshal(&pb.CMessages{Cm: []*pb.CMessage{smessage1, smessage2}})
	if err != nil {
		c.logger.WithError(err).Error("marshal")
		return
	}
	if err := wsConn.WriteMessage(websocket.BinaryMessage, welcomeDataBytes); err != nil {
		c.logger.WithError(err).Error("ws write")
		return
	}
	if c.id < c.writers {
		go c.Write(wsConn)
	}
	for {
		_, dataBytes, err := wsConn.ReadMessage()
		if err != nil {
			c.logger.Error("server gone")
			break
		}
		messages := &pb.SMessages{}
		err = proto.Unmarshal(dataBytes, messages)
		if err != nil {
			c.logger.WithError(err).Error("unmarshal err")
			break
		}
		for _, m := range messages.Sm {
			switch m := m.SMessage.(type) {
			case *pb.SMessage_Delta:
				if c.revision+1 == m.Delta.Id {
					c.mutex.Lock()
					c.RenderDelta(m.Delta.Ops)
					c.FindMagicWord()
					for true {
						c.revision += 1
						ops, exist := c.opsMap[c.revision+1]
						if !exist {
							break
						}
						c.RenderDelta(ops)
						c.FindMagicWord()
					}
					c.mutex.Unlock()
				} else if c.revision+1 < m.Delta.Id {
					_, exist := c.opsMap[m.Delta.Id]
					if exist {
						c.logger.WithField("deltaId", m.Delta.Id).Info("server rewrites delta id")
					}
					c.opsMap[m.Delta.Id] = m.Delta.Ops
				}
			case *pb.SMessage_Document:
				c.mutex.Lock()
				c.text = []rune{}
				c.RenderDelta(m.Document.Ops)
				c.FindMagicWord()
				c.revision = m.Document.Revision
				c.mutex.Unlock()
				c.opsMap = map[uint32][]*pb.Op{}
			}
		}
	}
}
func (c *Client) Connect(serverUrl string) {
	var dialer *websocket.Dialer
	wsConn, _, err := dialer.Dial("ws://"+serverUrl+"/.ws", nil)
	if err != nil {
		c.logger.WithError(err).Error("websocket connect")
		os.Exit(1)
		return
	}
	go c.Process(wsConn)
}
