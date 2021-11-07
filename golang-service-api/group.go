package api

import "github.com/globalsign/mgo/bson"

type RigGroup struct {
	Id       bson.ObjectId   `json:"id" bson:"_id"`
	Name     string          `json:"name"`
	Address  string          `json:"address"`
	CoinBase string          `json:"coin_base"`
	Key      string          `json:"key"`
	Settings GroupSettings   `json:"settings"`
	Owner    bson.ObjectId   `json:"-"`
	SubUsers []bson.ObjectId `json:"sub_users"`
	CreateAt int64           `json:"create_at"`
	UpdateAt int64           `json:"update_at"`
}

type GroupSettings struct {
	AutoProfile string            `json:"auto_profile"`
	OcTable     map[string]string `json:"oc_table"`
	Modules     []Module          `json:"modules"`
}
