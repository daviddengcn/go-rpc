go-rpc
======

An RPC framework using http requests.

[GoDoc](http://godoc.org/github.com/daviddengcn/go-rpc)


Different from the <code>net/rpc</code> package. The methods of the serving object can be of much flexible.
The only requirement is the parameters and return values are be represented by JSON.

Example
=======

```go
  	//// Shared
		type Arith int
		
		func (*Arith) Add(a, b int) int {
			return a + b
		}
		
		//// Server
		Register(new(Arith))
	
		http.ListenAndServe(":1235", nil)
    
		
		//// Client
		client := NewClient(http.DefaultClient, "http://localhost:1235")
	
		A, B := 1, 2
		var C int
		err := client.Call(2, "Add", A, B, &C)
```

You needn't define a type of input arguments for each method, as in using net/rpc.

LICENSE
=======
BSD license.
