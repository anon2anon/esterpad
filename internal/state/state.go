package state

import (
	"strconv"
	"strings"
	"sync"

	ep "github.com/anon2anon/esterpad/internal/types"
)

type State struct {
	mongo      ep.PersistentStorage
	clients    map[uint32]*ep.Connection
	clientsMux sync.RWMutex
	users      map[uint32]*ep.User
	usersMux   sync.RWMutex
	pads       map[uint32]*ep.Pad
	padsMux    sync.RWMutex
}

func (s *State) AddUser(user *ep.User) {
	s.usersMux.Lock()
	UserCounter++
	user.Id = UserCounter
	UserMap[UserCounter] = user
	s.usersMux.Unlock()
	if email != "" {
		s.mongo.RegisterFinish(user, email.(string))
	} else {
		user.Nickname = "guest-" + strconv.FormatInt(int64(user.Id), 10)
	}
	s.mongo.RegisterUser(user)
}

func (s *State) GetUser(userId uint32) *ep.User {
	s.usersMux.RLock()
	user := UserMap[userId]
	s.usersMux.RUnlock()
	return user
}

func (s *State) GetPad(name string) *ep.Pad {
	name = strings.TrimSpace(name)
	if len(name) == 0 || strings.IndexRune(name, '/') >= 0 || strings.IndexRune(name, '.') >= 0 {
		return nil
	}
	needInsert := false
	s.padsMux.Lock()
	pad := PadMap[name]
	if pad == nil {
		PadCounter++
		pad = PadLoad(PadCounter, name)
		PadMap[name] = pad
		needInsert = true
	}
	s.padsMux.Unlock()
	if needInsert {
		s.mongo.InsertPad(pad.Id, pad.Name)
		update := pb.SPadList{Pads: []string{pad.Name}}
		s.BroadcastClientMessage(update)
	}
	return pad
}

func (s *State) BroadcastClientMessage(msg interface{}) {
	s.clientsMux.RLock()
	for _, client := range s.clients {
		select {
		case client.Messages <- msg:
		default:
		}
	}
	s.clientsMux.RUnlock()
}

func New(mongo ep.PersistentStorage) (*State, error) {
	state := State{mongo: mongo}

	users, err := mongo.LoadUsers()
	if err != nil {
		return nil, err
	}
	state.users = *users

	pads, err := mongo.LoadPads()
	if err != nil {
		return nil, err
	}
	state.pads = *pads

	return &state, nil
}
