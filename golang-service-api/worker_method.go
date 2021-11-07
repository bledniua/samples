package api

import (
	"fmt"
	"github.com/bledniua/rpc/jsonrpc"
	"github.com/globalsign/mgo/bson"
	"github.com/json-iterator/go"
	"io"
	"math"
	"sort"
	"strings"
	"time"
)

type WorkerListFilter struct {
	Page         int    `json:"page"  validate:"min=0"`
	ByPage       int    `json:"by_page" validate:"required,min=8,max=512"`
	UpdateAfter  int    `json:"update_after,omitempty"`
	UpdateBefore int    `json:"update_before,omitempty"`
	GroupId      string `json:"group_id,omitempty"`
}

type WorkerGroup struct {
	Id      bson.ObjectId `json:"group" bson:"_id"`
	Workers []BasicWorker `bson:"workers" json:"workers"`
}

type AdminGroup struct {
	Id      bson.ObjectId `json:"group" bson:"_id"`
	Workers []AdminWorker `bson:"workers" json:"workers"`
}

type Release struct {
	Sysname    string `json:"sysname"`
	Nodename   string `json:"nodename"`
	Release    string `json:"release"`
	Version    string `json:"version"`
	Machine    string `json:"machine"`
	Domainname string `json:"domainname"`
}

func (r Release) MarshalJSON() ([]byte, error) {
	return []byte("\"" + strings.Join([]string{r.Sysname, r.Nodename, r.Release, r.Version, r.Version, r.Machine, r.Domainname}, " ") + "\""), nil
}

type BasicWorker struct {
	Id           bson.ObjectId                 `json:"id" bson:"_id"`
	Version      int                           `json:"version"`
	Release      Release                       `json:"build"`
	Name         string                        `json:"name"`
	Profile      string                        `json:"profile"`
	Mac          string                        `json:"mac"`
	LocalIp      string                        `json:"local_ip"`
	Static       GpuModuleMemory               `json:"static,omitempty"`
	List         interface{}                   `json:"list"`
	Current      interface{}                   `json:"current"`
	Temp         map[string]BasicGpuStatusInfo `json:"temp,omitempty"`
	Khs          float64                       `json:"khs"`
	Efi          []int                         `json:"efi"`
	SIdx         int                           `json:"status" bson:"sidx"`
	Ping         time.Duration                 `json:"ping"`
	StartAt      int64                         `json:"start_at" bson:"start_at"`
	UpdateAt     int64                         `json:"update_at" bson:"update_at"`
	HashUpdateAt int64                         `json:"hash_update_at" bson:"hash_update_at"`
}

type BasicGpuStatusInfo struct {
	Temperature int `json:"t" bson:"t"`
	FanSpeed    int `json:"fs" bson:"fs"`
}

type BasicFullWorker struct {
	Id           bson.ObjectId          `json:"id" bson:"_id"`
	Version      int                    `json:"version"`
	Release      Release                `json:"build"`
	Name         string                 `json:"name"`
	Profile      string                 `json:"profile"`
	Modules      []Module               `json:"modules"`
	Mac          string                 `json:"mac"`
	LocalIp      string                 `json:"local_ip"`
	Status       GpuModule              `json:"status"`
	Static       GpuModule              `json:"static"`
	Overclock    Overclock              `json:"overclock" bson:"overclock"`
	List         interface{}            `json:"list"`
	Current      interface{}            `json:"current"`
	Khs          float64                `json:"khs"`
	Dkhs         []float64              `json:"dkhs"`
	Efi          []int                  `json:"efi"`
	Ping         time.Duration          `json:"ping"`
	Chart        map[int64]float64      `json:"chart,omitempty"`
	Config       map[string]interface{} `json:"config" bson:"config"`
	StartAt      int64                  `json:"start_at" bson:"start_at"`
	UpdateAt     int64                  `json:"update_at" bson:"update_at"`
	HashUpdateAt int64                  `json:"hash_update_at" bson:"hash_update_at"`
	GroupId      bson.ObjectId          `json:"group_id" bson:"groupid"`
	MirrorAddr   string                 `json:"mirror_addr,omitempty" bson:"mirroraddr,omitempty"`
}

type Overclock struct {
	Amd     map[string]SetAmdGpuStaticInfo `json:"amd"`
	Disable bool                           `json:"disable" bson:"disable"`
}

type Module struct {
	Name  string        `json:"name"`
	Value []interface{} `json:"value"`
}

type AdminWorker struct {
	Id       bson.ObjectId `json:"id" bson:"_id"`
	Version  int           `json:"version"`
	Release  Release       `json:"build"`
	Name     string        `json:"name"`
	Profile  string        `json:"profile"`
	Modules  []Module      `json:"modules"`
	Mac      string        `json:"mac"`
	LocalIp  string        `json:"local_ip"`
	List     interface{}   `json:"list"`
	Current  interface{}   `json:"current"`
	Khs      float64       `json:"khs"`
	Ping     time.Duration `json:"ping"`
	StartAt  int64         `json:"start_at" bson:"start_at"`
	UpdateAt int64         `json:"update_at" bson:"update_at"`

	Messages jsoniter.RawMessage `json:"messages,omitempty"`
}

type GpuModule struct {
	AmdModule string              `json:"amd_module,omitempty"`
	Amd       jsoniter.RawMessage `json:"amd,omitempty"`
	NvModule  string              `json:"nv_module,omitempty"`
	Nv        jsoniter.RawMessage `json:"nv,omitempty"`
}

type GpuModuleMemory struct {
	AmdModule string              `json:"-"`
	Amd       jsoniter.RawMessage `json:"amd,omitempty" bson:"amd"`
	//NvModule  string              `json:"nv_module,omitempty"`
	//Nv        jsoniter.RawMessage `json:"nv,omitempty"`
}

type MemoryModelRecord struct {
	MemoryModel string `json:"memory_model"`
}

func (_this *GpuModuleMemory) MarshalJSON() ([]byte, error) {
	if _this.AmdModule != "" {
		iter := jsoniter.ParseBytes(jsoniter.ConfigFastest, _this.Amd)
		records := map[string]MemoryModelRecord{}
		iter.ReadVal(&records)
		if iter.Error != nil {
			return jsoniter.Marshal(nil)
		}
		raw, err := jsoniter.Marshal(records)
		if err != nil {
			return jsoniter.Marshal(nil)
		}

		return jsoniter.Marshal(GpuModuleMemory{Amd: raw, AmdModule: _this.AmdModule})
	}
	return jsoniter.Marshal(nil)
}

type WorkersList struct {
	Token  string           `json:"token" validate:"required"`
	Filter WorkerListFilter `json:"filter"`
	Sort   map[string]int8  `json:"sort"`
}

func (WorkersList) Method(id float64, raw []byte, w io.Writer) error {
	result := new(WorkersList)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Validate token
	var uid bson.ObjectId

	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER, WORKERLIST); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Get wallets list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	var data = make([]WorkerGroup, 0)

	match := bson.M{}
	//check is it user group
	if bson.IsObjectIdHex(result.Filter.GroupId) {
		for _, gr := range list {
			if gr.Hex() == result.Filter.GroupId {
				match["groupid"] = bson.ObjectIdHex(result.Filter.GroupId)
				break
			}
		}
	} else {
		match["groupid"] = bson.M{"$in": list}
	}

	if result.Filter.UpdateAfter > 0 {
		match["update_at"] = bson.M{"$gt": result.Filter.UpdateAfter}
	}
	if result.Filter.UpdateBefore > 0 {
		match["update_at"] = bson.M{"$lt": result.Filter.UpdateBefore}
	}

	sort_by := bson.M{}
	for key, value := range result.Sort {
		sort_by[key] = value
	}

	if len(sort_by) == 0 {
		sort_by["create_at"] = 1
	}
	err = db.C(worker.String()).Pipe([]bson.M{
		{"$match": match},
		{"$sort": sort_by},
		{"$skip": result.Filter.Page * result.Filter.ByPage},
		{"$limit": result.Filter.ByPage},
		{"$group": bson.M{"_id": "$groupid", "workers": bson.M{"$push": "$$ROOT"}}},
	}).All(&data)

	if err != nil {
		fmt.Println(err)
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	total, err := db.C(worker.String()).Find(match).Count()
	return jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: []interface{}{result.Filter.Page, total, data}})
}

type Worker struct {
	Token string `json:"token" validate:"required"`
	Id    string `json:"id"`
}

type WorkerHash struct {
	Time  int64   `json:"time"`
	Count int64   `json:"count"`
	Khs   float64 `json:"khs"`
}

func (Worker) Method(id float64, raw []byte, w io.Writer) error {
	result := new(Worker)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	if !bson.IsObjectIdHex(result.Id) {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("id %s is not object id", result.Id)))
	}
	//Validate token
	var uid bson.ObjectId

	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER, WORKERLIST); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Get wallets list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	data := BasicFullWorker{}
	err = db.C(worker.String()).Find(bson.M{"_id": bson.ObjectIdHex(result.Id), "groupid": bson.M{"$in": list}}).One(&data)
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}
	wHash := new(WorkerHash)
	iter := db.C(workerhash.String()).Find(bson.M{"wid": data.Id}).Iter()
	if !iter.Done() {
		data.Chart = map[int64]float64{}
	}
	t := time.Now().Unix()
	current := t - t%600
	for iter.Next(wHash) {
		if wHash.Time == current {
			data.Chart[wHash.Time] = wHash.Khs / math.Max(float64(wHash.Count), 1)
		} else {
			data.Chart[wHash.Time] = wHash.Khs / 60
		}
	}
	if iter.Err() != nil {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}

	return jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: data})
}

type DeleteWorker struct {
	Token string `json:"token" validate:"required"`
	Id    string `json:"id"`
}

func (DeleteWorker) Method(id float64, raw []byte, w io.Writer) error {
	result := new(DeleteWorker)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	if !bson.IsObjectIdHex(result.Id) {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("id %s is not object id", result.Id)))
	}
	//Validate token
	var uid bson.ObjectId

	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER, WORKERLIST); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Get wallets list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	err = db.C(worker.String()).Remove(bson.M{"_id": bson.ObjectIdHex(result.Id), "groupid": bson.M{"$in": list}})
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}

	return jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: true})
}

type DeleteWorkers struct {
	Token string   `json:"token" validate:"required"`
	Ids   []string `json:"ids"`
}

func (DeleteWorkers) Method(id float64, raw []byte, w io.Writer) error {
	result := new(DeleteWorkers)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Validate token
	var uid bson.ObjectId

	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER, WORKERLIST); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Get wallets list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	arr := make([]bson.ObjectId, len(result.Ids))
	for idx, str := range result.Ids {
		arr[idx] = bson.ObjectIdHex(str)
	}

	_, err = db.C(worker.String()).RemoveAll(bson.M{"_id": bson.M{"$in": arr}, "groupid": bson.M{"$in": list}})
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}

	return jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: true})
}

type Call struct {
	Token string              `json:"token" validate:"required"`
	Id    string              `json:"id"`
	M     string              `json:"method" validate:"required"`
	Data  jsoniter.RawMessage `json:"data,omitempty"`
}

type SetAmdGpuStaticInfo struct {
	CoreClock   []int `json:"core_clock"`
	CoreVDDC    []int `json:"core_vddc"`
	MemoryClock []int `json:"memory_clock"`
	MemoryVDDC  []int `json:"memory_vddc"`
}

func (Call) Method(id float64, raw []byte, w io.Writer) error {
	//fmt.Println("call", string(raw))
	result := new(Call)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Validate token
	var uid bson.ObjectId

	if uid, err = ValidateAndGetUserId(result.Token, GetGroupForMethod(result.M), WORKERLIST); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}

	//Get wallets list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	if !bson.IsObjectIdHex(result.Id) {
		return jsonrpc.SendError(id, w, cantGetWorkerID)
	}

	// validate worker exist
	var n int
	n, err = db.C(worker.String()).Find(bson.M{"_id": bson.ObjectIdHex(result.Id), "groupid": bson.M{"$in": list}}).Count()
	if err != nil || n != 1 {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}

	switch result.M {
	case "Amdcall":
		table := map[string]SetAmdGpuStaticInfo{}
		iter := jsoniter.ParseBytes(jsoniter.ConfigFastest, result.Data)
		iter.ReadArray()
		iter.Skip()
		iter.ReadArray()
		iter.ReadVal(&table)

		err = db.C(worker.String()).Update(bson.M{"_id": bson.ObjectIdHex(result.Id)}, bson.M{"$set": bson.M{"overclock.amd": table}})
		if err != nil {
			return jsonrpc.SendError(id, w, cantGetWorker)
		}
		return jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: "amd overclock apply"})

	case "AmdDisable":
		err = db.C(worker.String()).Update(bson.M{"_id": bson.ObjectIdHex(result.Id)}, bson.M{"$set": bson.M{"overclock.disable": true}})
		if err != nil {
			return jsonrpc.SendError(id, w, cantGetWorker)
		}
		return jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: "amd overclock disabled"})

	case "AmdEnable":
		err = db.C(worker.String()).Update(bson.M{"_id": bson.ObjectIdHex(result.Id)}, bson.M{"$set": bson.M{"overclock.disable": false}})
		if err != nil {
			return jsonrpc.SendError(id, w, cantGetWorker)
		}
		return jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: "amd overclock enabled"})
	case "Setname":
		iter := jsoniter.ParseBytes(jsoniter.ConfigFastest, result.Data)
		iter.ReadArray()
		name := iter.ReadString()
		if iter.Error == nil && name != "" {
			err = db.C(worker.String()).Update(bson.M{"_id": bson.ObjectIdHex(result.Id)}, bson.M{"$set": bson.M{"name": name}})
			if err != nil {
				return jsonrpc.SendError(id, w, cantGetWorker)
			}
			//return jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: "name changed"})
		}
		//return jsonrpc.SendError(id, w, cantGetWorkerName)
		break
	case "Setconfig":
		iter := jsoniter.ParseBytes(jsoniter.ConfigFastest, result.Data)
		iter.ReadArray()
		name := iter.ReadString()
		//var val interface{}
		//iter.ReadVal(val)
		iter.ReadArray()
		val := iter.ReadAny().GetInterface()
		err = db.C(worker.String()).Update(bson.M{"_id": bson.ObjectIdHex(result.Id)}, bson.M{"$set": bson.M{"config": bson.M{name: val}}})
		if err != nil {
			return jsonrpc.SendError(id, w, cantGetWorker)
		}
		break
	}

	// add in queue
	err = dispatch(result.Id, result.M, result.Data)
	if err != nil {
		return jsonrpc.SendError(id, w, cantSendRequest)
	}
	CallList.Insert(CallItem{Id: bson.ObjectIdHex(result.Id), Method: result.M, User: uid, Time: time.Now().Unix()})
	_ = jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: "in queue"})

	return nil
}

type SetProfile struct {
	Token   string `json:"token" validate:"required"`
	Id      string `json:"id"`
	Profile string `json:"profile"`
}

func (SetProfile) Method(id float64, raw []byte, w io.Writer) error {
	result := new(SetProfile)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Validate token
	var uid bson.ObjectId

	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER, WORKERLIST); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}

	//Get wallets list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	// update worker
	err = db.C(worker.String()).Update(bson.M{"_id": bson.ObjectIdHex(result.Id), "groupid": bson.M{"$in": list}}, bson.M{"$set": bson.M{"profile": result.Profile}})
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}

	data := BasicFullWorker{}
	err = db.C(worker.String()).Find(bson.M{"_id": bson.ObjectIdHex(result.Id), "groupid": bson.M{"$in": list}}).One(&data)
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}

	err = notify("onstartup", map[string]interface{}{"w": data.Id.Hex(), "version": data.Version})
	if err != nil {
		return jsonrpc.SendError(id, w, cantSendRequest)
	}
	return jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: data})
}

type SetMirrorAddr struct {
	Token string `json:"token" validate:"required"`
	Id    string `json:"id"`
	Addr  string `json:"addr"`
}

func (SetMirrorAddr) Method(id float64, raw []byte, w io.Writer) error {
	result := new(SetMirrorAddr)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Validate token
	var uid bson.ObjectId

	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER, WORKERLIST); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}

	//Get wallets list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	// update worker
	err = db.C(worker.String()).Update(bson.M{"_id": bson.ObjectIdHex(result.Id), "groupid": bson.M{"$in": list}}, bson.M{"$set": bson.M{"mirroraddr": result.Addr}})
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}

	data := BasicFullWorker{}
	err = db.C(worker.String()).Find(bson.M{"_id": bson.ObjectIdHex(result.Id), "groupid": bson.M{"$in": list}}).One(&data)
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}

	return jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: data})
}

type Enable struct {
	Token string        `json:"token" validate:"required"`
	Id    string        `json:"id"`
	Name  string        `json:"name"`
	Value []interface{} `json:"value"`
}

func (Enable) Method(id float64, raw []byte, w io.Writer) error {
	result := new(Enable)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Validate token
	var uid bson.ObjectId

	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER, WORKERLIST); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}

	//Get wallets list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}
	// validate worker exist
	var wrk BasicFullWorker
	if !bson.IsObjectIdHex(result.Id) {
		//fmt.Println(result.Id)
		return jsonrpc.InvalidP(id, w)
	}
	err = db.C(worker.String()).Find(bson.M{"_id": bson.ObjectIdHex(result.Id), "groupid": bson.M{"$in": list}}).One(&wrk)
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}
	found := false
	for id, m := range wrk.Modules {
		if m.Name == result.Name {
			if len(result.Value) > 0 && result.Value[0] == "" {
				wrk.Modules = append(wrk.Modules[:id], wrk.Modules[id+1:]...)
				found = true
				break
			}
			if len(result.Value) == 0 {
				wrk.Modules = append(wrk.Modules[:id], wrk.Modules[id+1:]...)
				found = true
				break
			}
			wrk.Modules[id].Value = result.Value
			found = true
			break
		}
	}
	if !found {
		wrk.Modules = append(wrk.Modules, Module{Name: result.Name, Value: result.Value})
	}

	err = db.C(worker.String()).Update(bson.M{"_id": bson.ObjectIdHex(result.Id)}, bson.M{"$set": bson.M{"modules": wrk.Modules}})
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}
	_ = notify("refreshmodule", map[string]interface{}{"id": wrk.Id.Hex(), "name": result.Name, "value": result.Value})
	CallList.Insert(CallItem{Id: bson.ObjectIdHex(result.Id), Method: result.Name, User: uid, Time: time.Now().Unix()})

	_ = jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: map[string]interface{}{"module": result.Name, "value": result.Value}})
	return nil
}

type Do struct {
	Token string        `json:"token" validate:"required"`
	Id    string        `json:"id"`
	Name  string        `json:"name"`
	Value []interface{} `json:"value"`
}

func (Do) Method(id float64, raw []byte, w io.Writer) error {
	result := new(Enable)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Validate token
	var uid bson.ObjectId

	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER, WORKERLIST); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}

	//Get wallets list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}
	// validate worker exist
	var wrk BasicFullWorker
	if !bson.IsObjectIdHex(result.Id) {
		//fmt.Println(result.Id)
		return jsonrpc.InvalidP(id, w)
	}
	err = db.C(worker.String()).Find(bson.M{"_id": bson.ObjectIdHex(result.Id), "groupid": bson.M{"$in": list}}).One(&wrk)
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}

	_ = notify("refreshmodule", map[string]interface{}{"id": wrk.Id.Hex(), "name": result.Name, "value": result.Value})
	CallList.Insert(CallItem{Id: bson.ObjectIdHex(result.Id), Method: result.Name, User: uid, Time: time.Now().Unix()})

	_ = jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: map[string]interface{}{"module": result.Name, "value": result.Value}})
	return nil
}

var amdMethods = []string{"startprofile"}

func GetGroupForMethod(method string) Group {
	if sort.SearchStrings(amdMethods, method) > 0 {
		return ADMIN
	}
	return BASICUSER
}
