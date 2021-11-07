package api

import (
	"fmt"
	"github.com/bledniua/rpc/jsonrpc"
	"github.com/globalsign/mgo/bson"
	"github.com/json-iterator/go"
	"golang.org/x/crypto/bcrypt"
	"io"
	"math"
	"sort"
	"time"
)

const BCryptCost int = 12

type SignUp struct {
	Token    string `json:"token" validate:"required"`
	Login    string `json:"login" validate:"required,excludesall=!@#?"`
	Password string `json:"password" validate:"required,gt=6,lt=32"`
	Email    string `json:"email" validate:"required,email"`
	Key      string `json:"key" validate:"required,len=64"`
}

func (SignUp) Method(id float64, raw []byte, w io.Writer) error {
	result := new(SignUp)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//validate token
	if err := ValidateToken(result.Token); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//CheckIsLoginExist
	if user, err := findUserByLogin(result.Login); err == nil && user != nil {
		return jsonrpc.SendError(id, w, loginAlreadyExist)
	}
	//validate key
	key, err := GetKeyAndRemove(result.Key)
	if err != nil {
		return jsonrpc.SendError(id, w, keyNotFound)
	}
	//Encrypt password
	pass, err := bcrypt.GenerateFromPassword([]byte(result.Password), BCryptCost)
	if err != nil {
		return jsonrpc.SendError(id, w, cantCreateHashPassword)
	}
	//createNewUser
	user := &User{Login: result.Login, Email: result.Email, Password: pass, Groups: key.Groups}
	if err := createNewUser(user); err != nil {
		return jsonrpc.SendError(id, w, cantCreateNewUser)
	}
	if err := AddUserToken(result.Token, user); err != nil {
		return jsonrpc.SendError(id, w, cantSingIn)
	}
	return jsonrpc.Send(id, w, fmt.Sprintf("Hello %s", user.Login))
}

type SignIn struct {
	Token    string `json:"token" validate:"required"`
	Login    string `json:"login" validate:"required,excludesall=!@#?"`
	Password string `json:"password" validate:"required,gt=6,lt=32"`
}

func (SignIn) Method(id float64, raw []byte, w io.Writer) error {
	result := new(SignIn)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//validate token
	if err := ValidateToken(result.Token); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//CheckIsUserExist
	var user *User
	if user, err = findUserByLogin(result.Login); err != nil || user == nil {
		return jsonrpc.SendError(id, w, cantFindUser)
	}
	//CheckPassword
	if bcrypt.CompareHashAndPassword(user.Password, []byte(result.Password)) != nil {
		return jsonrpc.SendError(id, w, wrongPassword)
	}
	if err := AddUserToken(result.Token, user); err != nil {
		return jsonrpc.SendError(id, w, cantSingIn)
	}
	//fmt.Printf("Hello %s\n", user.Login)
	return jsonrpc.Send(id, w, fmt.Sprintf("Hello %s", user.Login))
}

type SingOut struct {
	Token string `json:"token" validate:"required"`
}

func (SingOut) Method(id float64, raw []byte, w io.Writer) error {
	result := new(SingOut)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Validate token
	if err := ValidateToken(result.Token); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	var data *UserData
	if data, err = RemoveUserToken(result.Token); err != nil {
		return jsonrpc.SendError(id, w, cantSingOut)
	}
	var user *User
	if user, err = findUserById(data.UserId); err != nil || user == nil {
		return jsonrpc.SendError(id, w, cantFindUser)
	}
	return jsonrpc.Send(id, w, fmt.Sprintf("Bye %s", user.Login))
}

type Me struct {
	Token string `json:"token" validate:"required"`
}

func (Me) Method(id float64, raw []byte, w io.Writer) error {
	result := new(SingOut)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	var uid bson.ObjectId
	var user *User

	//Validate token
	if uid, err = ValidateAndGetUserId(result.Token); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	if user, err = findUserById(uid); err != nil || user == nil {
		return jsonrpc.SendError(id, w, cantFindUser)
	}
	return jsonrpc.Send(id, w, fmt.Sprintf("Hello %s", user.Login))
}

type GetChart struct {
	Token string `json:"token" validate:"required"`
}

type ChartData struct {
	Aid  bson.ObjectId `json:"group_id" bson:"aid"`
	Time int64         `json:"time" bson:"time"`
	Khs  int64         `json:"khs" bson:"khs"`
}

func (GetChart) Method(id float64, raw []byte, w io.Writer) error {
	//fmt.Println("call", string(raw))
	result := new(GetChart)
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

	//Get groups list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	startAt := time.Now().Unix() - 86400
	startAt = startAt - startAt%3600
	endAt := startAt + 86400
	charts := make([]ChartData, 0, len(list)*24)
	err = db.C(accounthash.String()).Find(bson.M{"aid": bson.M{"$in": list}, "time": bson.M{"$gt": startAt, "$lte": endAt}}).All(&charts)
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}

	return jsonrpc.Send(id, w, 360, startAt, endAt, charts)
}

type GetFullChart struct {
	Token string `json:"token" validate:"required"`
}

func (GetFullChart) Method(id float64, raw []byte, w io.Writer) error {
	//fmt.Println("call", string(raw))
	result := new(GetFullChart)
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

	//Get groups list
	list, err := GetGroupsIdList(uid)
	if err != nil || len(list) == 0 {
		return jsonrpc.SendError(id, w, cantGetGroupList)
	}

	startAt := time.Now().Unix() - 2592000
	startAt = startAt - startAt%3600
	endAt := startAt + 2592000
	charts := make([]ChartData, 0, len(list)*24)
	err = db.C(accounthash.String()).Find(bson.M{"aid": bson.M{"$in": list}, "time": bson.M{"$gt": startAt, "$lte": endAt}}).All(&charts)
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetWorker)
	}

	return jsonrpc.Send(id, w, 360, startAt, endAt, charts)
}

type GetTotalInfo struct {
	Token string `json:"token" validate:"required"`
}

func (GetTotalInfo) Method(id float64, raw []byte, w io.Writer) error {
	result := new(GetTotalInfo)
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

	onlineLine := time.Now().Add(time.Second * 60 * -1).Unix()
	dayBefore := time.Now().Unix()
	dayBefore -= dayBefore % 3600
	dayBefore -= 82800

	var total, online int
	total, err = db.C(worker.String()).Find(bson.M{"groupid": bson.M{"$in": list}}).Count()
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetData)
	}
	online, err = db.C(worker.String()).Find(bson.M{"groupid": bson.M{"$in": list}, "update_at": bson.M{"$gte": onlineLine}}).Count()
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetData)
	}
	snaps := make([]AccountHashSnap, 0, 24)
	err = db.C(accounthash.String()).Find(bson.M{"aid": bson.M{"$in": list}, "time": bson.M{"$gte": dayBefore}}).All(&snaps)
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetData)
	}
	sort.Slice(snaps, func(i, j int) bool {
		return snaps[i].Time > snaps[j].Time
	})
	totalHour := float64(0)
	totalDay := float64(0)
	if len(snaps) > 0 {
		for _, snap := range snaps {
			if snap.Time == snaps[0].Time {
				totalHour += snap.Khs
			} else {
				totalDay += snap.Khs
			}
		}
	}
	return jsonrpc.Send(id, w, total, online, math.Round(totalHour/(float64(time.Now().Unix()%3600)/10)), math.Round(totalDay/8280))
}

type GetNews struct {
	Token string `json:"token" validate:"required"`
	Page  int    `json:"page" validate:"required,min=1"`
}

func (GetNews) Method(id float64, raw []byte, w io.Writer) error {
	result := new(GetNews)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}

	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//Validate token
	if _, err = ValidateAndGetUserId(result.Token, BASICUSER); err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	news := make([]Post, 8)
	err = db.C(post.String()).Find(nil).Limit(8).Sort("-createat").Skip((result.Page - 1) * 8).All(&news)
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetData)
	}
	return jsonrpc.Send(id, w, news)
}

type GetTime struct {
}

func (GetTime) Method(id float64, raw []byte, w io.Writer) error {
	//result := new(GetTime)
	//if jsoniter.Unmarshal(raw, result) != nil {
	//	return jsonrpc.InvalidP(id, w)
	//}

	//Validate message
	//err := validate.Struct(result)
	//if err != nil {
	//	return jsonrpc.SendError(id, w, validateError(err))
	//}

	return jsonrpc.Send(id, w, time.Now().Unix())
}
