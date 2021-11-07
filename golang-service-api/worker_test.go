package api

import (
	"testing"
)

func TestWorkersList(t *testing.T) {
	dbName = "miner-om"
	//defer func() {
	//err := db.DropDatabase()
	//if err != nil {
	//	panic(err)
	//}
	//}()
	err := StartUp(8001)
	if err != nil {
		panic(err)
	}
	_, err = DefaultAuth()
	if err != nil && err.Error() != "null" {
		panic(err)
	}
	_, err = CallTest(SignIn{Token: accesstoken, Login: "vtlkru", Password: "bioqcomp789"})
	if err != nil && err.Error() != "null" {
		panic(err)
	}
	_, err = CallTest(WorkersList{Token: accesstoken, Filter: WorkerListFilter{Page: 0, ByPage: 64}})
	if err != nil && err.Error() != "null" {
		panic(err)
	}
}

func TestGetTotalInfo_Method(t *testing.T) {
	dbName = "miner-om"
	//defer func() {
	//err := db.DropDatabase()
	//if err != nil {
	//	panic(err)
	//}
	//}()
	err := StartUp(8001)
	if err != nil {
		panic(err)
	}
	_, err = DefaultAuth()
	if err != nil && err.Error() != "null" {
		panic(err)
	}
	_, err = CallTest(SignIn{Token: accesstoken, Login: "vtlkru", Password: "bioqcomp789"})
	if err != nil && err.Error() != "null" {
		panic(err)
	}
	_, err = CallTest(GetTotalInfo{Token: accesstoken})
	if err != nil && err.Error() != "null" {
		panic(err)
	}
	//_, err = CallTest(GetTotalInfo{Token: "IS+ySOtmeumYnksGLsrrpYKM7YXajMVED5ncSuc2AI7csvGe6718Whl3GdxtIY5Nm7CP/ycedi1c8XzflTqf8g"})
	//if err != nil && err.Error() != "null" {
	//	panic(err)
	//}

}

func TestSetProfile_Method(t *testing.T) {
	err := StartUp(8001)
	if err != nil {
		panic(err)
	}
	_, err = DefaultAuth()
	if err != nil && err.Error() != "null" {
		panic(err)
	}
	_, err = CallTest(SignIn{Token: accesstoken, Login: "vtlkru", Password: "bioqcomp789"})
	if err != nil && err.Error() != "null" {
		panic(err)
	}
	_, err = CallTest(SetProfile{Token: accesstoken, Id: "5cd09575b2e2e417d2c2dfaf", Profile: "solo_may04"})
	if err != nil && err.Error() != "null" {
		panic(err)
	}
}
