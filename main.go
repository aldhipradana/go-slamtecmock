package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", defaultPort, "HTTP port to listen on")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	manager := NewRobotManager()
	router := manager.BuildRouter()

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("go-slamtecmock listening on %s", addr)
	log.Printf("Dashboard → http://localhost%s", addr)
	log.Printf("Robot #1 API → http://localhost%s/robot/1/api/core/system/v1/power/status", addr)
	log.Printf("Compat alias → http://localhost%s/api/core/system/v1/power/status", addr)
	log.Fatal(http.ListenAndServe(addr, router))
}

