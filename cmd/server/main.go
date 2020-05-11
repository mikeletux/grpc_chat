package main

import (
	"github.com/mikeletux/grpc_chat/server"
	"log"
)

func main() {
	config := server.Config{
		Addr:     "localhost",
		Port:     1234,
		Protocol: "tcp",
	}

	srv := server.NewGRPCServer(config)
	log.Fatal(srv.Serve())
}
