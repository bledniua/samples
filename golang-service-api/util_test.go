package api

import (
	"bytes"
	"fmt"
	"github.com/bledniua/rpc/jsonrpc"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/json-iterator/go"
	"github.com/tidwall/pretty"
	"gopkg.in/go-playground/validator.v9"
	"io"
	"net/http"
	"os"
	"testing"
)

func TestGetSecureToken(t *testing.T) {
	fmt.Println(GetSecureToken())
}

func TestGetHashRate_Method(t *testing.T) {
	validate = validator.New()
	session, err := mgo.Dial("localhost:27018")
	if err != nil {
		panic(err)
	}
	db = session.DB(dbName)

	d := GetHashRate{
		//EndAt:   time.Now().Unix(),
		//StartAt: time.Now().Add(time.Hour * -1).Unix(),
		EndAt:   1557774000,
		StartAt: 1557770400,
		Token:   GetSecureToken(),
	}

	raw, _ := jsoniter.Marshal(d)
	fmt.Println(string(pretty.Pretty(raw)))
	err = d.Method(0, raw, os.Stdout)
	if err != nil {
		panic(err)
	}
}

func TestGetHashRate_Method2(t *testing.T) {
	d := GetHashRate{
		//EndAt:   time.Now().Unix(),
		//StartAt: time.Now().Add(time.Hour * 24 * 7 * -1).Unix(),
		EndAt:   1559433600,
		StartAt: 1558915200,
		//
		Token: GetSecureToken(),
	}

	fmt.Println(d)
	msg := jsonrpc.Message{Id: 0, Method: "gethashrate", Params: d}
	raw, _ := jsoniter.Marshal(msg)
	resp, err := http.Post("http://134.209.81.194:2080", "application/json", bytes.NewReader(raw))
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
}

func TestCreatePayment_Method(t *testing.T) {
	validate = validator.New()
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	db = session.DB(dbName)

	d := CreatePayment{
		Status:  1,
		GroupId: bson.ObjectIdHex("5c71bb63b2e2e47206720453"),
		Wallet:  "labuda",
		Amount:  1,
		Link:    "http://labuda.com",
		Token:   GetSecureToken(),
	}

	raw, _ := jsoniter.Marshal(d)
	fmt.Println(string(pretty.Pretty(raw)))
	err = d.Method(0, raw, os.Stdout)
	if err != nil {
		panic(err)
	}
}

func TestCreatePost_Method(t *testing.T) {
	validate = validator.New()
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	db = session.DB(dbName)

	d := CreatePost{
		Link:  "https://telegra.ph/Test-06-13-55",
		Token: GetSecureToken(),
	}

	raw, _ := jsoniter.Marshal(d)
	fmt.Println(string(pretty.Pretty(raw)))
	err = d.Method(0, raw, os.Stdout)
	if err != nil {
		panic(err)
	}
}
