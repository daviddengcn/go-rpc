/*
	Package rpc makes a RPC service through HTTP protocol.
	Different from the net/rpc package. The methods of the serving object can be
	of much flexible. The only requirement is the parameters and return values
	are be represented by JSON.

	The first parameter could be *http.Request, which will be set to the
	instance in the handler, and not count to the NumIn in client.Call. This is
	especially usefully in Google App Engine.

	Example:
		//// Shared
		type Arith int

		func (*Arith) Add(a, b int) int {
			return a + b
		}

		func (*Arith) Sub(r *http.Request, a, b int) int {
			return a - b
		}

		//// Server
		Register(new(Arith))

		go http.ListenAndServe(":1235", nil)

		//// Client
		rpcClient := NewClient(http.DefaultClient, "http://localhost:1235")

		A, B := 1, 2
		var C int
		err := rpcClient.Call(2, "Add", A, B, &C)
		err = rpcClient.Call(2, "Sub", A, B, &C)
*/
package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
)

type methodInfo struct {
	funcValue   reflect.Value
	needRequest bool
	inTypes     []reflect.Type
}

/*
	Server represents an instance for server. *Server satisfies http.Handler
	interface.
*/
type Server struct {
	oValue  reflect.Value
	methods map[string]*methodInfo
}

/*
	The default path
*/
const DefaultPath = "/_http_rpc"

const (
	errCodeOk            int = iota // Ok
	errCodeUnknownMethod            // Unknown method name
)

type rpcResponse struct {
	Code int
	Outs []string
}

/**** Server ****/

// Implementation of http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mname := r.FormValue("method")

	mi := s.methods[mname]
	if mi == nil {
		res := rpcResponse{
			Code: errCodeUnknownMethod,
		}
		replyJson, _ := json.Marshal(res)
		w.Write(replyJson)
		return
	}

	// plus 1 for the receiver, and needRequest
	callArr := make([]reflect.Value, 1, len(mi.inTypes)+2)
	callArr[0] = s.oValue

	if mi.needRequest {
		callArr = append(callArr, reflect.ValueOf(r))
	}

	inJsons := r.Form["in"]
	// set parameters
	for i := range mi.inTypes {
		pInV := reflect.New(mi.inTypes[i])
		json.Unmarshal([]byte(inJsons[i]), pInV.Interface())
		callArr = append(callArr, reflect.Indirect(pInV))
	}

	outs := mi.funcValue.Call(callArr)

	outJsons := make([]string, len(outs))
	for i := range outs {
		outJson, _ := json.Marshal(outs[i].Interface())
		outJsons[i] = string(outJson)
	}
	resp := rpcResponse{
		Code: errCodeOk,
		Outs: outJsons,
	}
	replyJson, _ := json.Marshal(resp)
	w.Write(replyJson)
}

var (
	pHttpRequestType reflect.Type = reflect.TypeOf((*http.Request)(nil))
)

/*
	NewServer creates a *Server instance for an object o, whose methods are
	called for RPC service.
*/
func NewServer(o interface{}) *Server {
	server := &Server{
		oValue:  reflect.ValueOf(o),
		methods: make(map[string]*methodInfo),
	}

	oType := reflect.TypeOf(o)
	for m := 0; m < oType.NumMethod(); m++ {
		method := oType.Method(m)
		mi := &methodInfo{
			funcValue: method.Func,
		}
		tp := method.Type

		if tp.NumIn() > 1 {
			first := 1 // tp.In(0) is the receiver
			if tp.In(1) == pHttpRequestType {
				mi.needRequest = true
				first++
			}

			if tp.NumIn() > first {
				mi.inTypes = make([]reflect.Type, tp.NumIn()-first)
				for i := range mi.inTypes {
					mi.inTypes[i] = tp.In(i + first)
				}
			}
		}
		server.methods[method.Name] = mi
	}

	return server
}

/*
	Register calls RegisterPath with the DefaultPath.
*/
func Register(o interface{}) {
	RegisterPath(o, DefaultPath)
}

/*
	RegisterPath registers a Server with http.Handle
*/
func RegisterPath(o interface{}, path string) {
	http.Handle(path, NewServer(o))
}

/**** Client ****/

/*
	Client represents an RPC client.
*/
type Client struct {
	httpClient *http.Client
	host       string
}

/*
	NewClient creates a *Client with DefaultPath
*/
func NewClient(httpClient *http.Client, host string) *Client {
	return NewClientPath(httpClient, host, DefaultPath)
}

/*
	NewClientPath creates a *Client with specified path.
*/
func NewClientPath(httpClient *http.Client, host, path string) *Client {
	return &Client{
		httpClient: httpClient,
		host:       host + path,
	}
}

/*
	Call makes a RPC. numIn here is to distinguish the parameters(in's) from
	return values (out's) in inPOuts. For the return values, the pointers of
	each receiving variable is needed.
*/
func (c *Client) Call(numIn int, method string, inPOuts ...interface{}) error {
	inJsons := make([]string, numIn)
	for i := range inJsons {
		inJson, err := json.Marshal(inPOuts[i])
		if err != nil {
			return err
		}

		inJsons[i] = string(inJson)
	}
	resp, err := c.httpClient.PostForm(c.host, url.Values{
		"method": {method},
		"in":     inJsons,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	var res rpcResponse
	err = dec.Decode(&res)
	if err != nil {
		return err
	}

	if res.Code != errCodeOk {
		return errors.New(fmt.Sprintf("Error code: %d", res.Code))
	}

	for i := range res.Outs {
		err := json.Unmarshal([]byte(res.Outs[i]), inPOuts[numIn+i])
		if err != nil {
			return err
		}
	}

	return nil
}
