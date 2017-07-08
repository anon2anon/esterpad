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
	"container/list"
	"crypto/rand"
	"encoding/hex"
	. "esterpad_utils"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	PERM_NOTGUEST  = 1
	PERM_CHAT      = 1 << 1
	PERM_WRITE     = 1 << 2
	PERM_EDIT      = 1 << 3
	PERM_WHITEWASH = 1 << 4
	PERM_MOD       = 1 << 5
	PERM_ADMIN     = 1 << 6
)

const (
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 64 * 1024 * 1024
)

type ClientEnterPad struct {
}

type ClientLeavePad struct {
}

type ClientPadContext struct {
	MaxChatId  uint32
	MaxDeltaId uint32
	SentUsers  map[uint32]bool
}

type User struct {
	Id       uint32
	Nickname string
	Color    uint32
	Perms    uint32
}

type Client struct {
	Messages  chan interface{}
	User      *User
	UserId    uint32
	SessId    [16]byte
	Ip        string
	UserAgent string
	Pad       *Pad
	pc        *ClientPadContext
}

type SessionInfo struct {
	User      *User
	StartTime time.Time
}

var (
	clientLogger        = LogInit("client")
	ClientSessions      = map[[16]byte]*SessionInfo{}
	ClientSessionsMutex = &sync.RWMutex{}
	GlobalClients       = list.New()
	GlobalClientsMutex  = &sync.RWMutex{}
)

func (c *Client) AddUserInfo(buffer []*SMessage, user *User) []*SMessage {
	if c.User != user {
		_, exist := c.pc.SentUsers[user.Id]
		if !exist {
			smessage := &SUserInfo{
				UserId: user.Id, Nickname: user.Nickname, Color: user.Color,
				Perms: user.Perms, Online: false}
			clientLogger.Log(LOG_INFO, c.UserId, "send author", smessage)
			SMessageOneOf := &SMessage_UserInfo{smessage}
			buffer = append(buffer, &SMessage{SMessageOneOf})
			c.pc.SentUsers[user.Id] = false
		}
	}
	return buffer
}

func (c *Client) AddAllUsersFromOps(buffer []*SMessage, ops *list.List) []*SMessage {
	for op := ops.Front(); op != nil; op = op.Next() {
		switch op := op.Value.(type) {
		case *POpInsert:
			if op.Meta.Changemask&32 != 0 && op.Meta.User != nil {
				buffer = c.AddUserInfo(buffer, op.Meta.User)
			}
		case *POpRetain:
			if op.Meta.Changemask&32 != 0 && op.Meta.User != nil {
				buffer = c.AddUserInfo(buffer, op.Meta.User)
			}
		}
	}
	return buffer
}

func (c *Client) AddOfflineInfo(buffer []*SMessage) []*SMessage {
	for _, client := range c.Pad.CopyOnlineUsers() {
		if c != client {
			user := client.User
			if user != nil {
				smessage := &SUserInfo{
					UserId: user.Id, Nickname: user.Nickname, Color: user.Color,
					Perms: user.Perms, Online: true}
				if c.User.Perms&PERM_MOD != 0 {
					smessage.Ip = client.Ip
					smessage.UserAgent = client.UserAgent
				}
				clientLogger.Log(LOG_INFO, c.UserId, "send online user", smessage)
				SMessageOneOf := &SMessage_UserInfo{smessage}
				buffer = append(buffer, &SMessage{SMessageOneOf})
				c.pc.SentUsers[user.Id] = false
			}
		}
	}
	offlineChat := c.Pad.CopyChat(50)
	if offlineChat != nil {
		for _, pmessage := range offlineChat {
			if pmessage != nil {
				c.pc.MaxChatId = pmessage.Id
				smessage := &SChat{pmessage.Id, pmessage.User.Id, pmessage.Text}
				clientLogger.Log(LOG_INFO, c.UserId, "send history chat message", smessage)
				if pmessage.User != nil {
					buffer = c.AddUserInfo(buffer, pmessage.User)
				}
				SMessageOneOf := &SMessage_Chat{smessage}
				buffer = append(buffer, &SMessage{SMessageOneOf})
			}
		}
	}

	offlineDocument := c.Pad.CopyDocument()
	if offlineDocument != nil {
		c.pc.MaxDeltaId = offlineDocument.Revision
		buffer = c.AddAllUsersFromOps(buffer, offlineDocument.Ops)
		smessage := &SDocument{offlineDocument.Revision, DeltaToProtobuf(offlineDocument.Ops)}
		clientLogger.Log(LOG_INFO, c.UserId, "send document message", smessage)
		SMessageOneOf := &SMessage_Document{smessage}
		buffer = append(buffer, &SMessage{SMessageOneOf})
	}
	return buffer
}

func (c *Client) WritePumpProcessChan(message interface{}, buffer []*SMessage) []*SMessage {
	switch message := message.(type) {
	case *PChat:
		if c.pc != nil && message.Id > c.pc.MaxChatId {
			buffer = c.AddUserInfo(buffer, message.User)
			smessage := &SChat{message.Id, message.User.Id, message.Text}
			clientLogger.Log(LOG_INFO, c.UserId, "send broadcast chat message ", smessage)
			SMessageOneOf := &SMessage_Chat{smessage}
			buffer = append(buffer, &SMessage{SMessageOneOf})
		}
	case *PDelta:
		if c.pc != nil && message.Id > c.pc.MaxDeltaId {
			buffer = c.AddAllUsersFromOps(buffer, message.Ops)
			smessage := &SDelta{message.Id, message.UserId, DeltaToProtobuf(message.Ops)}
			clientLogger.Log(LOG_INFO, c.UserId, "send broadcast new delta message", smessage)
			SMessageOneOf := &SMessage_Delta{smessage}
			buffer = append(buffer, &SMessage{SMessageOneOf})
		}
	case *SDeltaDropped:
		if c.pc != nil {
			clientLogger.Log(LOG_INFO, c.UserId, "send delta dropped message", message)
			SMessageOneOf := &SMessage_DeltaDropped{message}
			buffer = append(buffer, &SMessage{SMessageOneOf})
		}
	case *SUserLeave:
		if c.pc != nil {
			clientLogger.Log(LOG_INFO, c.UserId, "send broadcast user logout", message)
			SMessageOneOf := &SMessage_UserLeave{message}
			buffer = append(buffer, &SMessage{SMessageOneOf})
		}
	case *SUserInfo:
		if c.pc != nil {
			clientLogger.Log(LOG_INFO, c.UserId, "send broadcast user info", message)
			c.pc.SentUsers[message.UserId] = false
			SMessageOneOf := &SMessage_UserInfo{message}
			buffer = append(buffer, &SMessage{SMessageOneOf})
		}
	case *SAuth:
		clientLogger.Log(LOG_INFO, c.UserId, "send auth success", message)
		SMessageOneOf := &SMessage_Auth{message}
		buffer = append(buffer, &SMessage{SMessageOneOf})
	case *SAuthError:
		clientLogger.Log(LOG_INFO, c.UserId, "send auth error", message)
		SMessageOneOf := &SMessage_AuthError{message}
		buffer = append(buffer, &SMessage{SMessageOneOf})
	case *SPadList:
		clientLogger.Log(LOG_INFO, c.UserId, "send pad list", message)
		SMessageOneOf := &SMessage_PadList{message}
		buffer = append(buffer, &SMessage{SMessageOneOf})
	case *CChatRequest:
		if c.pc != nil {
			clientLogger.Log(LOG_INFO, c.UserId, "processs chat request", message)
			offlineChat := c.Pad.CopyChatFrom(message.From, message.Count)
			if offlineChat != nil {
				for i := len(offlineChat); i > 0; i-- {
					pmessage := offlineChat[i-1]
					if pmessage != nil {
						if pmessage.User != nil {
							buffer = c.AddUserInfo(buffer, pmessage.User)
						}
						smessage := &SChat{pmessage.Id, pmessage.User.Id, pmessage.Text}
						SMessageOneOf := &SMessage_Chat{smessage}
						buffer = append(buffer, &SMessage{SMessageOneOf})
					}
				}
			}
		}
	case *CRevisionRequest:
		if c.pc != nil {
			clientLogger.Log(LOG_INFO, c.UserId, "processs revision request", message.Revision)
			document := c.Pad.CopyDocumentRevision(message.Revision)
			if document != nil {
				buffer = c.AddAllUsersFromOps(buffer, document.Ops)
				smessage := &SDocument{document.Revision, DeltaToProtobuf(document.Ops)}
				SMessageOneOf := &SMessage_Document{smessage}
				buffer = append(buffer, &SMessage{SMessageOneOf})
			}
			delta := c.Pad.CopyDeltaRevision(message.Revision)
			if delta != nil {
				buffer = c.AddAllUsersFromOps(buffer, delta.Ops)
				smessage := &SDelta{delta.Id, delta.UserId, DeltaToProtobuf(delta.Ops)}
				SMessageOneOf := &SMessage_Delta{smessage}
				buffer = append(buffer, &SMessage{SMessageOneOf})
			}
		}
	case ClientEnterPad:
		c.pc = &ClientPadContext{SentUsers: map[uint32]bool{}}
		buffer = c.AddOfflineInfo(buffer)
	case ClientLeavePad:
		c.pc = nil
	case *User:
		if c.pc != nil {
			clientLogger.Log(LOG_INFO, c.UserId, "process user info change", message)
			if _, exist := c.pc.SentUsers[message.Id]; exist {
				smessage := &SUserInfo{
					UserId: message.Id, Nickname: message.Nickname, Color: message.Color,
					Perms: message.Perms, Online: false}
				SMessageOneOf := &SMessage_UserInfo{smessage}
				buffer = append(buffer, &SMessage{SMessageOneOf})
			}

		}
	}
	return buffer
}

func (c *Client) WritePump(wsConn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		buffer := []*SMessage{}
	clientwrite1:
		for {
			select {
			case message, ok := <-c.Messages:
				if !ok {
					clientLogger.Log(LOG_INFO, c.UserId, "client closed")
					return
				}
				buffer = c.WritePumpProcessChan(message, buffer)
				if len(buffer) > 0 {
					break clientwrite1
				}
			case <-ticker.C:
				if err := wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
					clientLogger.Log(LOG_ERROR, c.UserId, "ws write ping err", err)
					wsConn.Close()
					return
				}
			}
		}
	clientwrite2:
		for {

			select {
			case message, ok := <-c.Messages:
				if !ok {
					clientLogger.Log(LOG_INFO, c.UserId, "client closed")
					return
				}
				buffer = c.WritePumpProcessChan(message, buffer)
			default:
				break clientwrite2
			}
		}
		data, err := proto.Marshal(&SMessages{Sm: buffer})
		if err == nil {
			if err := wsConn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				clientLogger.Log(LOG_ERROR, c.UserId, "ws write err", err)
				wsConn.Close()
				return
			}
		} else {
			clientLogger.Log(LOG_ERROR, c.UserId, "marshal err", err)
			return
		}
	}
}

func (c *Client) AuthSession(sessIdString string) bool {
	if sessIdSlice, err := hex.DecodeString(sessIdString); err == nil && len(sessIdSlice) == 16 {
		sessId := [16]byte{}
		copy(sessId[:], sessIdSlice)
		ClientSessionsMutex.RLock()
		sessInfo, exist := ClientSessions[sessId]
		ClientSessionsMutex.RUnlock()
		if exist {
			c.SessId = sessId
			c.User = sessInfo.User
			c.UserId = c.User.Id
			return true
		} else {
			clientLogger.Log(LOG_ERROR, c.UserId, "sessId doesn't exist", sessId)
		}
	}
	return false
}

func (c *Client) AuthNew(user *User) bool {
	sessId := [16]byte{}
	if _, err := rand.Read(sessId[:]); err == nil {
		c.SessId = sessId
		ClientSessionsMutex.Lock()
		ClientSessions[sessId] = &SessionInfo{user, time.Now()}
		ClientSessionsMutex.Unlock()
		c.User = user
		c.UserId = user.Id
		return true
	} else {
		clientLogger.Log(LOG_ERROR, c.UserId, "sessId gen err", err)
	}
	return false
}

func (c *Client) NewUser(email string, password string, nickname string) uint32 {
	if len(email) == 0 {
		return 2
	}
	colorBytes := [3]byte{}
	if _, err := rand.Read(colorBytes[:]); err != nil {
		clientLogger.Log(LOG_ERROR, c.UserId, "color gen err", err)
		return 3
	}
	if !MongoRegister(email, password) {
		return 2
	}
	user := User{
		Nickname: nickname,
		Color:    uint32(colorBytes[0])*256*256 + uint32(colorBytes[1])*256 + uint32(colorBytes[2]),
		Perms:    PERM_NOTGUEST | PERM_CHAT | PERM_WRITE | PERM_EDIT | PERM_WHITEWASH | PERM_MOD | PERM_ADMIN,
	}
	CacherAddUser(&user, email)
	if !c.AuthNew(&user) {
		return 3
	}
	return 0
}

func (c *Client) NewGuest() uint32 {
	colorBytes := [3]byte{}
	if _, err := rand.Read(colorBytes[:]); err != nil {
		clientLogger.Log(LOG_ERROR, c.UserId, "color gen err", err)
		return 3
	}
	user := User{
		Color: uint32(colorBytes[0])*256*256 + uint32(colorBytes[1])*256 + uint32(colorBytes[2]),
		Perms: PERM_CHAT | PERM_WRITE | PERM_EDIT | PERM_WHITEWASH | PERM_MOD | PERM_ADMIN}
	CacherAddUser(&user, nil)
	if !c.AuthNew(&user) {
		return 3
	}
	return 0
}

func (c *Client) Login(email string, password string) uint32 {
	if len(email) == 0 {
		return 2
	}
	userId := MongoLoginUser(email, password)
	if userId == nil {
		return 1
	}
	user := CacherGetUser(userId.(uint32))
	if user == nil || !c.AuthNew(user) {
		return 3
	}
	return 0
}

func (c *Client) LeavePad(clientListIter *list.Element, toWrite bool) {
	if c.Pad != nil {
		c.Pad.ClientsMutex.Lock()
		c.Pad.Clients.Remove(clientListIter)
		c.Pad.ClientsMutex.Unlock()
		c.Pad.SendUserLeave(c)
		if toWrite {
			c.Messages <- ClientLeavePad{}
		}
	}
	c.Pad = nil
}

func (c *Client) SendWelcome(wsConn *websocket.Conn, newSessId bool) {
	message := SAuth{UserId: c.UserId, Nickname: c.User.Nickname, Color: c.User.Color, Perms: c.User.Perms}
	if newSessId {
		message.SessId = hex.EncodeToString(c.SessId[:])
	}
	c.Messages <- &message
	pads := []string{}
	PadMutex.RLock()
	for k := range PadMap {
		pads = append(pads, k)
	}
	PadMutex.RUnlock()
	sort.Strings(pads)
	c.Messages <- &SPadList{pads}
}

func (c *Client) SendGlobalUserInfo(user *User) {
	GlobalClientsMutex.RLock()
	for clientIter := GlobalClients.Front(); clientIter != nil; clientIter = clientIter.Next() {
		neighbor := clientIter.Value.(*Client)
		if neighbor != c {
			select {
			case neighbor.Messages <- user:
			default:
			}
		}
	}
	GlobalClientsMutex.RUnlock()
}

func (c *Client) AdminUser(m *CAdminUser) {
	perms := c.User.Perms
	changemask := m.Changemask
	editedUser := CacherGetUser(m.UserId)
	editedUserPerms := editedUser.Perms
	if editedUser != nil {
		needUpdateGlobal := false
		if changemask&3 != 0 && perms&PERM_ADMIN != 0 ||
			perms&PERM_MOD != 0 && editedUserPerms&PERM_ADMIN == 0 ||
			perms&PERM_NOTGUEST != 0 && editedUserPerms&PERM_NOTGUEST == 0 {
			if changemask&1 != 0 {
				nickname := strings.TrimSpace(m.Nickname)
				editedUser.Nickname = nickname
				MongoChangeNickname(c.UserId, nickname)
			}
			if changemask&2 != 0 {
				editedUser.Color = m.Color
				MongoChangeColor(c.UserId, m.Color)
			}
			needUpdateGlobal = true
		}
		if changemask&4 != 0 && perms&PERM_ADMIN != 0 ||
			perms&PERM_MOD != 0 && editedUserPerms&PERM_ADMIN == 0 {
			newPerms := m.Perms&126 | editedUserPerms&PERM_NOTGUEST
			if perms&PERM_ADMIN == 0 {
				newPerms &= 62
			}
			editedUser.Perms = newPerms
			MongoChangePerms(editedUser.Id, newPerms)
			needUpdateGlobal = true
		}
		if changemask&8 != 0 && perms&PERM_ADMIN != 0 && len(m.Email) > 0 {
			MongoChangeEmail(editedUser.Id, m.Email)
		}
		if changemask&16 != 0 && perms&PERM_ADMIN != 0 {
			MongoChangePassword(editedUser.Id, m.Password)
		}
		if needUpdateGlobal {
			c.SendGlobalUserInfo(editedUser)
		}
	}
}

func (c *Client) Process(wsConn *websocket.Conn) {
	c.Messages = make(chan interface{}, 200)
	padClientIter := (*list.Element)(nil)
	GlobalClientsMutex.Lock()
	globalClientIter := GlobalClients.PushBack(c)
	GlobalClientsMutex.Unlock()
	wsConn.SetReadLimit(maxMessageSize)
	/*wsConn.SetReadDeadline(time.Now().Add(pongWait))
	wsConn.SetPongHandler(func(string) error {
		wsConn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})*/
	go c.WritePump(wsConn)
	for {
		_, dataBytes, err := wsConn.ReadMessage()
		if err != nil {
			clientLogger.Log(LOG_ERROR, c.UserId, "ws read err", err)
			break
		}
		messages := &CMessages{}
		err = proto.Unmarshal(dataBytes, messages)
		if err != nil {
			clientLogger.Log(LOG_ERROR, c.UserId, "unmarshal err", err)
			break
		}
		clientLogger.Log(LOG_INFO, c.UserId, "recv messages", messages)
		for _, m := range messages.Cm {
			switch m := m.CMessage.(type) {
			case *CMessage_EditUser:
				if c.User != nil {
					changemask := m.EditUser.Changemask
					if changemask&1 != 0 {
						nickname := strings.TrimSpace(m.EditUser.Nickname)
						c.User.Nickname = nickname
						MongoChangeNickname(c.UserId, nickname)
					}
					if changemask&2 != 0 {
						c.User.Color = m.EditUser.Color
						MongoChangeColor(c.UserId, m.EditUser.Color)
					}
					if changemask&4 != 0 && c.User.Perms&PERM_NOTGUEST != 0 && len(m.EditUser.Email) > 0 {
						MongoChangeEmail(c.UserId, m.EditUser.Email)
					}
					if changemask&8 != 0 && c.User.Perms&PERM_NOTGUEST != 0 {
						MongoChangePassword(c.UserId, m.EditUser.Password)
					}
					if changemask&3 != 0 {
						c.SendGlobalUserInfo(c.User)
					}
				}
			case *CMessage_Delta:
				if c.User != nil && c.Pad != nil && c.User.Perms&PERM_WRITE != 0 {
					c.Pad.SendDelta(c, m.Delta)
				}
			case *CMessage_Chat:
				if c.User != nil && c.Pad != nil && c.User.Perms&PERM_CHAT != 0 {
					c.Pad.SendChat(c, m.Chat)
				}
			case *CMessage_Logout:
				if c.User != nil {
					c.LeavePad(padClientIter, true)
					c.User = nil
					c.UserId = 0
				}
			case *CMessage_EnterPad:
				if c.User != nil {
					c.LeavePad(padClientIter, true)
					c.Pad = CacherGetPad(m.EnterPad.Name)
					if c.Pad != nil {
						c.Messages <- ClientEnterPad{}
						c.Pad.ClientsMutex.Lock()
						padClientIter = c.Pad.Clients.PushBack(c)
						c.Pad.ClientsMutex.Unlock()
						c.Pad.SendUserInfo(c)
					}
				}
			case *CMessage_LeavePad:
				if c.User != nil {
					c.LeavePad(padClientIter, true)
				}
			case *CMessage_Session:
				if c.User != nil {
					c.LeavePad(padClientIter, true)
					c.User = nil
					c.UserId = 0
				}
				if c.AuthSession(m.Session.SessId) {
					c.SendWelcome(wsConn, false)
				} else {
					c.Messages <- &SAuthError{4}
				}
			case *CMessage_Login:
				if c.User != nil {
					c.LeavePad(padClientIter, true)
					c.User = nil
					c.UserId = 0
				}
				if authError := c.Login(m.Login.Email, m.Login.Password); authError == 0 {
					c.SendWelcome(wsConn, true)
				} else {
					c.Messages <- &SAuthError{authError}
				}
			case *CMessage_Register:
				if c.User != nil {
					c.LeavePad(padClientIter, true)
					c.User = nil
					c.UserId = 0
				}
				if authError := c.NewUser(m.Register.Email, m.Register.Password, m.Register.Nickname); authError == 0 {
					c.SendWelcome(wsConn, true)
				} else {
					c.Messages <- &SAuthError{authError}
				}
			case *CMessage_GuestLogin:
				if authError := c.NewGuest(); authError == 0 {
					c.SendWelcome(wsConn, true)
				} else {
					c.Messages <- &SAuthError{authError}
				}
			case *CMessage_AdminUser:
				if c.User != nil {
					c.AdminUser(m.AdminUser)
				}
			case *CMessage_ChatRequest:
				if c.User != nil && c.Pad != nil {
					c.Messages <- m.ChatRequest
				}
			case *CMessage_RevisionRequest:
				if c.User != nil && c.Pad != nil {
					c.Messages <- m.RevisionRequest
				}
			case *CMessage_InvertDelta:
				if c.User != nil && c.User.Perms&PERM_MOD != 0 && c.Pad != nil {
					c.Pad.InvertDelta(c, m.InvertDelta.Id)
				}
			case *CMessage_InvertUserDelta:
				if c.User != nil && c.User.Perms&PERM_MOD != 0 && c.Pad != nil {
					c.Pad.InvertUserDelta(c, m.InvertUserDelta.UserId)
				}
			case *CMessage_RestoreRevision:
				if c.User != nil && c.User.Perms&PERM_MOD != 0 && c.Pad != nil {
					c.Pad.RestoreRevision(c, m.RestoreRevision.Rev)
				}
			}
		}
	}
	if c.User != nil {
		c.LeavePad(padClientIter, false)
		c.User = nil
		c.UserId = 0
	}
	GlobalClientsMutex.Lock()
	GlobalClients.Remove(globalClientIter)
	GlobalClientsMutex.Unlock()
	close(c.Messages)
}
