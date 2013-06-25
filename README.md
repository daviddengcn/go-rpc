go-rpc
======

An RPC framework using http requests.

[GoDoc](http://godoc.org/github.com/daviddengcn/go-rpc)


Different from the <code>net/rpc</code> package. The methods of the serving object can be of much flexible.
The only requirement is the parameters and return values are be represented by JSON.

The first parameter could be *http.Request, which will be set to the instance in the handler,
and not counts to the NumIn in client.Call. This is especially usefully in 
[Google App Engine](https://developers.google.com/appengine/).

Example
=======

```go
//// The server model
type Arith int
		
func (*Arith) Add(a, b int) int {
	return a + b
}

func (*Arith) Sub(r *http.Request, a, b int) int {
	return a - b
}
		
//// Server
Register(new(Arith))
	
http.ListenAndServe(":1235", nil)
    
		
//// Client
client := NewClient(http.DefaultClient, "http://localhost:1235")
	
A, B := 1, 2
var C int
err := client.Call(2, "Add", A, B, &C)
err = client.Call(2, "Sub", A, B, &C)
```

You needn't define a type of input arguments for each method, as in using <code>net/rpc</code>.

LICENSE
=======
BSD license.
