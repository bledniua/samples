package api

import (
	"encoding/json"
	"fmt"
	"github.com/bledniua/rpc/jsonrpc"
	"github.com/globalsign/mgo/bson"
	"github.com/json-iterator/go"
	"golang.org/x/crypto/bcrypt"
	"net"
	"testing"
)

func BenchmarkBcrypt4(b *testing.B) {
	benchBcrypt(b, 4)
}
func BenchmarkBcrypt5(b *testing.B) {
	benchBcrypt(b, 5)
}
func BenchmarkBcrypt6(b *testing.B) {
	benchBcrypt(b, 6)
}
func BenchmarkBcrypt7(b *testing.B) {
	benchBcrypt(b, 7)
}
func BenchmarkBcrypt8(b *testing.B) {
	benchBcrypt(b, 8)
}
func BenchmarkBcrypt9(b *testing.B) {
	benchBcrypt(b, 9)
}
func BenchmarkBcrypt10(b *testing.B) {
	benchBcrypt(b, 10)
}
func BenchmarkBcrypt11(b *testing.B) {
	benchBcrypt(b, 11)
}
func BenchmarkBcrypt12(b *testing.B) {
	benchBcrypt(b, 12)
}
func BenchmarkBcrypt13(b *testing.B) {
	benchBcrypt(b, 13)
}
func BenchmarkBcrypt14(b *testing.B) {
	benchBcrypt(b, 14)
}

func benchBcrypt(b *testing.B, cost int) {
	b.StopTimer()
	pass, _ := bcrypt.GenerateFromPassword([]byte("foo bar"), cost)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_ = bcrypt.CompareHashAndPassword(pass, []byte("foo bar"))
	}
}

func TestSingUp(t *testing.T) {
	//dbName = "miner-om-test"
	//defer func() {
	//	err := db.DropDatabase()
	//	if err != nil {
	//		panic(err)
	//	}
	//}()
	err := StartUp(8001)
	if err != nil {
		panic(err)
	}

	_, e := DefaultAuth()
	if e != nil && e.Error() != "null" {
		panic(e)
	}

	key, err := GetNewKey(Key{Groups: []Group{ADMIN, BASICUSER}})
	if err != nil {
		_, _ = GetKeyAndRemove(key)
		panic(err)
	}
	_, err = CallTest(SignUp{Token: accesstoken, Login: login, Key: key, Password: "bioqcomp789", Email: "bledniua@gmail.com"})
	if err != nil && err.Error() != "null" {
		panic(err)
	}
}

func TestSingIn(t *testing.T) {
	err := StartUp(8001)
	if err != nil {
		panic(err)
	}

	conn, err := net.Dial("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	message, err := json.Marshal(jsonrpc.Message{Id: 0, Method: "auth", Params: []interface{}{}})
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
	fmt.Printf("--> %s\n<-- %s\n", message, resp.Result)
	var arr []interface{}
	_ = jsoniter.Unmarshal(resp.Result, &arr)

	defer func(access string) {
		err = db.C(token.String()).Remove(bson.M{"access": access})
		if err != nil {
			panic(err)
		}
	}(arr[0].(string))

	key, err := GetNewKey(Key{Groups: []Group{BASICUSER, WORKERLIST}})
	if err != nil {
		panic(err)
	}
	defer func() {
		_, _ = GetKeyAndRemove(key)
	}()
	login := "vtlkru"

	message, err = json.Marshal(jsonrpc.Message{Id: 1, Method: "signup", Params: SignUp{Token: arr[0].(string), Login: login, Key: key, Password: "bioqcomp789", Email: "bledniua@gmail.com"}})
	_, err = conn.Write(message)
	if err != nil {
		panic(err)
	}

	err = dec.Decode(&resp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("--> %s\n<-- %s\n", message, resp.Result)
	defer func() {
		err = db.C(user.String()).Remove(bson.M{"login": login})
		if err != nil {
			panic(err)
		}
	}()

	message, err = json.Marshal(jsonrpc.Message{Id: 2, Method: "signin", Params: SignIn{Login: login, Token: arr[0].(string), Password: "bioqcomp789"}})
	_, err = conn.Write(message)
	if err != nil {
		panic(err)
	}

	err = dec.Decode(&resp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("--> %s\n<-- %s\n", message, resp.Result)
}

func TestSingOut(t *testing.T) {
	err := StartUp(8001)
	if err != nil {
		panic(err)
	}

	conn, err := net.Dial("tcp", ":8000")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	message, err := json.Marshal(jsonrpc.Message{Id: 0, Method: "auth", Params: []interface{}{}})
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
	fmt.Printf("--> %s\n<-- %s\n", message, resp.Result)
	var arr []interface{}
	_ = jsoniter.Unmarshal(resp.Result, &arr)

	defer func(access string) {
		err = db.C(token.String()).Remove(bson.M{"access": access})
		if err != nil {
			panic(err)
		}
	}(arr[0].(string))

	key, err := GetNewKey(Key{Groups: []Group{ADMIN}})
	if err != nil {
		panic(err)
	}
	defer func() {
		_, _ = GetKeyAndRemove(key)
	}()
	login := "vtlkru"

	message, err = json.Marshal(jsonrpc.Message{Id: 1, Method: "signup", Params: SignUp{Token: arr[0].(string), Login: login, Key: key, Password: "bioqcomp789", Email: "bledniua@gmail.com"}})
	_, err = conn.Write(message)
	if err != nil {
		panic(err)
	}

	err = dec.Decode(&resp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("--> %s\n<-- %s\n", message, resp.Result)
	defer func() {
		err = db.C(user.String()).Remove(bson.M{"login": login})
		if err != nil {
			panic(err)
		}
	}()

	message, err = json.Marshal(jsonrpc.Message{Id: 2, Method: "signin", Params: SignIn{Login: login, Token: arr[0].(string), Password: "bioqcomp789"}})
	_, err = conn.Write(message)
	if err != nil {
		panic(err)
	}

	err = dec.Decode(&resp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("--> %s\n<-- %s\n", message, resp.Result)

	message, err = json.Marshal(jsonrpc.Message{Id: 3, Method: "singout", Params: SingOut{Token: arr[0].(string)}})
	_, err = conn.Write(message)
	if err != nil {
		panic(err)
	}

	err = dec.Decode(&resp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("--> %s\n<-- %s\n", message, resp.Result)

}
