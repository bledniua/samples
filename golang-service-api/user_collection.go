package api

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

func findUserByLogin(login string) (result *User, err error) {
	result = new(User)
	return result, db.C(user.String()).Find(bson.M{"login": login}).One(result)
}

func findUserById(id bson.ObjectId) (result *User, err error) {
	result = new(User)
	return result, db.C(user.String()).FindId(id).One(result)
}

func createNewUser(u *User) (err error) {
	u.Id = bson.NewObjectId()
	u.CreateAt = time.Now().Unix()
	u.UpdateAt = time.Now().Unix()
	return db.C(user.String()).Insert(u)
}
