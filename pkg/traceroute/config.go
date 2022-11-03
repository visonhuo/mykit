package traceroute

import (
	"log"
	"net"
	"os"
	"time"
)

type Config struct {
	// errorLog specifies an optional logger for errors accepting
	// connections, unexpected behavior from handlers, and
	// underlying FileSystem errors.
	// If nil, logging is done via the log package's standard logger.
	ErrLogger       *log.Logger
	LocalSrcIP      net.IP
	PacketQueueSize int
	DispatchTimeout time.Duration
}

func (c *Config) init() {
	if c.ErrLogger == nil {
		c.ErrLogger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
	}
	if c.LocalSrcIP == nil {
		if localSrcIP, err := localIPv4(); err == nil {
			c.LocalSrcIP = localSrcIP
		} else {
			c.LocalSrcIP = net.IPv4zero
		}
	}
	if c.PacketQueueSize <= 0 {
		c.PacketQueueSize = 32
	}
	if c.DispatchTimeout <= 0 {
		c.DispatchTimeout = 100 * time.Millisecond
	}
}
