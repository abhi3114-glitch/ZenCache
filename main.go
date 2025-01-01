package main

import (
	"flag"
	"fmt"
	"log"
	"zencache/server"
)

func main() {
	port := flag.Int("port", 6379, "Port to listen on")
	capacity := flag.Int("capacity", 10000, "Maximum number of items in cache (LRU eviction)")
	flag.Parse()

	fmt.Printf("ZenCache v1.0\n")
	fmt.Printf("  Port: %d\n", *port)
	fmt.Printf("  Capacity: %d items\n", *capacity)
	fmt.Println("  Commands: SET, GET, DEL, PING, SUBSCRIBE, PUBLISH, SAVE, REPLICAOF, INFO, QUIT")
	fmt.Println("Starting server...")

	srv := server.NewServerWithCapacity(*port, *capacity)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
