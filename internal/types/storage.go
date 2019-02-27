package types

type Storage interface {
	LoginUser(email string, password string) (*User, error)
	Register(email string, password string) error
	RegisterFinish(user *User, email string) error
	RegisterGuest(user *User) error
	SetUserNickname(userId uint32, nickname string) error
	SetUserColor(userId uint32, color uint32) error
	SetUserEmail(userId uint32, email string) error
	SetUserPassword(userId uint32, password string) error
	SetUserPerms(userId uint32, perms uint32) error
	InsertPad(id uint32, name string) error
}
