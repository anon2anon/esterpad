package mongo

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	ep "github.com/anon2anon/esterpad/internal/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Url string
}

type Storage struct {
	connection *mgo.Session
	users      *mgo.Collection
	pads       *mgo.Collection
}

func ensureIndex(c *mgo.Collection, keys []string) {
	err := c.EnsureIndex(mgo.Index{
		Key:    keys,
		Unique: true,
		Sparse: true,
	})
	if err != nil {
		log.WithError(err).Fatal("cannot ensureIndex")
	}
}

func New(conf Config) (*Storage, error) {
	var s Storage
	log.Debug("connecting to mongo")
	db, err := mgo.Dial(conf.Url)
	if err != nil {
		return &s, err
	}
	s.connection = db
	s.users = db.DB("").C("user")
	s.pads = db.DB("").C("pad")
	ensureIndex(s.users, []string{"userid"})
	ensureIndex(s.users, []string{"email"})
	return &s, nil
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

func (s *Storage) Register(email string, password string) error {
	passhash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return errors.Wrapf(err, "hash generate err email=%v", email)
	}
	insert := bson.M{"email": email, "passhash": passhash}
	err = s.users.Insert(insert)
	if err != nil {
		return errors.Wrapf(err, "cannot insert user email=%v", email)
	}
	return nil
}

func (s *Storage) RegisterFinish(user *ep.User, email string) error {
	query := bson.M{"email": email}
	change := bson.M{
		"$set": bson.M{
			"userid":   user.Id,
			"nickname": user.Nickname,
			"color":    user.Color,
			"perms":    user.Perms,
		},
	}
	err := s.users.Update(query, change)
	if err != nil {
		return errors.Wrapf(err, "users update failed id=%v", user.Id)
	}
	return nil
}

func (s *Storage) RegisterGuest(user *ep.User) error {
	mongoUser := ep.User{Id: user.Id, Nickname: user.Nickname, Color: user.Color, Perms: user.Perms}
	err := s.users.Insert(mongoUser)
	if err != nil {
		return errors.Wrapf(err, "register guest failed id=%v", user.Id)
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

func (s *Storage) SetUserPerms(userId uint32, perms uint32) error {
	return s.setUserField(userId, "perms", perms)
}

func (s *Storage) InsertPad(id uint32, name string) error {
	change := ep.Pad{Id: id, Name: name}
	return s.pads.Insert(change)
}
