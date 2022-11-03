package traceroute

import (
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type packet struct {
	bytes    []byte
	size     int
	addr     *net.IPAddr
	recvTime time.Time
	identify int
}

type Server struct {
	config     Config
	rConn      net.PacketConn
	wConn      *ipv4.RawConn
	packetQ    chan packet
	close      chan struct{}
	shutdown   sync.Once
	ip2Session sync.Map
	bufPool    sync.Pool
}

func NewServer(cfg Config) (*Server, error) {
	cfg.init()
	srv := &Server{
		config:     cfg,
		packetQ:    make(chan packet, cfg.PacketQueueSize),
		close:      make(chan struct{}),
		ip2Session: sync.Map{},
		bufPool: sync.Pool{New: func() interface{} {
			return make([]byte, 1500)
		}},
	}
	for _, fn := range []func() error{
		srv.setupReadConn,
		srv.setupWriteConn,
	} {
		if err := fn(); err != nil {
			_ = srv.Shutdown()
			return nil, err
		}
	}

	go srv.server()
	go srv.dispatch()
	return srv, nil
}

func (s *Server) setupReadConn() error {
	conn, err := icmp.ListenPacket("ip4:icmp", net.IPv4zero.String())
	if err != nil {
		return err
	}
	s.rConn = conn
	return nil
}

func (s *Server) setupWriteConn() error {
	udpConn, err := net.ListenPacket("ip4:udp", s.config.LocalSrcIP.String())
	if err != nil {
		return err
	}
	rawConn, err := ipv4.NewRawConn(udpConn)
	if err != nil {
		return err
	}
	s.wConn = rawConn
	return nil
}

func (s *Server) server() {
	for {
		buf := s.bufPool.Get().([]byte)
		n, addr, err := s.rConn.ReadFrom(buf)
		if err != nil {
			_ = s.Shutdown()
			return
		}
		if n <= 0 {
			continue
		}

		select {
		case <-s.close:
			return
		case s.packetQ <- packet{bytes: buf, size: n, addr: addr.(*net.IPAddr), recvTime: time.Now()}:
		}
	}
}

func (s *Server) dispatch() {
	for {
		select {
		case <-s.close:
			return

		case pkt, ok := <-s.packetQ:
			if !ok {
				return
			}

			msg, err := icmp.ParseMessage(protocolICMPv4, pkt.bytes[:pkt.size])
			s.bufPool.Put(pkt.bytes)
			pkt.bytes = nil
			if err != nil {
				s.logf("Parse ICMPv4 message failed(len=%d, from=%v):%v", len(pkt.bytes), pkt.addr, err)
				continue
			}

			var originData []byte
			switch body := msg.Body.(type) {
			case *icmp.TimeExceeded:
				originData = body.Data
			case *icmp.DstUnreach:
				originData = body.Data
			default:
				continue
			}
			if len(originData) < ipv4.HeaderLen || originData[0]>>4 != ipv4.Version {
				s.logf("Invalid IP header from %v:%v", pkt.addr, originData)
				continue
			}
			originHeader, err := ipv4.ParseHeader(originData)
			if err != nil {
				s.logf("Parse IP header failed(%v): %v", originData, err)
				continue
			}

			pkt.bytes = nil
			pkt.identify = originHeader.ID
			value, _ := s.ip2Session.Load(originHeader.Dst.String())
			value.(*session).acceptPacket(pkt)
		}
	}
}

func (s *Server) write(header ipv4.Header, payload []byte) error {
	return s.wConn.WriteTo(&header, payload, nil)
}

func (s *Server) Traceroute(ctx context.Context, target string, opts Options) (*Future, error) {
	ipAddr, err := net.ResolveIPAddr("ip4", target)
	if err != nil {
		return nil, err
	}

	newSession := session{
		ctx:    ctx,
		server: s,
		dstIP:  ipAddr.IP,
	}
	value, loaded := s.ip2Session.LoadOrStore(ipAddr.IP.String(), &newSession)
	if loaded {
		return value.(*session).future, nil
	}

	_ = newSession.init()
	go newSession.run(opts)
	return newSession.future, nil
}

func (s *Server) Shutdown() error {
	var firstErr error
	s.shutdown.Do(func() {
		close(s.close)
		close(s.packetQ)
		if err := s.rConn.Close(); err != nil {
			firstErr = err
		}
		if err := s.wConn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	})
	return firstErr
}

func (s *Server) logf(format string, args ...interface{}) {
	s.config.ErrLogger.Printf(format, args...)
}
