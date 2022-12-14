package traceroute_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/visonhuo/mykit/pkg/traceroute"
)

func TestServer(t *testing.T) {
	srv, err := traceroute.NewServer(traceroute.Config{})
	require.NoError(t, err)

	host := "www.google.com"
	future, err := srv.Traceroute(context.Background(), host, traceroute.Options{})
	require.NoError(t, err)
	require.NoError(t, future.Error())
	result := future.Result()
	fmt.Println(result)
}
