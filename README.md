#  Mykit

Mykit is a personal golang toolkit project. **This project isn't aim to :**

* Provide a general, robust function set for people who want to use in production directly;
* Guaranteed to be 100% correct and responsible for any consequences of this project;

Instead, **this project is mainly aim to :** 

* **Share some ideas and practices** : Almost all of these ideas have been practiced in a real production environment, 
but since these codes are maintained by myself, they cannot be guaranteed to be 
100% correct, but can be used as a learning resource.
* **Implementation reference** : This doesn't mean you can't use the code here directly, just that when you do, 
it's best to do thorough testing to ensure it doesn't affect real users.

## Toolkit Dir
### traceroute
In computing, traceroute is computer network diagnostic commands for displaying possible routes
and measuring transit delays of packets across an Internet Protocol network.

In **pkg/traceroute** packet, we implement a simple traceroute library which allow us to 
emulate the behavior of the traceroute command on Linux. Usage example :

```go
func ExampleServer() {
	srv, err := traceroute.NewServer(traceroute.Config{})
	if err != nil {
		panic(err)
	}

	host := "www.google.com"
	future, err := srv.Traceroute(context.Background(), host, traceroute.Options{})
	if err != nil {
		panic(err)
	}
	if future.Error() != nil {
		panic(future.Error())
	}
	result := future.Result()
	fmt.Println(result.Reach)
}
```

In **cmd/traceroute** packet, we also implement a simple CLI traceroute tool which can diagnosis possible routes 
between local and multiple host.

```bash
# at cmd/traceroute dir
go build && sudo ./traceroute www.google.com www.baidu.com
```

Because we use raw connection in our implementation, so we should run our program by **root** user (or **setcap** on Linux). 
The use of raw connection mainly to achieve 2 purpose:
* Send custom UDP probe packet; (TTL field setting)
* Receive ICMP packet in our program;

### HTTP call tool
In **pkg/httpreq**, we make a tool to simplify the tedious HTTP client calling process in Golang.

It doesn't have any dependency on 3rd libraries, just basic logic encapsulation, 
simple enough and able to handle most HTTP API call scenarios.

```go
func ExampleClient() {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		_ = req.ParseForm()
		fmt.Fprintf(writer, "hello "+req.FormValue("name"))
	}))
	defer ts.Close()

	respBody, err := httpreq.NewClient(ts.Client()).
		NewRequest(context.Background(), http.MethodPost, ts.URL, url.Values{"name": []string{"world"}}).
		Header("Header", "h1", "h2").
		Query("query", "q1", "q2").
		Call().Expect200().Response()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(respBody))
	// Output:
	// hello world
}
```

Main supported feature:
* Simplify HTTP API call process.
* Support to custom **http.Client** instance.
* Support to register custom **Marshaller** implementation.
