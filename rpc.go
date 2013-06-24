package rpc

import (
	"errors"
	"reflect"
	"net/http"
	"encoding/json"
	"net/url"
	
	"fmt"
//	"io/ioutil"
)

type methodInfo struct {
	funcValue reflect.Value
	inTypes []reflect.Type
	outTypes []reflect.Type
}

type Server struct {
	oValue reflect.Value
	methods map[string]*methodInfo
}

const defaultRpcPath = "/_http_rpc"

const (
	errCodeOk int = iota
	errCodeUnknownMethod
)

type rpcResponse struct {
	Code int
	Outs []string
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mname := r.FormValue("method")
fmt.Println("[Server.ServeHTTP] method:", mname)
	
	mi := s.methods[mname]
	if mi == nil {
		res := rpcResponse{
			Code: errCodeUnknownMethod,
		}
		replyJson, _ := json.Marshal(res)
		w.Write(replyJson)
		return
	}
	
fmt.Println("[Server.ServeHTTP] mi.inTypes:", mi.inTypes)
	callArr := make([]reflect.Value, len(mi.inTypes) + 1) // plus 1 for the receiver
	callArr[0] = s.oValue
	
	inJsons := r.Form["in"]
fmt.Println("[Server.ServeHTTP] inJsons:", inJsons)
	// set parameters
	for i := range mi.inTypes {
		pInV := reflect.New(mi.inTypes[i])
		json.Unmarshal([]byte(inJsons[i]), pInV.Interface())
		callArr[i + 1] = reflect.Indirect(pInV)
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

func Register(o interface{}) *Server {
	server := &Server {
		oValue: reflect.ValueOf(o),
		methods: make(map[string]*methodInfo),
	}
	
	oType := reflect.TypeOf(o)
	for m := 0; m < oType.NumMethod(); m ++ {
		method := oType.Method(m)
		tp := method.Type
		
		inTypes := make([]reflect.Type, tp.NumIn() - 1)
		for i := range inTypes {
			inTypes[i] = tp.In(i + 1) // first is receiver, so i + 1
		}
		outTypes := make([]reflect.Type, tp.NumOut())
		for i := range outTypes {
			outTypes[i] = tp.Out(i)
		}
		
		server.methods[method.Name] = &methodInfo {
			funcValue: method.Func,
			inTypes: inTypes,
			outTypes: outTypes,
		}
	}
	
	http.Handle(defaultRpcPath, server)
	
	return server
}

type Client struct {
	httpClient *http.Client
	host string
}

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
		"in": inJsons,
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
		return errors.New(fmt.Sprintf("Error code: %s", res.Code))
	}
	
	for i := range res.Outs {
		err := json.Unmarshal([]byte(res.Outs[i]), inPOuts[numIn + i])
		if err != nil {
			return err
		}
	}

	return nil
}

func NewClient(httpClient *http.Client, host string) *Client {
	return &Client {
		httpClient: httpClient,
		host: "http://" + host + defaultRpcPath,
	}
}
