package packet_test

import (
	"net"
	"testing"

	"encoding/binary"
	"github.com/stretchr/testify/require"
	"github.com/visonhuo/mykit/internal/net/packet"
	"golang.org/x/net/ipv4"
)

func TestUDPv4_Marshal(t *testing.T) {
	udp := packet.UDPv4{
		SrcPort: 8080,
		DstPort: 9090,
	}
	header := ipv4.Header{
		Protocol: 17, // udp protocol
		Src:      net.ParseIP("10.2.64.100"),
		Dst:      net.ParseIP("8.8.8.8"),
	}
	payload := make([]byte, 8)

	udpBytes, err := udp.Marshal(header, payload)
	require.NoError(t, err)
	header.TotalLen = ipv4.HeaderLen + len(udpBytes)
	require.Equal(t, udp.SrcPort, binary.BigEndian.Uint16(udpBytes[:2]))
	require.Equal(t, udp.DstPort, binary.BigEndian.Uint16(udpBytes[2:4]))
	require.Equal(t, uint16(16), binary.BigEndian.Uint16(udpBytes[4:6]))
	require.Equal(t, uint16(0x6246), binary.BigEndian.Uint16(udpBytes[6:8]))
	require.Len(t, udpBytes[8:], 8)
}
