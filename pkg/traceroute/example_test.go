package traceroute_test

import (
	"context"
	"fmt"

	"github.com/visonhuo/mykit/pkg/traceroute"
)

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

	// Output:
	// true
}
