package state

import (
	"strconv"
	"strings"
	"sync"

	"github.com/anon2anon/esterpad/internal/mongo"
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

func ClearAll() {
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

func AddUser(user *User, email interface{}) {
	UserMutex.Lock()
	UserCounter++
	user.Id = UserCounter
	UserMap[UserCounter] = user
	UserMutex.Unlock()
	if email != nil {
		mongo.RegisterFinish(user, email.(string))
	} else {
		user.Nickname = "guest-" + strconv.FormatInt(int64(user.Id), 10)
		mongo.RegisterGuest(user)
	}
}

func GetPad(name string) *Pad {
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
		mongo.InsertPad(pad.Id, pad.Name)
		message := pb.SPadList{Pads: []string{pad.Name}}
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

func GetUser(userId uint32) *User {
	UserMutex.RLock()
	user := UserMap[userId]
	UserMutex.RUnlock()
	return user
}

func Init() {
	userIter := UserCollection.Find(nil).Iter()
	user := mongo.User{}
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
	pad := mongo.Pad{}
	for padIter.Next(&pad) {
		PadCounter = pad.Id
		PadMap[pad.Name] = PadLoad(pad.Id, pad.Name)
	}
	if err := padIter.Close(); err != nil {
		mongoLogger.Log(LOG_ERROR, "mongo find err", err)
	}
}
