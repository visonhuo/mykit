package traceroute

import (
	"net"
	"time"
)

type Result struct {
	DstIP net.IP
	Reach bool
	Hops  []Hop
	Opts  Options
}

type Hop struct {
	TTL   int
	Nodes []Node
}

type Node struct {
	IP   net.IP
	RTTs []time.Duration
}

func (r *Result) aggregate(ttl int, from net.IP, rtt time.Duration) {
	if from.Equal(r.DstIP) {
		r.Reach = true
	}

	for i := range r.Hops {
		if r.Hops[i].TTL == ttl { // exist hop record
			for j := range r.Hops[i].Nodes {
				if r.Hops[i].Nodes[j].IP.Equal(from) { // exist node record
					r.Hops[i].Nodes[j].RTTs = append(r.Hops[i].Nodes[j].RTTs, rtt)
					return
				}
				r.Hops[i].Nodes = append(r.Hops[i].Nodes, Node{
					IP:   from,
					RTTs: []time.Duration{rtt},
				})
				return
			}
		}
	}

	nodes := make([]Node, 0, r.Opts.Attempts)
	nodes = append(nodes, Node{IP: from, RTTs: []time.Duration{rtt}})
	if r.Hops == nil {
		r.Hops = make([]Hop, 0, r.Opts.MaxHop-r.Opts.FirstHop)
	}
	r.Hops = append(r.Hops, Hop{TTL: ttl, Nodes: nodes})
	return
}

type Future struct {
	finish   chan struct{}
	finished bool
	result   Result
	err      error
}

func (f *Future) Error() error {
	<-f.finish
	return f.err
}

func (f *Future) Result() Result {
	<-f.finish
	return f.result
}

func (f *Future) done(result Result, err error) {
	f.result = result
	f.err = err
	f.finished = true
	close(f.finish)
}

func (f *Future) isDone() bool {
	return f.finished
}
