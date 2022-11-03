package traceroute

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	netpacket "github.com/visonhuo/mykit/internal/net/packet"
	"golang.org/x/net/ipv4"
)

type probePacket struct {
	identify int
	sendTime time.Time
}

type session struct {
	ctx     context.Context
	server  *Server
	dstIP   net.IP
	packetQ chan packet
	future  *Future
}

func (s *session) init() error {
	s.packetQ = make(chan packet, 16)
	s.future = &Future{finish: make(chan struct{})}
	return nil
}

func (s *session) run(opts Options) {
	opts.init()
	var result = Result{DstIP: s.dstIP, Opts: opts}
	var err error
	defer func() {
		if e := recover(); e != nil {
			s.future.done(result, fmt.Errorf("panic: %v", e))
		} else {
			s.future.done(result, err)
		}
	}()

	pc := make(chan probePacket, 16)
	go s.sendProbePackets(pc, opts)

	id2SendTime := make(map[int]time.Time, (opts.MaxHop-opts.FirstHop)*opts.Attempts)
	var timeC <-chan time.Time
	timeout := false
	for !timeout {
		select {
		case <-s.ctx.Done():
			err = s.ctx.Err()
			return

		case <-s.server.close:
			err = errors.New("server closed")
			return

		case <-timeC:
			timeout = true
			break

		case probe, ok := <-pc:
			if !ok { // finish sending
				timeC = time.NewTimer(opts.Timeout).C
				pc = nil
				break
			}
			id2SendTime[probe.identify] = probe.sendTime

		case pkt := <-s.packetQ:
			sendTime, ok := id2SendTime[pkt.identify]
			if !ok {
				continue
			}
			rtt := pkt.recvTime.Sub(sendTime)
			if rtt > opts.Timeout {
				continue
			}

			ttl := ((pkt.identify - 1) / opts.Attempts) + opts.FirstHop
			result.aggregate(ttl, pkt.addr.IP, rtt)
		}
	}
}

func (s *session) sendProbePackets(pc chan<- probePacket, opts Options) {
	defer close(pc)

	var identify int
	srcPort := randomPort()
	payload := make([]byte, opts.PacketSize)
	for ttl := opts.FirstHop; ttl <= opts.MaxHop; ttl++ {
		for i := 0; i < opts.Attempts; i++ {
			if s.future.isDone() {
				return
			}

			identify += 1
			dstPort := identify + opts.Port
			header, udpPktBytes, err := s.generalUDPPacket(srcPort, dstPort, identify, ttl, payload)
			if err != nil {
				s.logf("General UDP packet failed (%v):%v", s.dstIP, err)
				continue
			}

			sendTime := time.Now()
			err = s.server.write(header, udpPktBytes)
			if err != nil {
				s.logf("Write UDP packet failed (%v):%v", s.dstIP, err)
				continue
			}
			pc <- probePacket{
				identify: identify,
				sendTime: sendTime,
			}
		}
	}
}

func (s *session) acceptPacket(pkt packet) {
	if s == nil {
		return
	}
	timer := time.NewTimer(s.server.config.DispatchTimeout)
	defer timer.Stop()
	select {
	case s.packetQ <- pkt:
	case <-timer.C:
		s.logf("Handle packet timeout: %v", pkt)
	}
}

func (s *session) generalUDPPacket(srcPort, dstPort, identify, ttl int, payload []byte) (ipv4.Header, []byte, error) {
	ipHeader := ipv4.Header{
		Version:  ipv4.Version,
		Len:      ipv4.HeaderLen,
		ID:       identify,
		Flags:    ipv4.DontFragment,
		TTL:      ttl,
		Protocol: protocolUDP,
		Src:      s.server.config.LocalSrcIP,
		Dst:      s.dstIP,
	}
	udp := netpacket.UDPv4{
		SrcPort: uint16(srcPort),
		DstPort: uint16(dstPort),
	}
	udpBytes, err := udp.Marshal(ipHeader, payload)
	if err != nil {
		return ipHeader, nil, err
	}
	ipHeader.TotalLen = ipv4.HeaderLen + len(udpBytes)
	return ipHeader, udpBytes, nil
}

func (s *session) logf(format string, args ...interface{}) {
	s.server.logf(format, args...)
}
