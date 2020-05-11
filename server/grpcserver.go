package server

import (
	"fmt"
	"github.com/mikeletux/grpc_chat/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"sync"
)

var chat Chat

//List of connected clients
type Chat struct {
	mux        sync.Mutex
	clientList map[string]clientNode //they key of the map will be the user's name
}

type clientNode struct {
	Availabe bool
	Username string
	Channel  chan string
}

type GRPCServer struct {
	config Config
}

func NewGRPCServer(config Config) Server {
	chat.clientList = make(map[string]clientNode)
	return &GRPCServer{config}
}

func (s *GRPCServer) Serve() (err error) {
	ln, err := net.Listen(s.config.Protocol, fmt.Sprintf("%v:%v", s.config.Addr, s.config.Port))
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("couldn't connect to %v:%v", s.config.Addr, s.config.Port))
	}
	log.Printf("Server listening on %v:%v", s.config.Addr, s.config.Port)
	serv := grpc.NewServer()
	chatServer := ChatServer{}
	proto.RegisterChatServer(serv, &chatServer)
	if err = serv.Serve(ln); err != nil {
		return fmt.Errorf(fmt.Sprintf("there was an error when serving from gRPC server"))
	}
	return nil
}
