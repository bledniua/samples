package api

import (
	"fmt"
	"github.com/bledniua/rpc/jsonrpc"
	"github.com/globalsign/mgo/bson"
	jsoniter "github.com/json-iterator/go"
	"io"
	"time"
)

type Task struct {
	Id        bson.ObjectId   `json:"id" bson:"_id"`
	GroupId   bson.ObjectId   `json:"group_id"`
	WorkersId []bson.ObjectId `json:"workers_id"`
	Filter    TaskFilter      `json:"filter"`
	UserId    bson.ObjectId   `json:"user_id"`
	WorkOrder []Work          `json:"order"`
	Duration  int64           `json:"duration"`
	Limit     int             `json:"limit"`
	Skip      int             `json:"skip"`
	Status    int             `json:"status"`
	CreateAt  int64           `json:"create_at" bson:"create_at"`
	NextAt    int64           `json:"next_at" bson:"next_at"`
	UpdateAt  int64           `json:"update_at" bson:"update_at"`
}

type TaskFilter struct {
	Online int `json:"online"`
}

type Work struct {
	Method string              `json:"method"`
	Params jsoniter.RawMessage `json:"params,omitempty"`
}

type NewTask struct {
	Token     string          `json:"token" validate:"required"`
	GroupId   string          `json:"group_id" validate:"required"`
	WorkersId []bson.ObjectId `json:"workers_id"`
	Filter    TaskFilter      `json:"filter"`
	WorkOrder []Work          `json:"order"`
	Duration  int64           `json:"duration"`
	Limit     int             `json:"limit"`
	NextAt    int64           `json:"start_at"`
}

func (NewTask) Method(id float64, raw []byte, w io.Writer) error {
	result := new(NewTask)
	if err := jsoniter.Unmarshal(raw, result); err != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	if !bson.IsObjectIdHex(result.GroupId) {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("%s not object id", result.GroupId)))
	}
	gid := bson.ObjectIdHex(result.GroupId)
	//Validate token
	var uid bson.ObjectId

	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER, WORKERLIST); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Get groups list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	//validate Access
	found := false
	for _, k := range list {
		if k == gid {
			found = true
			break
		}
	}
	if !found {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("group %s access denied", gid.Hex())))
	}

	var task Task
	_ = jsoniter.Unmarshal(raw, &task)
	task.Filter = result.Filter
	task.WorkersId = result.WorkersId
	task.GroupId = gid

	task.Id = bson.NewObjectId()
	task.CreateAt = time.Now().Unix()
	task.UpdateAt = time.Now().Unix()
	task.NextAt = result.NextAt
	task.Limit = result.Limit
	task.WorkOrder = result.WorkOrder

	task.UserId = uid
	task.Skip = 0
	task.Status = 0

	//create new task
	err = db.C(workertasklist.String()).Insert(task)
	if err != nil {
		return jsonrpc.SendError(id, w, cantCreateNewTask)
	}

	return jsonrpc.Send(id, w, task)
}

type GetTaskList struct {
	Token   string `json:"token" validate:"required"`
	GroupId string `json:"group_id" validate:"required"`
}

func (GetTaskList) Method(id float64, raw []byte, w io.Writer) error {
	result := new(GetTaskList)
	if err := jsoniter.Unmarshal(raw, result); err != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	if !bson.IsObjectIdHex(result.GroupId) {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("%s not object id", result.GroupId)))
	}
	gid := bson.ObjectIdHex(result.GroupId)
	//Validate token
	var uid bson.ObjectId

	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER, WORKERLIST); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Get groups list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	//validate Access
	found := false
	for _, k := range list {
		if k == gid {
			found = true
			break
		}
	}
	if !found {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("group %s access denied", gid.Hex())))
	}

	taskList := make([]Task, 0)
	err = db.C(workertasklist.String()).Find(bson.M{"groupid": gid}).Sort("-create_at").Limit(15).All(&taskList)
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetTaskList)
	}

	return jsonrpc.Send(id, w, taskList)
}

type SetTaskStatus struct {
	Token   string `json:"token" validate:"required"`
	Id      string `json:"id" validate:"required"`
	GroupId string `json:"group_id" validate:"required"`
	Status  int    `json:"status" validate:"required"`
}

func (SetTaskStatus) Method(id float64, raw []byte, w io.Writer) error {
	result := new(SetTaskStatus)
	if err := jsoniter.Unmarshal(raw, result); err != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	if !bson.IsObjectIdHex(result.Id) {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("%s not object id", result.Id)))
	}
	tid := bson.ObjectIdHex(result.Id)
	if !bson.IsObjectIdHex(result.GroupId) {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("%s not object id", result.GroupId)))
	}
	gid := bson.ObjectIdHex(result.GroupId)

	//Validate token
	var uid bson.ObjectId

	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER, WORKERLIST); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Get groups list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	//validate Access
	found := false
	for _, k := range list {
		if k == gid {
			found = true
			break
		}
	}
	if !found {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("group %s access denied", gid.Hex())))
	}
	var task Task
	err = db.C(workertasklist.String()).Update(bson.M{"groupid": gid, "_id": tid}, bson.M{"$set": bson.M{"status": result.Status}})
	if err != nil {
		return jsonrpc.SendError(id, w, cantCreateNewTask)
	}
	err = db.C(workertasklist.String()).Find(bson.M{"groupid": gid, "_id": tid}).One(&task)
	if err != nil {
		return jsonrpc.SendError(id, w, cantCreateNewTask)
	}

	return jsonrpc.Send(id, w, task)
}
