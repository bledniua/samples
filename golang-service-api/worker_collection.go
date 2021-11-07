package api

import "github.com/globalsign/mgo/bson"

func FindWorkerByLogin(login string) (result *User, err error) {
	result = new(User)
	return result, db.C(user.String()).Find(bson.M{"login": login}).One(result)
}
