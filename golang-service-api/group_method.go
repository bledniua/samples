package api

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/bledniua/rpc/jsonrpc"
	"github.com/globalsign/mgo/bson"
	"github.com/json-iterator/go"
	"io"
	"time"
)

type CreateNewGroup struct {
	Token    string `json:"token" validate:"required"`
	Address  string `json:"address"`
	CoinBase string `json:"coin_base"`
}

func (CreateNewGroup) Method(id float64, raw []byte, w io.Writer) error {
	result := new(CreateNewGroup)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//validate token
	var uid bson.ObjectId
	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//
	//Todo: Create limit group create
	//

	group := new(RigGroup)
	group.Address = result.Address
	group.CoinBase = result.CoinBase
	group.Owner = uid
	hash := make([]byte, 32)
	_, _ = rand.Read(hash)
	group.Key = base64.RawStdEncoding.EncodeToString(hash)
	if newRigGroup(group) != nil {
		return jsonrpc.SendError(id, w, cantCreateNewGroup)
	}

	return jsonrpc.Send(id, w, group)
}

type GetGroupList struct {
	Token string `json:"token" validate:"required"`
}

func (GetGroupList) Method(id float64, raw []byte, w io.Writer) error {
	result := new(GetGroupList)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//validate token
	var uid bson.ObjectId
	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}

	list := make([]RigGroup, 0)
	if GetGroupsList(uid, &list) != nil {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	return jsonrpc.Send(id, w, list)
}

type GetGroup struct {
	Token string `json:"token" validate:"required"`
	Id    string `json:"id" validate:"required"`
}

func (GetGroup) Method(id float64, raw []byte, w io.Writer) error {
	result := new(GetGroup)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//validate token
	var uid bson.ObjectId
	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}

	g := new(RigGroup)
	err = db.C(group.String()).Find(bson.M{"owner": uid, "_id": bson.ObjectIdHex(result.Id)}).One(g)
	if err != nil {
		return jsonrpc.SendError(id, w, cantFindGroup)
	}

	return jsonrpc.Send(id, w, g)
}

type EditGroup struct {
	Token    string           `json:"token" validate:"required"`
	Id       string           `json:"id" bson:"_id" validate:"required"`
	Name     *string          `json:"name,omitempty"`
	Address  *string          `json:"address,omitempty"`
	CoinBase *string          `json:"coin_base,omitempty"`
	Settings *GroupSettings   `json:"settings,omitempty"`
	SubUsers *[]bson.ObjectId `json:"sub_users,omitempty"`
}

func (EditGroup) Method(id float64, raw []byte, w io.Writer) error {
	result := new(EditGroup)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//validate token
	var uid bson.ObjectId
	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}

	set := bson.M{}
	if result.Name != nil {
		set["name"] = result.Name
	}
	if result.Address != nil {
		set["address"] = result.Address
	}
	if result.CoinBase != nil {
		set["coinbase"] = result.CoinBase
	}
	if result.Settings != nil {
		set["settings"] = result.Settings
	}
	if result.SubUsers != nil {
		set["subusers"] = result.SubUsers
	}

	set["updateat"] = time.Now().Unix()
	err = db.C(group.String()).Update(bson.M{"owner": uid, "_id": bson.ObjectIdHex(result.Id)}, bson.M{
		"$set": set,
	})
	if err != nil {
		return jsonrpc.SendError(id, w, cantFindGroup)
	}

	g := new(RigGroup)
	err = db.C(group.String()).Find(bson.M{"owner": uid, "_id": bson.ObjectIdHex(result.Id)}).One(g)
	if err != nil {
		return jsonrpc.SendError(id, w, cantFindGroup)
	}

	return jsonrpc.Send(id, w, g)
}

type DeleteGroup struct {
	Token string `json:"token" validate:"required"`
	Id    string `json:"id" bson:"_id" validate:"required"`
}

func (DeleteGroup) Method(id float64, raw []byte, w io.Writer) error {
	result := new(DeleteGroup)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//validate token
	var uid bson.ObjectId
	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}

	total, err := db.C(worker.String()).Find(bson.M{"owner": uid, "groupid": bson.ObjectIdHex(result.Id)}).Count()
	if err == nil && total > 0 {
		return jsonrpc.SendError(id, w, cantRemoveWorkersExist)
	}

	err = db.C(group.String()).RemoveId(bson.ObjectIdHex(result.Id))
	if err != nil {
		return jsonrpc.SendError(id, w, cantFindGroup)
	}

	return jsonrpc.Send(id, w, true)
}
