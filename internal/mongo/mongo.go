package mongo

import (
	"fmt"
	"strconv"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	ep "github.com/anon2anon/esterpad/internal/types"
	"github.com/pkg/errors"
)

type Config struct {
	Url string
}

type Storage struct {
	conn  *mgo.Session
	db    *mgo.Database
	users *mgo.Collection
	pads  *mgo.Collection
}

func ensureIndex(c *mgo.Collection, keys []string) error {
	err := c.EnsureIndex(mgo.Index{
		Key:    keys,
		Unique: true,
		Sparse: true,
	})
	if err != nil {
		return errors.Wrap(err, "cannot ensureIndex")
	}
	return nil
}

func New(conf Config) (*Storage, error) {
	var s Storage
	conn, err := mgo.Dial(conf.Url)
	if err != nil {
		return &s, err
	}

	s.conn = conn
	s.db = conn.DB("")
	s.users = s.db.C("user")
	s.pads = s.db.C("pad")

	err = ensureIndex(s.users, []string{"userid"})
	if err != nil {
		return &s, err
	}
	err = ensureIndex(s.users, []string{"email"})
	if err != nil {
		return &s, err
	}

	return &s, nil
}

func (s *Storage) LoadUsers() (*map[uint32]*ep.User, error) {
	res := make(map[uint32]*ep.User)
	userIter := UserCollection.Find(nil).Iter()
	user := ep.User{}
	for userIter.Next(&user) {
		res[user.UserId] = &ep.User{user}
	}
	if err := userIter.Close(); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Storage) LoadPads() (map[uint32]*ep.Pad, error) {
	res := make(map[uint32]*ep.Pad)
	userIter := PadCollection.Find(nil).Iter()
	user := mongo.Pad{}
	for userIter.Next(&user) {
		// TODO
	}
	if err := userIter.Close(); err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Storage) LoginUser(email string, password string) (*ep.User, error) {
	user := ep.User{}
	err := s.users.Find(bson.M{"email": email}).Select(bson.M{"userid": 1, "passhash": 1}).One(&user)
	if err != nil {
		return nil, errors.Wrap(err, "login find failed")
	}
	if user.Passhash == nil {
		return nil, fmt.Errorf("empty passhash")
	}
	if err := bcrypt.CompareHashAndPassword(user.Passhash, []byte(password)); err != nil {
		return nil, fmt.Errorf("wrong pass for %v", email)
	}
	return &user, nil
}

func (s *Storage) RegisterUser(user ep.User, password string) error {
	passhash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return errors.Wrapf(err, "hash generate err email=%v", user.Email)
	}
	insert := bson.M{"email": email, "passhash": passhash}
	err = s.users.Upsert(
		bson.M{"email": user.Email},
		bson.M{
			"$set": bson.M{
				"userid":   user.Id,
				"nickname": user.Nickname,
				"color":    user.Color,
				"perms":    user.Perms,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "cannot upsert user, email=%v", user.Email)
	}
	return nil
}

func (s *Storage) setUserField(userId uint32, field string, data interface{}) error {
	query := bson.M{"userid": userId}
	change := bson.M{"$set": bson.M{field: data}}
	return s.users.Update(query, change)
}

func (s *Storage) SetUserNickname(userId uint32, nickname string) error {
	return s.setUserField(userId, "nickname", nickname)
}

func (s *Storage) SetUserColor(userId uint32, color uint32) error {
	return s.setUserField(userId, "color", color)
}

func (s *Storage) SetUserEmail(userId uint32, email string) error {
	return s.setUserField(userId, "email", email)
}

func (s *Storage) SetUserPassword(userId uint32, password string) error {
	passhash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return errors.Wrap(err, "generate hash failed")
	}
	return s.setUserField(userId, "passhash", passhash)
}

func (s *Storage) SetUserPerms(userId uint32, perms ep.UserPerms) error {
	return s.setUserField(userId, "perms", perms)
}

func (s *Storage) InsertPad(id uint32, name string) error {
	change := ep.Pad{Id: id, Name: name}
	return s.pads.Insert(change)
}

func (s *Storage) getPadCollection(name string, padId uint32) *mgo.Collection {
	// TODO: cache it?
	return s.db.C(name + strconv.FormatInt(int64(padId), 10))
}
