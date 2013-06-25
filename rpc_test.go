package rpc

import (
	"fmt"
	"net/http"
	"testing"
)

type Arith int

func (*Arith) Add(a, b int) int {
	return a + b
}

func (*Arith) Sub(r *http.Request, a, b int) int {
	return a - b
}

func TestBasic(t *testing.T) {
	/** Server **/
	RegisterPath(new(Arith), "/rpc")
	go http.ListenAndServe(":1234", nil)

	/** Client **/
	rpcClient := NewClientPath(http.DefaultClient, "http://localhost:1234", "/rpc")

	var C int
	err := rpcClient.Call(2, "Add", 1, 2, &C)
	if err != nil {
		t.Errorf("rpcClient.Call failed: %v", err)
	} else {
		if C != 3 {
			t.Errorf("C should be 3")
		}
	}

	err = rpcClient.Call(2, "Sub", 2, 5, &C)
	if err != nil {
		t.Errorf("rpcClient.Call failed: %v", err)
	} else {
		if C != -3 {
			t.Errorf("C should be -3")
		}
	}
}

func ExampleArith() {
	Register(new(Arith))

	go http.ListenAndServe(":1235", nil)

	rpcClient := NewClient(http.DefaultClient, "http://localhost:1235")

	A, B := 1, 2
	var C int
	err := rpcClient.Call(2, "Add", A, B, &C)
	if err != nil {
		fmt.Printf("rpcClient.Call failed: %v\n", err)
	} else {
		fmt.Printf("%d + %d = %d\n", A, B, C)
	}
	err = rpcClient.Call(2, "Sub", 3, 7, &C)
	if err != nil {
		fmt.Printf("rpcClient.Call failed: %v\n", err)
	} else {
		fmt.Printf("%d - %d = %d\n", 3, 7, C)
	}
	// Output:
	// 1 + 2 = 3
	// 3 - 7 = -4
}
