package server

import (
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/mikeletux/grpc_chat/proto"
	"golang.org/x/net/context"
	"log"
	"sync"
	"time"
)

type localChatMessage struct {
	timestamp time.Time
	username  string
	text      string
}

type clientNode struct {
	Availabe bool
	Username string
	Channel  chan localChatMessage
}

type chatServer struct {
	mux        sync.Mutex
	clientList map[string]clientNode //they key of the map will be the user's name
}

func NewChatServer() *chatServer {
	return &chatServer{
		clientList: make(map[string]clientNode),
	}
}

func (c *chatServer) Connect(beginMessage *proto.BeginMessage, stream proto.Chat_ConnectServer) (err error) {
	if _, ok := c.clientList[beginMessage.GetName()]; ok {
		return fmt.Errorf("user already exists, please choose another one")
	}
	//Insert client into map
	comm := make(chan localChatMessage)
	done := make(chan bool)
	c.mux.Lock()
	c.clientList[beginMessage.GetName()] = clientNode{true, beginMessage.GetName(), comm}
	log.Printf("connected user %v\n", beginMessage.GetName())
	c.mux.Unlock()
	go sendToWorker(&clientNode{true, beginMessage.GetName(), comm}, done, stream)
	<-done
	//Remove user from chatServer map
	delete(c.clientList, beginMessage.GetName())
	log.Printf("user %v deleted from chat server\n", beginMessage.GetName())
	c.ShowCurrentActiveUsers()
	return nil
}

func (c *chatServer) SendMessage(ctx context.Context, chatMessage *proto.ChatMessage) (status *proto.Status, err error) {
	c.mux.Lock()
	log.Printf("Message received from [%v] - Text [%v]", chatMessage.GetNameFrom(), chatMessage.GetText())
	defer c.mux.Unlock()
	timestamp, err := ptypes.Timestamp(chatMessage.GetTimestamp())
	if err != nil {
		//Return error to user
	}
	msg := localChatMessage{timestamp: timestamp, username: chatMessage.GetNameFrom(), text: chatMessage.GetText()}
	for _, v := range c.clientList {
		v.Channel <- msg //Poor error control. If c.Channel is not available it will block.
	}
	return &proto.Status{Ok: true, ErrorMessage: ""}, nil
}

func (c *chatServer) ShowCurrentActiveUsers() {
	var logString string
	logString += "Users left in the server: "
	for k := range c.clientList {
		logString += fmt.Sprintf("%v |", k)
	}
	log.Printf("%v\n", logString)
}

func sendToWorker(client *clientNode, done chan bool, stream proto.Chat_ConnectServer) {
	for {
		msg := <-client.Channel
		//We create the message
		timestamp, err := ptypes.TimestampProto(msg.timestamp)
		outboundMsg := &proto.ChatMessage{Timestamp: timestamp, NameFrom: msg.username, Text: msg.text}
		if err != nil {
			//Handle timestamp error
		}
		//Send the message to its client
		if err := stream.Send(outboundMsg); err != nil {
			log.Printf("%v\n", err)
			done <- true
		}
	}
}
