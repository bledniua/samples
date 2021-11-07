package api

import (
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"testing"
)

func TestCreateNewWallet(t *testing.T) {
	//dbName = "miner-om-test"
	//defer func() {
	//err := db.DropDatabase()
	//if err != nil {
	//	panic(err)
	//}
	//}()
	err := StartUp(8002)
	if err != nil {
		panic(err)
	}
	_, err = DefaultAuth()
	if err != nil && err.Error() != "null" {
		panic(err)
	}

	key, err := GetNewKey(Key{Groups: []Group{ADMIN, BASICUSER}})
	if err != nil {
		_, _ = GetKeyAndRemove(key)
		panic(err)
	}
	_, err = CallTest(SignIn{Token: accesstoken, Login: login, Password: "bioqcomp789"})
	if err != nil && err.Error() != "null" {
		panic(err)
	}
	_, err = CallTest(CreateNewGroup{Token: accesstoken, CoinBase: "", Address: ""})
	if err != nil && err.Error() != "null" {
		panic(err)
	}
}

func TestGetAccountList(t *testing.T) {
	dbName = "miner-om-test"
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	db = session.DB(dbName)

	list, err := GetGroupsIdList(bson.ObjectIdHex("5c711d3bb2e2e46f1c4f33e2"))
	if err != nil {
		panic(err)
	}
	fmt.Println(list)
}
