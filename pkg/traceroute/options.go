package traceroute

import "time"

const (
	defaultPort       = 33434
	defaultFirstHop   = 1
	defaultMaxHop     = 64
	defaultTimeout    = 1 * time.Second
	defaultAttempts   = 3
	defaultPacketSize = 16
)

type Options struct {
	Port       int
	FirstHop   int
	MaxHop     int
	Attempts   int
	Timeout    time.Duration
	PacketSize int
}

func (o *Options) init() {
	if o.Port <= 0 {
		o.Port = defaultPort
	}
	if o.MaxHop <= 0 {
		o.MaxHop = defaultMaxHop
	}
	if o.FirstHop <= 0 {
		o.FirstHop = defaultFirstHop
	}
	if o.Timeout <= 0 {
		o.Timeout = defaultTimeout
	}
	if o.PacketSize <= 0 {
		o.PacketSize = defaultPacketSize
	}
	if o.Attempts <= 0 {
		o.Attempts = defaultAttempts
	}
}
