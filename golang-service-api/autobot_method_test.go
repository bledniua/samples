package api

import (
	"testing"
)

func TestGetTaskList_Method(t *testing.T) {
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
	_, err = CallTest(GetTaskList{Token: accesstoken, GroupId: "5c71bb63b2e2e47206720453"})
	if err != nil && err.Error() != "null" {
		panic(err)
	}

}

func TestNewTask_Method(t *testing.T) {
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
	_, err = CallTest(NewTask{Token: accesstoken, GroupId: "5c71bb63b2e2e47206720453", Limit: 1, Filter: TaskFilter{Online: 0}, Duration: 60, WorkOrder: []Work{{Method: "sreboot"}, {Method: "hreboot"}}})
	if err != nil && err.Error() != "null" {
		panic(err)
	}
}

func TestSetTaskStatus_Method(t *testing.T) {
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
	_, err = CallTest(SetTaskStatus{Token: accesstoken, GroupId: "5c71bb63b2e2e47206720453", Status: 0, Id: "5cfe86afb2e2e4129f1c8510"})
	if err != nil && err.Error() != "null" {
		panic(err)
	}
}
