package packet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/net/ipv4"
)

type UDPv4 struct {
	SrcPort uint16
	DstPort uint16
	length  uint16
	csum    uint16
}

type pseudoHeader struct {
	srcIP [4]byte
	dstIP [4]byte
	zero  uint8
	proto uint8
	len   uint16
}

func (u *UDPv4) Marshal(ipHeader ipv4.Header, payload []byte) ([]byte, error) {
	err := u.checkSum(ipHeader, payload)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	err = binary.Write(&b, binary.BigEndian, u)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&b, binary.BigEndian, payload)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (u *UDPv4) checkSum(ipHeader ipv4.Header, payload []byte) error {
	u.csum = 0
	if ipHeader.Src.To4() == nil {
		return fmt.Errorf("invalid src ip: %v", ipHeader.Src)
	}
	if ipHeader.Dst.To4() == nil {
		return fmt.Errorf("invalid dst ip: %v", ipHeader.Dst)
	}

	srcIP := ([]byte)(ipHeader.Src.To4())
	dstIP := ([]byte)(ipHeader.Dst.To4())
	u.length = uint16(8 + len(payload)) // UDP header length = 8
	ph := pseudoHeader{
		srcIP: [4]byte{srcIP[0], srcIP[1], srcIP[2], srcIP[3]},
		dstIP: [4]byte{dstIP[0], dstIP[1], dstIP[2], dstIP[3]},
		zero:  0,
		proto: uint8(ipHeader.Protocol),
		len:   u.length,
	}
	var b bytes.Buffer
	_ = binary.Write(&b, binary.BigEndian, &ph)
	_ = binary.Write(&b, binary.BigEndian, u)
	_ = binary.Write(&b, binary.BigEndian, &payload)
	u.csum = checksum(b.Bytes())
	return nil
}

func checksum(buf []byte) uint16 {
	sum := uint32(0)

	for ; len(buf) >= 2; buf = buf[2:] {
		sum += uint32(buf[0])<<8 | uint32(buf[1])
	}
	if len(buf) > 0 {
		sum += uint32(buf[0]) << 8
	}
	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}
	csum := ^uint16(sum)
	/*
	 * From RFC 768:
	 * If the computed checksum is zero, it is transmitted as all ones (the
	 * equivalent in one's complement arithmetic). An all zero transmitted
	 * checksum value means that the transmitter generated no checksum (for
	 * debugging or for higher level protocols that don't care).
	 */
	if csum == 0 {
		csum = 0xffff
	}
	return csum
}
