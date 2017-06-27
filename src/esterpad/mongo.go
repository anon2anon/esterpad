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
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	mongoLogger     = LogInit("mongo")
	MongoConnection *mgo.Session
	UserCollection  *mgo.Collection
	PadCollection   *mgo.Collection
)

type MongoChat struct {
	Id     uint32 `bson:"_id,omitempty"`
	UserId uint32
	Text   string
}

type MongoDelta struct {
	Id     uint32 `bson:"_id,omitempty"`
	UserId uint32
	Ops    []*MongoDeltaOp
}

type MongoDeltaOp struct {
	Insert interface{} `bson:",omitempty"`
	Delete *uint32     `bson:",omitempty"`
	Retain *uint32     `bson:",omitempty"`
	Meta   *PMeta      `bson:",omitempty"`
}

type MongoOpMeta struct {
	Bold      interface{} `bson:",omitempty"`
	Italic    interface{} `bson:",omitempty"`
	Underline interface{} `bson:",omitempty"`
	Strike    interface{} `bson:",omitempty"`
	FontSize  interface{} `bson:",omitempty"`
	UserId    interface{} `bson:",omitempty"`
}

type MongoPad struct {
	Id   uint32 `bson:"_id,omitempty"`
	Name string
}

type MongoUser struct {
	UserId   uint32
	Email    string `bson:",omitempty"`
	Passhash []byte `bson:",omitempty"`
	Nickname string
	Color    uint32
	Perms    uint32
}

func (meta *PMeta) GetBSON() (interface{}, error) {
	if meta == nil {
		return nil, nil
	}
	ret := MongoOpMeta{}
	if meta.Changemask&1 != 0 {
		ret.Bold = meta.Bold
	}
	if meta.Changemask&2 != 0 {
		ret.Italic = meta.Italic
	}
	if meta.Changemask&4 != 0 {
		ret.Underline = meta.Underline
	}
	if meta.Changemask&8 != 0 {
		ret.Strike = meta.Strike
	}
	if meta.Changemask&16 != 0 {
		ret.FontSize = meta.FontSize
	}
	if meta.Changemask&32 != 0 {
		if meta.User != nil {
			ret.UserId = meta.User.Id
		} else {
			ret.UserId = 0
		}
	}
	return ret, nil
}

func (meta *PMeta) SetBSON(raw bson.Raw) error {
	decoded := MongoOpMeta{}
	bsonErr := raw.Unmarshal(&decoded)
	if bsonErr != nil {
		return bsonErr
	}
	changemask := uint32(0)
	if decoded.Bold != nil {
		changemask |= 1
		meta.Bold = decoded.Bold.(bool)
	}
	if decoded.Italic != nil {
		changemask |= 2
		meta.Italic = decoded.Italic.(bool)
	}
	if decoded.Underline != nil {
		changemask |= 4
		meta.Underline = decoded.Underline.(bool)
	}
	if decoded.Strike != nil {
		changemask |= 8
		meta.Strike = decoded.Strike.(bool)
	}
	if decoded.FontSize != nil {
		changemask |= 16
		meta.FontSize = uint32(decoded.FontSize.(int))
	}
	if decoded.UserId != nil {
		changemask |= 32
		if userId := uint32(decoded.UserId.(int)); userId != 0 {
			meta.User = CacherGetUser(userId)
		}
	}
	meta.Changemask = changemask
	return nil
}

func MongoInit() {
	db, err := mgo.Dial(Config["mongodb"]["url"].(string))
	if err != nil {
		mongoLogger.Log(LOG_FATAL, "cannot dial mongo", err)
	}
	MongoConnection = db
	UserCollection = db.DB("").C("user")
	userIdIndex := mgo.Index{
		Key:    []string{"userid"},
		Unique: true,
		Sparse: true,
	}
	err = UserCollection.EnsureIndex(userIdIndex)
	if err != nil {
		mongoLogger.Log(LOG_FATAL, "mongo set scheme err", err)
	}
	userLoginIndex := mgo.Index{
		Key:    []string{"email"},
		Unique: true,
		Sparse: true,
	}
	err = UserCollection.EnsureIndex(userLoginIndex)
	if err != nil {
		mongoLogger.Log(LOG_FATAL, "mongo set scheme err", err)
	}
	PadCollection = db.DB("").C("pad")
}

func MongoLoginUser(email string, password string) interface{} {
	user := MongoUser{}
	err := UserCollection.Find(bson.M{"email": email}).Select(bson.M{"userid": 1, "passhash": 1}).One(&user)
	if err != nil {
		mongoLogger.Log(LOG_ERROR, "mongo login find err", err)
	} else {
		if user.Passhash != nil {
			if err := bcrypt.CompareHashAndPassword(user.Passhash, []byte(password)); err == nil {
				return user.UserId
			} else {
				mongoLogger.Log(LOG_ERROR, "mongo login password check err", email, err)
			}
		} else {
			mongoLogger.Log(LOG_ERROR, "mongo login no password", email, err)
		}
	}
	return nil
}

func MongoRegister(email string, password string) bool {
	passhash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err == nil {
		insert := bson.M{"email": email, "passhash": passhash}
		merr := UserCollection.Insert(insert)
		if merr == nil {
			return true
		} else {
			mongoLogger.Log(LOG_ERROR, "mongo register err", email, merr)
		}
	} else {
		mongoLogger.Log(LOG_ERROR, "mongo register hash generate err", email, err)
	}
	return false
}

func MongoRegisterFinish(user *User, email string) {
	query := bson.M{"email": email}
	change := bson.M{"$set": bson.M{"userid": user.Id, "nickname": user.Nickname,
		"color": user.Color, "perms": user.Perms}}
	err := UserCollection.Update(query, change)
	if err != nil {
		mongoLogger.Log(LOG_ERROR, "mongo register finish err", user.Id, err)
	}
}

func MongoRegisterGuest(user *User) bool {
	mongoUser := MongoUser{UserId: user.Id, Nickname: user.Nickname, Color: user.Color, Perms: user.Perms}
	err := UserCollection.Insert(mongoUser)
	if err == nil {
		return true
	} else {
		mongoLogger.Log(LOG_ERROR, "mongo register guest err", user.Id, err)
	}
	return false
}

func MongoChangeNickname(userId uint32, nickname string) {
	query := bson.M{"userid": userId}
	change := bson.M{"$set": bson.M{"nickname": nickname}}
	err := UserCollection.Update(query, change)
	if err != nil {
		mongoLogger.Log(LOG_ERROR, "mongo change nickname err", userId, err)
	}
}

func MongoChangeColor(userId uint32, color uint32) {
	query := bson.M{"userid": userId}
	change := bson.M{"$set": bson.M{"color": color}}
	err := UserCollection.Update(query, change)
	if err != nil {
		mongoLogger.Log(LOG_ERROR, "mongo change color err", userId, err)
	}
}

func MongoChangeEmail(userId uint32, email string) {
	query := bson.M{"userid": userId}
	change := bson.M{"$set": bson.M{"email": email}}
	err := UserCollection.Update(query, change)
	if err != nil {
		mongoLogger.Log(LOG_ERROR, "mongo change email err", userId, err)
	}
}
func MongoChangePassword(userId uint32, password string) {
	passhash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err == nil {
		query := bson.M{"userid": userId}
		change := bson.M{"$set": bson.M{"passhash": passhash}}
		err := UserCollection.Update(query, change)
		if err != nil {
			mongoLogger.Log(LOG_ERROR, "mongo change password err", userId, err)
		}
	} else {
		mongoLogger.Log(LOG_ERROR, "mongo change password hash generate err", userId, err)
	}

}

func MongoChangePerms(userId uint32, perms uint32) {
	query := bson.M{"userid": userId}
	change := bson.M{"$set": bson.M{"perms": perms}}
	err := UserCollection.Update(query, change)
	if err != nil {
		mongoLogger.Log(LOG_ERROR, "mongo change perms err", userId, err)
	}
}

func MongoInsertPad(id uint32, name string) {
	change := MongoPad{Id: id, Name: name}
	if err := PadCollection.Insert(change); err != nil {
		mongoLogger.Log(LOG_ERROR, "mongo insert err", err)
	}
}
