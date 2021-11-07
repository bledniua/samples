package api

import (
	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
	"time"
)

type Group int

const (
	NORULE Group = iota
	BANNED
	ADMIN
	BASICUSER
	WORKERLIST
	SUBWORKERS
	PAYMENTS
)

type Token struct {
	Id        bson.ObjectId `bson:"_id"`
	Agent     string        `json:"agent,omitempty"`
	Refresh   string        `bson:"refresh"`
	Access    string        `bson:"access"`
	Data      UserData      `bson:"data"`
	RefreshAt int64         `bson:"refresh_at"`
	AccessTo  int64         `bson:"access_to"`
}

type TokenCacheItem struct {
	Access   string   `bson:"access"`
	AccessTo int64    `bson:"access_to"`
	Data     UserData `bson:"data"`
}

type UserData struct {
	UserId bson.ObjectId `bson:"user_id,omitempty"`
	Groups []Group       `bson:"groups,omitempty"`
}

var TokenCache = map[string]Token{}

func contains(s []Group, e Group) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ValidateToken(access string, groups ...Group) error {
	if t, ok := TokenCache[access]; ok && t.AccessTo > time.Now().Unix() {
		for _, group := range groups {
			if !contains(t.Data.Groups, group) {
				return errors.New("invalid rule")
			}
		}
		return nil
	}
	return errors.New("token not exit")
}

func ValidateAndGetUserId(access string, groups ...Group) (bson.ObjectId, error) {
	if t, ok := TokenCache[access]; ok {
		if t.AccessTo < time.Now().Unix() {
			return bson.ObjectId(""), errors.New("token time out")
		}
		for _, group := range groups {
			if !contains(t.Data.Groups, group) {
				return bson.ObjectId(""), errors.New("invalid rule")
			}
		}
		return t.Data.UserId, nil
	}
	return bson.ObjectId(""), errors.New("token not exit")
}

func AddUserToken(access string, user *User) error {
	t, ok := TokenCache[access]
	if !ok {
		return errors.New("not found")
	}
	t.Data.UserId = user.Id
	t.Data.Groups = user.Groups
	TokenCache[access] = t
	return db.C(token.String()).UpdateId(t.Id, t)
}

func RemoveUserToken(access string) (data *UserData, err error) {
	t, ok := TokenCache[access]
	if !ok {
		return nil, errors.New("not found")
	}
	data = new(UserData)
	*data = t.Data
	t.Data = UserData{}
	TokenCache[access] = t
	return data, db.C(token.String()).UpdateId(t.Id, t)
}

func RefreshUserToken(access string, user *User) error {
	t, ok := TokenCache[access]
	if !ok {
		return errors.New("not found")
	}
	t.Data.UserId = user.Id
	t.Data.Groups = user.Groups
	TokenCache[access] = t
	return db.C(token.String()).UpdateId(t.Id, t)
}
