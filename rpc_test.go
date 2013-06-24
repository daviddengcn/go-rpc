package rpc

import (
	"testing"
	"net/http"
)

type Arith int

func (*Arith) Add(a, b int) int {
	return a + b
}

func TestBasic(t *testing.T) {
	Register(new(Arith))
	
	go http.ListenAndServe(":1234", nil)
	
	rpcClient := NewClient(http.DefaultClient, "localhost:1234")
	
	var C int
	err := rpcClient.Call(2, "Add", 1, 2, &C)
	if err != nil {
		t.Errorf("rpcClient.Call failed: %v", err)
	}
	
	if C != 3 {
		t.Errorf("C should be 3")
	}
}