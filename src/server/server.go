package main

import (
	"fmt"
	"log"
	"net"
	"time"

	pb "github.com/pahanini/go-grpc-bidirectional-streaming-example/src/proto"

	"google.golang.org/grpc"
)

type message struct {
	topic   string
	content string
}

type server struct {
	pb.UnimplementedTopicServiceServer
	messageCh chan message
}

var (
	messages = []message{
		{"foo", "foo1"},
		{"foo", "foo2"},
		{"foo", "foo3"},
		{"bar", "barA"},
		{"bar", "barZ"},
	}
)

func generateMessage(ch chan message) {
	ind := 0
	for {
		fmt.Printf("Generating message %v\n", messages[ind])
		ch <- messages[ind]
		ind = (ind + 1) % len(messages)
		time.Sleep(time.Millisecond * 500)
	}
}

func (s server) Report(srv pb.TopicService_ReportServer) error {
	log.Println("start new server")
	ctx := srv.Context()
	req, err := srv.Recv()
	if err != nil {
		log.Printf("receive error %v", err)
		return err
	}
	log.Printf("Received request for topic %v", req.Topic)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case m := <-s.messageCh:
			if m.topic != req.Topic {
				continue
			}
			resp := pb.Response{Message: m.content}
			if err := srv.Send(&resp); err != nil {
				log.Printf("send error %v", err)
			}
			log.Printf("send message: %v", "bar")
		}
	}
}

func main() {
	log.Printf("Starting the server at port localhost:50005.")
	lis, err := net.Listen("tcp", ":50005")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := &server{
		messageCh: make(chan message),
	}
	go generateMessage(srv.messageCh)
	grpcServer := grpc.NewServer()
	pb.RegisterTopicServiceServer(grpcServer, srv)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
