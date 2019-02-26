package types

import (
	"container/list"
	"time"
	// "github.com/anon2anon/esterpad/internal/pad"
)

type User struct {
	Id       uint32 `bson:"_id,omitempty"`
	Email    string `bson:",omitempty"`
	Passhash []byte `bson:",omitempty"`
	Nickname string
	Color    uint32
	Perms    uint32
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

// type Pad struct {
// 	Id              uint32
// 	Name            string
// 	CacherChannel   chan interface{}
// 	Clients         *list.List
// 	ClientsMutex    sync.RWMutex
// 	ChatCounter     uint32
// 	ChatArray       []*PChat
// 	ChatMutex       sync.RWMutex
// 	DeltaArray      []*PDelta
// 	DocumentArray   []*PDocument
// 	DeltaCounter    uint32
// 	DeltaMutex      sync.RWMutex
// 	ChatCollection  *mgo.Collection
// 	DeltaCollection *mgo.Collection
// }

// type DeltaOp struct {
// 	Insert interface{} `bson:",omitempty"`
// 	Delete *uint32     `bson:",omitempty"`
// 	Retain *uint32     `bson:",omitempty"`
// 	Meta   *pad.Meta   `bson:",omitempty"`
// }
