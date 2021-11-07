package api

import (
	"fmt"
	"github.com/bledniua/rpc/jsonrpc"
	"github.com/globalsign/mgo/bson"
	jsoniter "github.com/json-iterator/go"
	"io"
)

type GetPayments struct {
	Token string `json:"token" validate:"required"`
	Page  int    `json:"page" validate:"required,min=1"`
	Limit int    `json:"limit" validate:"required,min=1,max=32"`
}

func (GetPayments) Method(id float64, raw []byte, w io.Writer) error {
	result := new(GetPayments)
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

	//Get wallets list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	pList := make([]Payment, 0)
	err = db.C(payments.String()).Find(bson.M{"groupid": bson.M{"$in": list}}).Sort("-createat").Skip(result.Page - 1).Limit(result.Limit).All(&pList)
	if err != nil {
		return jsonrpc.SendError(id, w, cantFindPayments)
	}

	return jsonrpc.Send(id, w, pList)
}

type GetPayment struct {
	Token   string `json:"token" validate:"required"`
	GroupId string `json:"group_id" validate:"required"`
	Page    int    `json:"page" validate:"required,min=1"`
	Limit   int    `json:"limit" validate:"required,min=1,max=32"`
}

func (GetPayment) Method(id float64, raw []byte, w io.Writer) error {
	result := new(GetPayment)
	if jsoniter.Unmarshal(raw, result) != nil {
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

	//validate token
	var uid bson.ObjectId
	if uid, err = ValidateAndGetUserId(result.Token, BASICUSER); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}

	//Get wallets list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}
	found := false
	for _, gr := range list {
		if gr.Hex() == result.GroupId {
			found = true
			break
		}
	}
	if !found {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("premission denied")))
	}

	pList := make([]Payment, 0)
	err = db.C(payments.String()).Find(bson.M{"groupid": bson.ObjectIdHex(result.GroupId)}).Sort("-createat").Skip(result.Page - 1).Limit(result.Limit).All(&pList)
	if err != nil {
		return jsonrpc.SendError(id, w, cantFindPayments)
	}

	return jsonrpc.Send(id, w, pList)
}
