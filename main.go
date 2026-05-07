package main

import (
	"flag"
	"log"
)

func main() {
	port := flag.Int("port", defaultPort, "HTTP port to listen on")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("Starting go-slamtecmock on port %d", *port)

	state := defaultRobotState()
	_ = NewMockRobot(*port, state)

	// Block forever — the HTTP server runs in goroutines
	select {}
}

