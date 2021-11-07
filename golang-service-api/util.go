package api

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"gitlab.com/toby3d/telegraph"
	"time"
)

func GetSecureToken() string {
	t := time.Now().Unix()
	t = t - t%10
	h := sha256.Sum256([]byte(fmt.Sprintf("%dPSANWGU8kx##Xc+6%d", t, t)))
	h2 := sha256.Sum256([]byte(fmt.Sprintf("%dA2$ppLTyDVKDtv?D%d", t%1000, t%2000)))
	return base64.RawStdEncoding.EncodeToString(h[:]) + base64.RawStdEncoding.EncodeToString(h2[:])
}

type AccountHashSnap struct {
	GroupId bson.ObjectId `json:"account_id" bson:"_id"`
	Count   int           `json:"count"`
	Khs     float64       `json:"khs" bson:"khs"`
	Time    int           `json:"time"`
}

type GetHashItem struct {
	Id     bson.ObjectId   `json:"id"`
	Login  string          `json:"login"`
	Groups []HashItemGroup `json:"groups"`
}

type HashItemGroup struct {
	Id       bson.ObjectId `json:"id"`
	Key      string        `json:"key"`
	Address  string        `json:"address"`
	CoinBase string        `json:"coin_base"`
	Avg      float64       `json:"avg"`
}

type Post struct {
	Id      bson.ObjectId    `json:"id" bson:"_id"`
	Title   string           `json:"title"`
	Content []telegraph.Node `json:"content"`
	CreatAt int64            `json:"creat_at"`
}
