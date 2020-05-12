package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/mikeletux/grpc_chat/proto"
	"google.golang.org/grpc"
	"log"
	"os"
	"strings"
	"time"
)

var (
	username string
	host     string
	port     int
)

func main() {
	Init()
	done := make(chan bool)
	conn, err := grpc.Dial(fmt.Sprintf("%v:%v", host, port), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	log.Printf("Connected to %v:%v\n", host, port)
	defer conn.Close()
	client := proto.NewChatClient(conn)

	stream, err := client.Connect(context.Background(), &proto.BeginMessage{Name: username})
	if err != nil {
		panic(err)
	}

	go SenderWorker(client, done)
	go ReceiveFromWorker(stream, done)
	<-done

}

func Init() {
	flag.StringVar(&username, "username", "Annonymous", "Name of the user to be used")
	flag.StringVar(&host, "host", "localhost", "Address of the server to connect")
	flag.IntVar(&port, "port", 1234, "Server port to connect")
	flag.Parse()
}

func SenderWorker(client proto.ChatClient, done chan bool) (err error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		if strings.Compare("exit", text) == 0 {
			log.Println("exiting...")
			done <- true
			break
		}
		//Create send message
		timestamp, err := ptypes.TimestampProto(time.Now())
		if err != nil {
			//Handle the error
		}
		msg := &proto.ChatMessage{Timestamp: timestamp, NameFrom: username, Text: text}
		client.SendMessage(context.Background(), msg)
	}
	return nil
}

func ReceiveFromWorker(stream proto.Chat_ConnectClient, done chan bool) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Println(err)
			done <- true
			break
		}
		timestamp, err := ptypes.Timestamp(msg.GetTimestamp())
		if err != nil {
			//Handle issue
		}
		fmt.Printf("[%v](%v) - %v\n", timestamp, msg.GetNameFrom(), msg.GetText())
	}
}
