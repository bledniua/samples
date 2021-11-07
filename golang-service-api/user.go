package api

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"strings"
)

type User struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Login    string        `bson:"login"`
	Password []byte        `json:"-" bson:"password"`
	Email    string        `bson:"email"`
	Groups   []Group       `bson:"groups" json:"-"`
	CreateAt int64         `bson:"create_at"`
	UpdateAt int64         `bson:"update_at"`
}

type Key struct {
	Id     bson.ObjectId `bson:"_id" json:"-"`
	Data   string        `bson:"data" json:"-"`
	Groups []Group       `bson:"groups" json:"-"`
}

func GetNewKey(key Key) (string, error) {
	r := make([]byte, 29)
	_, _ = rand.Read(r)
	key.Id = bson.NewObjectId()
	key.Data = base64.RawStdEncoding.EncodeToString(r)
	return fmt.Sprintf("%s.%s", key.Id.Hex(), key.Data), db.C(keys.String()).Insert(key)
}

func GetKeyAndRemove(key string) (*Key, error) {
	spl := strings.Split(key, ".")
	if len(spl) < 2 {
		return nil, errors.New("invalid key")
	}
	if !bson.IsObjectIdHex(spl[0]) {
		return nil, errors.New("invalid key")
	}
	check := new(Key)
	check.Id = bson.ObjectIdHex(spl[0])
	if db.C(keys.String()).FindId(check.Id).One(check) != nil {
		return nil, errors.New("cant validate key")
	}
	if check.Data != spl[1] {
		return nil, errors.New("invalid key")
	}
	if db.C(keys.String()).RemoveId(check.Id) != nil {
		return nil, errors.New("cant validate key")
	}
	return check, nil
}
