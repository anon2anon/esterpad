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
.	"esterpad_utils"
	"container/list"
	"gopkg.in/mgo.v2"
	"strconv"
	"strings"
	"sync"
)

var (
	padLogger       = LogInit("pad")
	DefaultDocument = list.New()
)

type PChat struct {
	Id   uint32
	User *User
	Text string
}

type PDelta struct {
	Id     uint32
	UserId uint32
	Ops    *list.List
}

type PDocument struct {
	Revision uint32
	Ops      *list.List
}

type POpInsert struct {
	Text   []rune
	Meta   *PMeta
	TextRO bool
}

type POpDelete struct {
	Len uint32
}

type POpRetain struct {
	Len  uint32
	Meta *PMeta
}

type PMeta struct {
	Changemask uint32
	Bold       bool
	Italic     bool
	Underline  bool
	Strike     bool
	FontSize   uint32
	UserId     uint32
	User       *User
}

type Pad struct {
	Id              uint32
	Name            string
	CacherChannel   chan interface{}
	Clients         *list.List
	ClientsMutex    sync.RWMutex
	ChatCounter     uint32
	ChatArray       []*PChat
	ChatMutex       sync.RWMutex
	DeltaArray      []*PDelta
	DocumentArray   []*PDocument
	DeltaCounter    uint32
	DeltaMutex      sync.RWMutex
	ChatCollection  *mgo.Collection
	DeltaCollection *mgo.Collection
}

func PadLoad(id uint32, name string) *Pad {
	p := Pad{Id: id, Name: name, CacherChannel: make(chan interface{}, 200), Clients: list.New(),
		ClientsMutex: sync.RWMutex{}, ChatMutex: sync.RWMutex{}, DeltaMutex: sync.RWMutex{},
		DocumentArray: []*PDocument{&PDocument{Ops: DefaultDocument}}}
	p.ChatCollection = MongoConnection.DB("").C("chat" + strconv.FormatInt(int64(p.Id), 10))
	p.DeltaCollection = MongoConnection.DB("").C("delta" + strconv.FormatInt(int64(p.Id), 10))
	chatIter := p.ChatCollection.Find(nil).Sort("_id").Iter()
	chat := MongoChat{}
	for chatIter.Next(&chat) {
		for i := p.ChatCounter + 1; i < chat.Id; i++ {
			p.ChatArray = append(p.ChatArray, nil)
		}
		p.ChatCounter = chat.Id
		user := CacherGetUser(chat.UserId)
		p.ChatArray = append(p.ChatArray, &PChat{chat.Id, user, chat.Text})
	}
	if err := chatIter.Close(); err != nil {
		padLogger.Log(LOG_ERROR, p.Id, "mongo find err", err)
	}
	deltaIter := p.DeltaCollection.Find(nil).Sort("_id").Iter()
	delta := MongoDelta{}
	oldDocument := DefaultDocument
	for deltaIter.Next(&delta) {
		for i := p.DeltaCounter + 1; i < delta.Id; i++ {
			p.DeltaArray = append(p.DeltaArray, nil)
			p.DocumentArray = append(p.DocumentArray, &PDocument{i, oldDocument})
		}
		p.DeltaCounter = delta.Id
		newOps := list.New()
		for _, op := range delta.Ops {
			if op.Insert != nil {
				newOps.PushBack(&POpInsert{[]rune(op.Insert.(string)), (*PMeta)(op.Meta), false})
			} else if op.Delete != nil {
				newOps.PushBack(&POpDelete{*op.Delete})
			} else if op.Retain != nil {
				newOps.PushBack(&POpRetain{*op.Retain, (*PMeta)(op.Meta)})
			}
		}
		newDocument := DeltaComposeOld(newOps, oldDocument)
		if newDocument == nil {
			padLogger.Log(LOG_ERROR, p.Id, "can't compose delta on load", DeltaToString(newOps), DeltaToString(oldDocument))
			p.DeltaArray = append(p.DeltaArray, nil)
			p.DocumentArray = append(p.DocumentArray, &PDocument{delta.Id, oldDocument})
		} else {
			pdelta := PDelta{delta.Id, delta.UserId, newOps}
			p.DeltaArray = append(p.DeltaArray, &pdelta)
			p.DocumentArray = append(p.DocumentArray, &PDocument{delta.Id, newDocument})
			oldDocument = newDocument
		}
	}
	if err := deltaIter.Close(); err != nil {
		padLogger.Log(LOG_ERROR, p.Id, "mongo find err", err)
	}

	go p.CacherHandler()
	return &p
}

func (p *Pad) CacherHandler() {
	for {
		select {
		case pmessage, ok := <-p.CacherChannel:
			if !ok {
				return
			}
			switch pmessage := pmessage.(type) {
			case *PChat:
				mongoMessage := &MongoChat{pmessage.Id, pmessage.User.Id, pmessage.Text}
				if err := p.ChatCollection.Insert(mongoMessage); err != nil {
					padLogger.Log(LOG_ERROR, p.Id, "mongo insert err", err)
				}
			case *PDelta:
				mongoMessage := &MongoDelta{
					pmessage.Id, pmessage.UserId,
					make([]*MongoDeltaOp, pmessage.Ops.Len())}
				count := 0
				for op := pmessage.Ops.Front(); op != nil; op = op.Next() {
					mongoOp := MongoDeltaOp{}
					switch op := op.Value.(type) {
					case *POpInsert:
						mongoOp.Insert = string(op.Text)
						mongoOp.Meta = op.Meta
					case *POpDelete:
						mongoOp.Delete = &op.Len
					case *POpRetain:
						mongoOp.Retain = &op.Len
						mongoOp.Meta = op.Meta
					}
					mongoMessage.Ops[count] = &mongoOp
					count++
				}
				if err := p.DeltaCollection.Insert(mongoMessage); err != nil {
					padLogger.Log(LOG_ERROR, p.Id, "mongo insert err", err)
				}
			}
		}
	}
}

func (p *Pad) SendChat(c *Client, clientChat *CChat) {
	text := c.User.Nickname
	if c.User.Perms&PERM_NOTGUEST == 0 {
		text += " (guest)"
	}
	text += ": " + strings.TrimSpace(clientChat.Text)
	pmessage := PChat{User: c.User, Text: text}
	padLogger.Log(LOG_INFO, p.Id, c.UserId, "broadcast chat message", text)
	p.ChatMutex.Lock()
	p.ChatCounter++
	pmessage.Id = p.ChatCounter
	p.ChatArray = append(p.ChatArray, &pmessage)
	p.ChatMutex.Unlock()
	p.CacherChannel <- &pmessage

	p.ClientsMutex.RLock()
	for clientIter := p.Clients.Front(); clientIter != nil; clientIter = clientIter.Next() {
		neighbor := clientIter.Value.(*Client)
		if neighbor != c {
			select {
			case neighbor.Messages <- &pmessage:
			default:
			}
		}
	}
	p.ClientsMutex.RUnlock()
}

func (p *Pad) SendDelta(c *Client, clientDelta *CDelta) {
	padLogger.Log(LOG_INFO, p.Id, c.UserId, "broadcast delta message", clientDelta)
	canWriteWash := c.User.Perms&PERM_WHITEWASH != 0
	//canEdit := c.User.Perms&PERM_EDIT != 0
	opsList := DeltaValidateFromClient(clientDelta.Ops, canWriteWash, c.UserId)
	p.DeltaMutex.Lock()
	for rev := clientDelta.Revision; rev < p.DeltaCounter; rev++ {
		newOpsList := DeltaTransform(opsList, p.DeltaArray[rev].Ops)
		if newOpsList == nil {
			padLogger.Log(LOG_ERROR, p.Id, c.UserId, "can't transform delta", rev, DeltaToString(opsList), DeltaToString(p.DeltaArray[rev].Ops))
			p.DeltaMutex.Unlock()
			return
		}
		opsList = newOpsList
	}
	oldDocument := p.DocumentArray[p.DeltaCounter].Ops
	//newOps := DeltaCompose(opsList, oldDocument, canWriteWash, canEdit, c.User)
	//if newOps[0] == nil {
	newOps := DeltaComposeOld(opsList, oldDocument)
	if newOps == nil {
		padLogger.Log(LOG_ERROR, p.Id, c.UserId, "can't compose delta", DeltaToString(opsList), DeltaToString(oldDocument))
		p.DeltaMutex.Unlock()
		return
	}
	p.DeltaCounter++
	//delta := PDelta{p.DeltaCounter, c.UserId, newOps[0]}
	delta := PDelta{p.DeltaCounter, c.UserId, opsList}
	p.DeltaArray = append(p.DeltaArray, &delta)
	//p.DocumentArray = append(p.DocumentArray, &PDocument{p.DeltaCounter, newOps[1]})
	p.DocumentArray = append(p.DocumentArray, &PDocument{p.DeltaCounter, newOps})
	p.DeltaMutex.Unlock()

	p.CacherChannel <- &delta

	p.ClientsMutex.RLock()
	for clientIter := p.Clients.Front(); clientIter != nil; clientIter = clientIter.Next() {
		neighbor := clientIter.Value.(*Client)
		select {
		case neighbor.Messages <- &delta:
		default:
		}
	}
	p.ClientsMutex.RUnlock()
}

func (p *Pad) InvertDelta(c *Client, id uint32) {
	padLogger.Log(LOG_INFO, p.Id, c.UserId, "process invert delta message", id)
	p.DeltaMutex.Lock()
	if p.DeltaCounter <= id {
		p.DeltaMutex.Unlock()
		return
	}
	opsList := DeltaInvert(p.DeltaArray[id].Ops, p.DocumentArray[id].Ops)
	if opsList == nil {
		padLogger.Log(LOG_ERROR, p.Id, c.UserId, "can't invert delta", id, DeltaToString(p.DeltaArray[id].Ops), DeltaToString(p.DocumentArray[id].Ops))
		p.DeltaMutex.Unlock()
		return
	}
	for rev := id + 1; rev < p.DeltaCounter; rev++ {
		newOpsList := DeltaTransform(opsList, p.DeltaArray[rev].Ops)
		if newOpsList == nil {
			padLogger.Log(LOG_ERROR, p.Id, c.UserId, "can't transform inverted delta", rev, DeltaToString(opsList), DeltaToString(p.DeltaArray[rev].Ops))
			p.DeltaMutex.Unlock()
			return
		}
		opsList = newOpsList
	}
	oldDocument := p.DocumentArray[p.DeltaCounter].Ops
	newOps := DeltaComposeOld(opsList, oldDocument)
	if newOps == nil {
		padLogger.Log(LOG_ERROR, p.Id, c.UserId, "can't compose inverted delta", DeltaToString(opsList), DeltaToString(oldDocument))
		p.DeltaMutex.Unlock()
		return
	}
	p.DeltaCounter++
	delta := PDelta{p.DeltaCounter, c.UserId, opsList}
	p.DeltaArray = append(p.DeltaArray, &delta)
	p.DocumentArray = append(p.DocumentArray, &PDocument{p.DeltaCounter, newOps})
	p.DeltaMutex.Unlock()

	p.CacherChannel <- &delta

	p.ClientsMutex.RLock()
	for clientIter := p.Clients.Front(); clientIter != nil; clientIter = clientIter.Next() {
		neighbor := clientIter.Value.(*Client)
		select {
		case neighbor.Messages <- &delta:
		default:
		}
	}
	p.ClientsMutex.RUnlock()
}

func (p *Pad) InvertUserDelta(c *Client, userId uint32) {
	padLogger.Log(LOG_INFO, p.Id, c.UserId, "process invert user delta message", userId)
	opsList := (*list.List)(nil)
	p.DeltaMutex.Lock()
	for rev := uint32(0); rev < p.DeltaCounter; rev++ {
		if p.DeltaArray[rev].UserId == userId {
			invertedOpsList := DeltaInvert(p.DeltaArray[rev].Ops, p.DocumentArray[rev].Ops)
			if invertedOpsList == nil {
				padLogger.Log(LOG_ERROR, p.Id, c.UserId, "can't invert delta", rev, DeltaToString(p.DeltaArray[rev].Ops), DeltaToString(p.DocumentArray[rev].Ops))
				p.DeltaMutex.Unlock()
				return
			}
			if opsList != nil {
				newOpsList := DeltaComposeOld(opsList, invertedOpsList)
				if newOpsList == nil {
					padLogger.Log(LOG_ERROR, p.Id, c.UserId, "can't compose inverted user delta", rev, DeltaToString(invertedOpsList), DeltaToString(opsList))
					p.DeltaMutex.Unlock()
					return
				}
				opsList = newOpsList
			} else {
				opsList = invertedOpsList
			}
		} else if opsList != nil {
			newOpsList := DeltaTransform(opsList, p.DeltaArray[rev].Ops)
			if newOpsList == nil {
				padLogger.Log(LOG_ERROR, p.Id, c.UserId, "can't transform inverted user delta", rev, DeltaToString(opsList), DeltaToString(p.DeltaArray[rev].Ops))
				p.DeltaMutex.Unlock()
				return
			}
			opsList = newOpsList
		}
	}
	oldDocument := p.DocumentArray[p.DeltaCounter].Ops
	newOps := (*list.List)(nil)
	if opsList != nil {
		newOps = DeltaComposeOld(opsList, oldDocument)
		if newOps == nil {
			padLogger.Log(LOG_ERROR, p.Id, c.UserId, "can't compose inverted user delta", DeltaToString(opsList), DeltaToString(oldDocument))
			p.DeltaMutex.Unlock()
			return
		}
	} else {
		opsList = list.New()
		newOps = oldDocument
	}
	p.DeltaCounter++
	delta := PDelta{p.DeltaCounter, c.UserId, opsList}
	p.DeltaArray = append(p.DeltaArray, &delta)
	p.DocumentArray = append(p.DocumentArray, &PDocument{p.DeltaCounter, newOps})
	p.DeltaMutex.Unlock()

	p.CacherChannel <- &delta

	p.ClientsMutex.RLock()
	for clientIter := p.Clients.Front(); clientIter != nil; clientIter = clientIter.Next() {
		neighbor := clientIter.Value.(*Client)
		select {
		case neighbor.Messages <- &delta:
		default:
		}
	}
	p.ClientsMutex.RUnlock()
}

func (p *Pad) SendUserInfo(c *Client) {
	message := &SUserInfo{
		UserId: c.UserId, Nickname: c.User.Nickname, Color: c.User.Color, Perms: c.User.Perms, Online: true}
	messageMod := SUserInfo{c.UserId, c.User.Nickname, c.User.Color, c.User.Perms, true, c.Ip, c.UserAgent}
	padLogger.Log(LOG_INFO, p.Id, c.UserId, "broadcast user info", &message)

	p.ClientsMutex.RLock()
	for clientIter := p.Clients.Front(); clientIter != nil; clientIter = clientIter.Next() {
		neighbor := clientIter.Value.(*Client)
		if neighbor != c {
			if neighbor.User.Perms&PERM_MOD != 0 {
				select {
				case neighbor.Messages <- &messageMod:
				default:
				}
			} else {
				select {
				case neighbor.Messages <- &message:
				default:
				}
			}
		}
	}
	p.ClientsMutex.RUnlock()
}

func (p *Pad) SendUserLeave(c *Client) {
	message := SUserLeave{c.UserId}
	padLogger.Log(LOG_INFO, p.Id, c.UserId, "broadcast user logout", message)

	p.ClientsMutex.RLock()
	for clientIter := p.Clients.Front(); clientIter != nil; clientIter = clientIter.Next() {
		neighbor := clientIter.Value.(*Client)
		if neighbor != c {
			select {
			case neighbor.Messages <- &message:
			default:
			}
		}
	}
	p.ClientsMutex.RUnlock()
}

func (p *Pad) CopyOnlineUsers() []*Client {
	p.ClientsMutex.RLock()
	count := 0
	ret := make([]*Client, p.Clients.Len())
	for clientIter := p.Clients.Front(); clientIter != nil; clientIter = clientIter.Next() {
		ret[count] = clientIter.Value.(*Client)
		count++
	}
	p.ClientsMutex.RUnlock()
	return ret
}

func (p *Pad) CopyChat(count uint32) []*PChat {
	if count == 0 {
		return nil
	}
	ret := []*PChat{}
	p.ChatMutex.RLock()
	len := uint32(len(p.ChatArray))
	if len == 0 {
		p.ChatMutex.RUnlock()
		return nil
	}
	if count > len {
		count = len
	}
	ret = append(ret, p.ChatArray[len-count:]...)
	p.ChatMutex.RUnlock()
	return ret
}

func (p *Pad) CopyChatFrom(from uint32, count uint32) []*PChat {
	if count > from {
		count = from
	}
	if count == 0 {
		return nil
	}
	to := from - count
	ret := []*PChat{}
	p.ChatMutex.RLock()
	len := uint32(len(p.ChatArray))
	if to >= len {
		p.ChatMutex.RUnlock()
		return nil
	}
	if from > len {
		from = len
	}
	ret = append(ret, p.ChatArray[to:from]...)
	p.ChatMutex.RUnlock()
	return ret
}

func (p *Pad) CopyDocument() *PDocument {
	p.DeltaMutex.RLock()
	if p.DeltaCounter == 0 {
		p.DeltaMutex.RUnlock()
		return nil
	}
	ret := p.DocumentArray[p.DeltaCounter]
	p.DeltaMutex.RUnlock()
	return ret
}

func (p *Pad) CopyDeltaRevision(revision uint32) *PDelta {
	p.DeltaMutex.RLock()
	if p.DeltaCounter <= revision {
		p.DeltaMutex.RUnlock()
		return nil
	}
	ret := p.DeltaArray[revision]
	p.DeltaMutex.RUnlock()
	return ret
}

func (p *Pad) CopyDocumentRevision(revision uint32) *PDocument {
	p.DeltaMutex.RLock()
	if p.DeltaCounter < revision {
		p.DeltaMutex.RUnlock()
		return nil
	}
	ret := p.DocumentArray[revision]
	p.DeltaMutex.RUnlock()
	return ret
}
