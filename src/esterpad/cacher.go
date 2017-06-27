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
	"strconv"
	"strings"
	"sync"
)

var (
	cacherLogger         = LogInit("cacher")
	cacherChannel        = make(chan interface{}, 200)
	UserMap              = map[uint32]*User{}
	UserMutex            = &sync.RWMutex{}
	UserCounter   uint32 = 0
	PadMap               = map[string]*Pad{}
	PadMutex             = &sync.RWMutex{}
	PadCounter    uint32 = 0
)

func CacherClearAll() {
	PadMutex.Lock()
	for _, p := range PadMap {
		p.ChatMutex.Lock()
		p.ChatArray = []*PChat{}
		p.ChatCounter = 0
		p.ChatCollection.RemoveAll(nil)
		p.ChatMutex.Unlock()
		p.DeltaMutex.Lock()
		p.DeltaArray = []*PDelta{}
		p.DocumentArray = []*PDocument{}
		p.DeltaCounter = 0
		p.DeltaCollection.RemoveAll(nil)
		p.DeltaMutex.Unlock()
	}
	PadMap = map[string]*Pad{}
	PadCounter = 0
	PadCollection.RemoveAll(nil)
	PadMutex.Unlock()
	UserMutex.Lock()
	UserMap = map[uint32]*User{}
	UserCounter = 0
	UserCollection.RemoveAll(nil)
	UserMutex.Unlock()
}

func CacherAddUser(user *User, email interface{}) {
	UserMutex.Lock()
	UserCounter++
	user.Id = UserCounter
	UserMap[UserCounter] = user
	UserMutex.Unlock()
	if email != nil {
		MongoRegisterFinish(user, email.(string))
	} else {
		user.Nickname = "guest-" + strconv.FormatInt(int64(user.Id), 10)
		MongoRegisterGuest(user)
	}
}

func CacherGetPad(name string) *Pad {
	name = strings.TrimSpace(name)
	if len(name) == 0 || strings.IndexRune(name, '/') >= 0 || strings.IndexRune(name, '.') >= 0 {
		return nil
	}
	needInsert := false
	PadMutex.Lock()
	pad := PadMap[name]
	if pad == nil {
		PadCounter++
		pad = PadLoad(PadCounter, name)
		PadMap[name] = pad
		needInsert = true
	}
	PadMutex.Unlock()
	if needInsert {
		MongoInsertPad(pad.Id, pad.Name)
		message := SPadList{[]string{pad.Name}}
		GlobalClientsMutex.RLock()
		for clientIter := GlobalClients.Front(); clientIter != nil; clientIter = clientIter.Next() {
			client := clientIter.Value.(*Client)
			select {
			case client.Messages <- &message:
			default:
			}
		}
		GlobalClientsMutex.RUnlock()
	}
	return pad
}

func CacherGetUser(userId uint32) *User {
	UserMutex.RLock()
	user := UserMap[userId]
	UserMutex.RUnlock()
	return user
}

func CacherInit() {
	userIter := UserCollection.Find(nil).Iter()
	user := MongoUser{}
	for userIter.Next(&user) {
		if UserCounter < user.UserId {
			UserCounter = user.UserId
		}
		UserMap[user.UserId] = &User{user.UserId, user.Nickname, user.Color, user.Perms}
	}
	if err := userIter.Close(); err != nil {
		mongoLogger.Log(LOG_ERROR, "mongo find err", err)
	}
	padIter := PadCollection.Find(nil).Sort("_id").Iter()
	pad := MongoPad{}
	for padIter.Next(&pad) {
		PadCounter = pad.Id
		PadMap[pad.Name] = PadLoad(pad.Id, pad.Name)
	}
	if err := padIter.Close(); err != nil {
		mongoLogger.Log(LOG_ERROR, "mongo find err", err)
	}
}
