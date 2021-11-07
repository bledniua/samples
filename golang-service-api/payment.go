package api

import "github.com/globalsign/mgo/bson"

type Payment struct {
	Id       bson.ObjectId `json:"id" bson:"_id"`
	GroupId  bson.ObjectId `json:"group_id"`
	Amount   float64       `json:"amount"`
	Wallet   string        `json:"wallet"`
	Link     string        `json:"link"`
	Status   int           `json:"status"`
	CreateAt int64         `json:"create_at"`
}
