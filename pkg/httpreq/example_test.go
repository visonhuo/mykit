package httpreq_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/visonhuo/mykit/pkg/httpreq"
)

func ExampleRequest() {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(writer, "hello")
	}))
	defer ts.Close()

	respBody, err := httpreq.NewRequest(context.Background(), http.MethodGet, ts.URL, nil).
		Header("Header", "h1", "h2").
		Query("query", "q1", "q2").
		Call().Expect200().Response()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(respBody))
	// Output:
	// hello
}

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
