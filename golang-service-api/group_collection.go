package api

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

func newRigGroup(w *RigGroup) error {
	w.Id = bson.NewObjectId()
	w.CreateAt = time.Now().Unix()
	w.UpdateAt = time.Now().Unix()
	w.Name = w.Id.Hex()[20:24]

	return db.C(group.String()).Insert(w)
}

type GroupList struct {
	Ids []bson.ObjectId `json:"ids"`
}

func GetGroupsIdList(owner bson.ObjectId) ([]bson.ObjectId, error) {
	result := GroupList{}
	q := db.C(group.String()).Pipe([]bson.M{
		{"$match": bson.M{"owner": owner}},
		{"$group": bson.M{"_id": owner, "ids": bson.M{"$push": "$_id"}}},
	})
	return result.Ids, q.One(&result)
}

func GetGroupsList(owner bson.ObjectId, list *[]RigGroup) error {
	return db.C(group.String()).Find(bson.M{"owner": owner}).All(list)
}
