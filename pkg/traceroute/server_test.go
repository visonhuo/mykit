package traceroute_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/visonhuo/mykit/pkg/traceroute"
)

func TestServer(t *testing.T) {
	localSrcIP, err := localIPv4()
	require.NoError(t, err)

	srv, err := traceroute.NewServer(traceroute.Config{LocalSrcIP: localSrcIP})
	require.NoError(t, err)

	domain := "www.google.com"
	future, err := srv.Traceroute(context.Background(), domain, traceroute.Options{})
	require.NoError(t, err)
	require.NoError(t, future.Error())
	result := future.Result()
	fmt.Println(result)
}

func localIPv4() (ip net.IP, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if len(ipNet.IP.To4()) == net.IPv4len {
				addr := make([]byte, 4)
				copy(addr, ipNet.IP.To4())
				ip = addr
				return
			}
		}
	}
	err = errors.New("no valid local ipv4 address")
	return
}
