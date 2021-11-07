package api

import (
	"bufio"
	"fmt"
	"github.com/bledniua/rpc"
	"github.com/bledniua/rpc/jsonrpc"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/go-playground/validator.v9"
	"net"
	"net/http"
	"runtime"
	"time"
)

const MaxBufferWriteSize int = 4096

var (
	bulk     *mgo.Bulk
	CallList *mgo.Bulk
	queue    *mgo.Bulk

	validate *validator.Validate
	server   *jsonrpc.Server
	db       *mgo.Database
)

func init() {
	runtime.GOMAXPROCS(1)
}

type ResultItem struct {
	Id     bson.ObjectId       `bson:"id"`
	Result jsoniter.RawMessage `bson:"result"`
	Time   int64               `json:"time"`
}

type QueueItem struct {
	Id      bson.ObjectId       `bson:"_id"`
	Request jsoniter.RawMessage `bson:"request"`
}

type CallItem struct {
	Id     bson.ObjectId `bson:"id"`
	User   bson.ObjectId `json:"user"`
	Method string        `bson:"method"`
	Time   int64         `json:"time"`
}

func dispatch(WorkerID string, DispatchName string, data ...interface{}) error {
	var msg []byte
	var err error
	if len(data) > 1 {
		msg, err = jsoniter.Marshal(jsonrpc.Response{Id: DispatchName, Result: data})
	} else {
		msg, err = jsoniter.Marshal(jsonrpc.Response{Id: DispatchName, Result: data[0]})
	}
	if err != nil {
		return err
	}
	bulk.Insert(ResultItem{Id: bson.ObjectIdHex(WorkerID), Result: msg, Time: time.Now().Unix()})

	return nil
}

func notify(DispatchName string, params interface{}) error {
	msg, err := jsoniter.Marshal(jsonrpc.Message{Method: DispatchName, Params: params})
	if err != nil {
		return err
	}
	queue.Insert(QueueItem{Id: bson.NewObjectId(), Request: msg})

	return nil
}

func Build(Db *mgo.Database) error {
	validate = validator.New()
	server = jsonrpc.NewServer(jsonrpc.NewBasicServerHandler())

	db = Db
	bulk = db.C("result").Bulk()
	queue = db.C("queue").Bulk()
	CallList = db.C(calllist.String()).Bulk()

	result := new(Token)
	iter := db.C(token.String()).Find(nil).Iter()
	for iter.Next(&result) {
		TokenCache[result.Access] = *result
	}
	if err := iter.Close(); err != nil {
		return err
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		server.AcceptOne(writer, request.Body)
		request.Body.Close()
	})

	list := []rpc.Method{
		Auth{},
		Refresh{},
		SignUp{},
		SignIn{},
		SingOut{},
		Me{},
		CreateNewGroup{},
		GetGroupList{},
		GetGroup{},
		EditGroup{},
		DeleteGroup{},
		WorkersList{},
		Worker{},
		SetProfile{},
		SetMirrorAddr{},
		Call{},
		DeleteWorker{},
		DeleteWorkers{},
		NewTask{},
		GetTaskList{},
		SetTaskStatus{},
		Enable{},
		Do{},

		GetPayments{},
		GetPayment{},
		GetTotalInfo{},
		GetChart{},
		GetFullChart{},
		GetNews{},
		GetTime{},

	}
	for _, method := range list {
		if err := server.NewMethod(method); err != nil {
			return err
		}
	}

	admList := []rpc.Method{GetHashRate{}, CreatePayment{}, EditPayment{}, CreatePost{}, UpdatePost{}, DeletePost{}}
	for _, method := range admList {
		if err := server.NewMethod(method); err != nil {
			return err
		}
	}

	return nil
}

func Flush() (*mgo.BulkResult, error) {
	tmp := CallList
	CallList = db.C(calllist.String()).Bulk()
	_, _ = tmp.Run()

	tmp = queue
	queue = db.C("queue").Bulk()
	_, _ = tmp.Run()

	tmp = bulk
	bulk = db.C("result").Bulk()
	return tmp.Run()
}

func RemoveTokens() error {
	info, err := db.C("token").RemoveAll(bson.M{"refresh_at": bson.M{"$lte": time.Now().Unix()}})
	if err != nil {
		return err
	}

	cleared := 0
	for idx, token := range TokenCache {
		if token.RefreshAt < time.Now().Unix() {
			delete(TokenCache, idx)
			cleared++
		}
	}
	fmt.Println("removed", info.Removed, "cleared", cleared, "tokens")
	return nil
}

func Run(listener net.Listener) error {
	server.Start()
	go func() {
		for {
			_, err := Flush()
			if err != nil {
				fmt.Println(err)
			}
			<-time.After(time.Millisecond * 500)
		}
	}()
	go func() {
		for {
			err := RemoveTokens()
			if err != nil {
				fmt.Println(err)
			}
			<-time.After(time.Minute * 5)
		}
	}()

	for server.IsRun() {
		conn, err := listener.Accept()
		if err != nil {
			_ = conn.Close()
		}
		go func() {
			fmt.Println("Connect from", conn.RemoteAddr())
			writer := bufio.NewWriterSize(conn, MaxBufferWriteSize)
			server.AcceptStream(writer, conn, writer.Flush)
		}()
	}
	return listener.Close()
}
