package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/bledniua/rpc"
	"github.com/bledniua/rpc/jsonrpc"
	"github.com/globalsign/mgo"
	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/pretty"
	"gopkg.in/go-playground/validator.v9"
	"net"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

var statOnce = sync.Once{}
var conn net.Conn
var dbName = "miner-om"

var (
	accesstoken string
	login       = "vtlkru"
)

func StartUp(port int) error {
	statOnce.Do(func() {
		validate = validator.New()
		server = jsonrpc.NewServer(jsonrpc.NewBasicServerHandler())

		session, err := mgo.Dial("localhost:27017")
		if err != nil {
			panic(err)
		}
		db = session.DB(dbName)
		bulk = db.C("result").Bulk()
		queue = db.C("queue").Bulk()

		list := []rpc.Method{SignUp{},
			SignIn{},
			SingOut{},
			Auth{},
			Refresh{},
			CreateNewGroup{},
			WorkersList{},
			GetTotalInfo{},
			SetProfile{},
			NewTask{},
			GetTaskList{},
			SetTaskStatus{},
		}
		for _, method := range list {
			if err := server.NewMethod(method); err != nil {
				panic(err)
			}
		}

		l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			panic(err)
		}
		server.Start()
		go func() {
			defer l.Close()
			for server.IsRun() {
				conn, err := l.Accept()
				if err != nil {
					_ = conn.Close()
				}
				go func() {
					writer := bufio.NewWriterSize(conn, MaxBufferWriteSize)
					server.AcceptStream(writer, conn, writer.Flush)
				}()
			}
		}()

		conn, err = net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			panic(err)
		}
	})
	return nil
}

var msgid = int64(0)

func CallTest(params interface{}) ([]interface{}, error) {
	method := strings.ToLower(reflect.TypeOf(params).String())
	method = strings.Replace(method, "*", "", 1)
	method = strings.Split(method, ".")[1]

	reflect.TypeOf(params).String()
	message, err := json.Marshal(jsonrpc.Message{Id: atomic.AddInt64(&msgid, 1), Method: method, Params: params})
	_, err = conn.Write(message)
	if err != nil {
		panic(err)
	}

	dec := jsoniter.NewDecoder(conn)
	resp := jsonrpc.ResponseReceive{}

	err = dec.Decode(&resp)
	if err != nil {
		panic(err)
	}
	str, _ := jsoniter.Marshal(resp)
	fmt.Printf("--> %s\n<-- %s\n", pretty.Pretty(message), pretty.Pretty(str))
	buff := make([]byte, 1)
	for n, err := dec.Buffered().Read(buff); n > 0; n, err = dec.Buffered().Read(buff) {
		err = dec.Decode(&resp)
		if err != nil {
			panic(err)
		}
		str, _ := jsoniter.Marshal(resp)
		fmt.Printf("<-- %s\n", pretty.Pretty(str))
	}
	var arr []interface{}
	_ = jsoniter.Unmarshal(resp.Result, &arr)
	return arr, resp.Error
}

func DefaultAuth() ([]interface{}, error) {
	arr, err := CallTest(Auth{})
	accesstoken = arr[0].(string)
	return arr, err
}

func TestAuth(t *testing.T) {
	dbName = "miner-om"
	err := StartUp(8001)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, _ = DefaultAuth()
	_ = db.DropDatabase()
}

func TestRefresh(t *testing.T) {
	dbName = "miner-om"
	err := StartUp(8001)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	//auth
	arr, err := DefaultAuth()
	if err != nil {
		panic(err)
	}
	_, err = CallTest(Refresh{Token: arr[2].(string)})
	if err != nil {
		panic(err)
	}

	_ = db.DropDatabase()
}
