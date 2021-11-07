package api

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/bledniua/rpc/jsonrpc"
	"github.com/globalsign/mgo/bson"
	"github.com/json-iterator/go"
	"io"
	"time"
)

//time.Minute*30
var (
	AccessTime  = time.Minute * 30 + time.Second*5
	RefreshTime = time.Hour * 72
)

type Auth struct {
	Agent string `json:"agent"`
}

func (Auth) Method(id float64, raw []byte, w io.Writer) error {
	result := new(Auth)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	t := Token{
		Id:        bson.NewObjectId(),
		Agent:     result.Agent,
		AccessTo:  time.Now().Add(AccessTime).Unix(),
		RefreshAt: time.Now().Add(RefreshTime).Unix(),
		Data:      UserData{},
	}
	buff := make([]byte, 64)
	_, _ = rand.Read(buff)
	t.Refresh = base64.RawStdEncoding.EncodeToString(buff)
	_, _ = rand.Read(buff)
	t.Access = base64.RawStdEncoding.EncodeToString(buff)
	TokenCache[t.Access] = t
	if err := db.C(token.String()).Insert(t); err != nil {
		return jsonrpc.SendError(id, w, cantCreateNewToken(err))
	}

	return jsonrpc.Send(id, w, t.Access, t.AccessTo-5, t.Refresh, t.RefreshAt)
}

type Refresh struct {
	Token string `json:"token"`
}

func (Refresh) Method(id float64, raw []byte, w io.Writer) error {
	// get message
	result := new(Refresh)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	// get old token
	oldToken := new(Token)
	if err := db.C(token.String()).Find(bson.M{"refresh": result.Token}).One(&oldToken); err != nil {
		return jsonrpc.SendError(id, w, cantFindToken(err))
	}
	// remove old token
	if err := db.C(token.String()).RemoveId(oldToken.Id); err != nil {
		return jsonrpc.SendError(id, w, cantRefreshToken(err))
	}
	delete(TokenCache, oldToken.Access)

	// create new token
	t := Token{
		Id:        bson.NewObjectId(),
		AccessTo:  time.Now().Add(AccessTime).Unix(),
		RefreshAt: time.Now().Add(RefreshTime).Unix(),
	}
	// if refresh time not end set data from old token
	if oldToken.RefreshAt > time.Now().Unix() {
		t.Data = oldToken.Data
	}
	buff := make([]byte, 64)
	_, _ = rand.Read(buff)
	t.Refresh = base64.RawStdEncoding.EncodeToString(buff)
	_, _ = rand.Read(buff)
	t.Access = base64.RawStdEncoding.EncodeToString(buff)
	// caching token
	TokenCache[t.Access] = t
	if err := db.C(token.String()).Insert(t); err != nil {
		return jsonrpc.SendError(id, w, cantCreateNewToken(err))
	}

	return jsonrpc.Send(id, w, t.Access, t.AccessTo-5, t.Refresh, t.RefreshAt)
}
