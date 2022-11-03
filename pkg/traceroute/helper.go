package traceroute

import (
	"errors"
	"math/rand"
	"net"
)

const (
	protocolICMPv4 = 1
	protocolUDP    = 17
)

const (
	defaultMinPort = 30000
	defaultMaxPort = 65535
)

func randomPort() int {
	return defaultMinPort + rand.Intn(defaultMaxPort-defaultMinPort)
}

func localIPv4() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for i := range addrs {
		if ipNet, ok := addrs[i].(*net.IPNet); ok && !ipNet.IP.IsLoopback() && len(ipNet.IP.To4()) == net.IPv4len {
			addr := make([]byte, 4)
			copy(addr, ipNet.IP.To4())
			return addr, nil
		}
	}
	return nil, errors.New("no valid local ipv4 address")
}
