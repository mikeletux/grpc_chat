package server

import (
	"fmt"
	"github.com/mikeletux/grpc_chat/proto"
	"golang.org/x/net/context"
	"log"
)

type ChatServer struct{}

func (c *ChatServer) Connect(beginMessage *proto.BeginMessage, stream proto.Chat_ConnectServer) (err error) {
	if _, ok := chat.clientList[beginMessage.GetName()]; ok {
		return fmt.Errorf("user already exists, please choose another one")
	}
	//Insert client into map
	comm := make(chan string)
	done := make(chan bool)
	chat.mux.Lock()
	chat.clientList[beginMessage.GetName()] = clientNode{true, beginMessage.GetName(), comm}
	log.Printf("connected user %v", beginMessage.GetName())
	chat.mux.Unlock()
	go sendToWorker(&clientNode{true, beginMessage.GetName(), comm}, done, stream)
	<-done

	return nil
}

func (c *ChatServer) SendMessage(ctx context.Context, chatMessage *proto.ChatMessage) (status *proto.Status, err error) {
	chat.mux.Lock()
	log.Printf("Message received from [%v] - Text [%v]", chatMessage.GetNameFrom(), chatMessage.GetText())
	defer chat.mux.Unlock()
	for _, v := range chat.clientList {
		v.Channel <- chatMessage.GetText()
	}
	return &proto.Status{Ok: true, ErrorMessage: ""}, nil
}

func sendToWorker(client *clientNode, done chan bool, stream proto.Chat_ConnectServer) {
	for {
		msg := <-client.Channel
		if err := stream.Send(&proto.ChatMessage{NameFrom: client.Username, Text: msg}); err != nil {
			log.Printf("%v\n", err)
			done <- true
		}
	}
}
