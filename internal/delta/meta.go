package delta

import "gopkg.in/mgo.v2/bson"

type OpMeta struct {
	Changemask uint32 `bson:""`
	Bold       bool
	Italic     bool
	Underline  bool
	Strike     bool
	FontSize   uint32
	UserId     uint32
}

func (meta *OpMeta) GetBSON() (interface{}, error) {
	if meta == nil {
		return nil, nil
	}
	ret := OpMeta{}
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

func (meta *OpMeta) SetBSON(raw bson.Raw) error {
	decoded := OpMeta{}
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
