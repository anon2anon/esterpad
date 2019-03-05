package types

import (
	"container/list"
	"time"
)

type Env struct {
	Mongo PersistentStorage
}

type User struct {
	Id       uint32 `bson:"_id,omitempty"`
	Email    string `bson:",omitempty"`
	Passhash []byte `bson:",omitempty"`
	Nickname string
	Color    uint32
	Perms    UserPerms
}

type UserPerms struct {
	NotGuest  bool
	Chat      bool
	Write     bool
	Edit      bool
	Whitewash bool
	Mod       bool
	Admin     bool
}

// type Client struct {
// 	Messages  chan interface{}
// 	User      *User
// 	UserId    uint32
// 	SessId    [16]byte
// 	Ip        string
// 	UserAgent string
// 	// Pad       *pad.Pad
// 	// pc        *ClientPadContext
// }

type SessionInfo struct {
	User      *User
	StartTime time.Time
}

type ChatMessage struct {
	Id     uint32 `bson:"_id,omitempty"`
	UserId uint32
	Text   string
}

type Document struct {
	Revision uint32
	Ops      *list.List
}

type Pad struct {
	Id   uint32 `bson:"_id,omitempty"`
	Name string
}
