package api

import (
	"fmt"
	"github.com/bledniua/rpc/jsonrpc"
	"github.com/globalsign/mgo/bson"
	jsoniter "github.com/json-iterator/go"
	"gitlab.com/toby3d/telegraph"
	"io"
	"net/url"
	"path"
	"sort"
	"time"
)

type GetHashRate struct {
	Token   string `json:"token" validate:"required"`
	StartAt int64  `json:"start_at" validate:"required"`
	EndAt   int64  `json:"end_at" validate:"required"`
}

func (GetHashRate) Method(id float64, raw []byte, w io.Writer) error {
	result := new(GetHashRate)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}

	result.StartAt -= result.StartAt % 3600
	result.EndAt -= result.EndAt % 3600

	//validate token
	if GetSecureToken() != result.Token {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("token timeout")))
	}

	snaps := make([]AccountHashSnap, 0)
	err = db.C(accounthash.String()).Pipe(
		[]bson.M{{
			"$match": bson.M{
				"$and": []bson.M{{
					"time": bson.M{"$gte": result.StartAt},
				}, {
					"time": bson.M{"$lt": result.EndAt},
				}},
			},
		}, {
			"$group": bson.M{
				"_id":   "$aid",
				"count": bson.M{"$sum": "$count"},
				"khs":   bson.M{"$sum": "$khs"},
			},
		}}).All(&snaps)
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetDataFromDatabase(err))
	}

	ids := make([]bson.ObjectId, len(snaps))
	for id, snap := range snaps {
		ids[id] = snap.GroupId
	}

	groups := make([]RigGroup, 0)
	err = db.C(group.String()).Find(bson.M{"_id": bson.M{"$in": ids}}).All(&groups)
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetDataFromDatabase(err))
	}
	ids = make([]bson.ObjectId, len(groups))
	for id, gr := range groups {
		ids[id] = gr.Owner
	}

	users := make([]User, 0)
	err = db.C(user.String()).Find(bson.M{"_id": bson.M{"$in": ids}}).All(&users)
	if err != nil {
		return jsonrpc.SendError(id, w, cantGetDataFromDatabase(err))
	}

	sort.Slice(snaps, func(i, j int) bool {
		return snaps[i].GroupId < snaps[j].GroupId
	})
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Id < groups[j].Id
	})
	sort.Slice(users, func(i, j int) bool {
		return users[i].Id < users[j].Id
	})

	list := make([]GetHashItem, len(users))
	for idx, user := range users {
		list[idx].Id = user.Id
		list[idx].Login = user.Login
		list[idx].Groups = make([]HashItemGroup, 0)
	}

	d := float64(result.EndAt-result.StartAt) / 10
	for _, gHash := range snaps {
		gIdx := sort.Search(len(groups), func(i2 int) bool {
			return groups[i2].Id >= gHash.GroupId
		})
		if !(gIdx < len(groups) && groups[gIdx].Id == gHash.GroupId) {
			continue
		}
		gr := groups[gIdx]
		owner := gr.Owner
		uIdx := sort.Search(len(users), func(i2 int) bool {
			return users[i2].Id >= owner
		})
		if !(uIdx < len(users) && users[uIdx].Id == gr.Owner) {
			continue
		}
		list[uIdx].Groups = append(list[uIdx].Groups, HashItemGroup{
			Id:       gr.Id,
			Address:  gr.Address,
			CoinBase: gr.CoinBase,
			Key:      gr.Key,
			Avg:      gHash.Khs / d,
		})
	}

	return jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: list})
}

type CreatePayment struct {
	Token   string        `json:"token" validate:"required"`
	GroupId bson.ObjectId `json:"group_id"`
	Amount  float64       `json:"amount"`
	Wallet  string        `json:"wallet"`
	Link    string        `json:"link"`
	Status  int           `json:"status"`
}

func (CreatePayment) Method(id float64, raw []byte, w io.Writer) error {
	result := new(CreatePayment)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//validate token
	if GetSecureToken() != result.Token {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("token timeout")))
	}

	payment := new(Payment)
	if jsoniter.Unmarshal(raw, payment) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	payment.Id = bson.NewObjectId()
	payment.CreateAt = time.Now().Unix()
	err = db.C(payments.String()).Insert(payment)
	if err != nil {
		return jsonrpc.SendError(id, w, cantCreatePayment(err))
	}

	return jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: payment})
}

type EditPayment struct {
	Token   string        `json:"token" validate:"required"`
	Id      bson.ObjectId `json:"id" validate:"required"`
	GroupId bson.ObjectId `json:"group_id" validate:"required"`
	Amount  float64       `json:"amount" validate:"required"`
	Wallet  string        `json:"wallet"`
	Link    string        `json:"link"`
	Status  int           `json:"status"`
}

func (EditPayment) Method(id float64, raw []byte, w io.Writer) error {
	result := new(EditPayment)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//validate token
	if GetSecureToken() != result.Token {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("token timeout")))
	}

	payment := new(Payment)
	if jsoniter.Unmarshal(raw, payments) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	payment.CreateAt = time.Now().Unix()
	err = db.C(payments.String()).UpdateId(payment.Id, payment)
	if err != nil {
		return jsonrpc.SendError(id, w, cantUpdatePayment(err))
	}

	return jsonrpc.SendResponse(id, w, jsonrpc.Response{Id: id, Result: payment})
}

type CreatePost struct {
	Token string `json:"token" validate:"required"`
	Link  string `json:"link"`
}

func (CreatePost) Method(id float64, raw []byte, w io.Writer) error {
	fmt.Println(string(raw))
	result := new(CreatePost)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//validate token
	if GetSecureToken() != result.Token {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("token timeout")))
	}

	u, err := url.Parse(result.Link)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	page, err := telegraph.GetPage(path.Base(u.Path), true)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	p := new(Post)
	p.Id = bson.NewObjectId()
	p.Title = page.Title
	p.Content = page.Content
	p.CreatAt = time.Now().Unix()
	err = db.C(post.String()).Insert(p)
	if err != nil {
		return jsonrpc.SendError(id, w, cantUpdatePayment(err))
	}

	return jsonrpc.Send(id, w, p.Id)
}

type UpdatePost struct {
	Token string `json:"token" validate:"required"`
	Id    string `json:"id"`
	Link  string `json:"link"`
}

func (UpdatePost) Method(id float64, raw []byte, w io.Writer) error {
	result := new(UpdatePost)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	if !bson.IsObjectIdHex(result.Id) {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("%s not object id", result.Id)))
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//validate token
	if GetSecureToken() != result.Token {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("token timeout")))
	}

	u, err := url.Parse(result.Link)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	page, err := telegraph.GetPage(path.Base(u.Path), true)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	err = db.C(post.String()).UpdateId(bson.ObjectIdHex(result.Id), bson.M{"$set": bson.M{"content": page.Content, "title": page.Title}})
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}

	return jsonrpc.Send(id, w, result.Id)
}

type DeletePost struct {
	Token string `json:"token" validate:"required"`
	Id    string `json:"id"`
}

func (DeletePost) Method(id float64, raw []byte, w io.Writer) error {
	result := new(DeletePost)
	if jsoniter.Unmarshal(raw, result) != nil {
		return jsonrpc.InvalidP(id, w)
	}
	if !bson.IsObjectIdHex(result.Id) {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("%s not object id", result.Id)))
	}
	//Validate message
	err := validate.Struct(result)
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}
	//validate token
	if GetSecureToken() != result.Token {
		return jsonrpc.SendError(id, w, validateError(fmt.Errorf("token timeout")))
	}

	err = db.C(post.String()).RemoveId(bson.ObjectIdHex(result.Id))
	if err != nil {
		return jsonrpc.SendError(id, w, validateError(err))
	}

	return jsonrpc.Send(id, w, result.Id)
}
