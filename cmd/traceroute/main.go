package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/visonhuo/mykit/pkg/traceroute"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: traceroute host1 host2")
		return
	}

	srv, err := traceroute.NewServer(traceroute.Config{})
	if err != nil {
		log.Fatalf("Create traceroute server failed: %v\n", err)
	}
	defer srv.Shutdown()

	hosts := os.Args[1:]
	results := make(map[string]*traceroute.Future, len(hosts))
	for i := range hosts {
		future, err := srv.Traceroute(context.Background(), hosts[i], traceroute.Options{})
		if err != nil {
			fmt.Println("Invalid host name: ", hosts[i])
			continue
		}
		results[hosts[i]] = future
	}

	for i := range hosts {
		future, ok := results[hosts[i]]
		if !ok {
			continue
		}
		printResult(hosts[i], future)
	}
}

func printResult(host string, future *traceroute.Future) {
	result := future.Result()
	err := future.Error()
	fmt.Printf("traceroute to %v (%v), %v hops max, %v byte packets\n",
		host, result.DstIP.String(), result.Opts.MaxHop, result.Opts.PacketSize)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	sort.Slice(result.Hops, func(i, j int) bool {
		return result.Hops[i].TTL < result.Hops[j].TTL
	})
	for i := range result.Hops {
		fmt.Printf(" %d", i+1)
		for j := range result.Hops[i].Nodes {
			fmt.Printf("\t%v", result.Hops[i].Nodes[j].IP)
			for k := range result.Hops[i].Nodes[j].RTTs {
				fmt.Printf("\t%.3f ms", float64(result.Hops[i].Nodes[j].RTTs[k].Microseconds())/1000)
			}
			fmt.Println()
		}
	}
}
