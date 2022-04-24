package main

import (
	"context"
	"io"
	"log"

	pb "github.com/pahanini/go-grpc-bidirectional-streaming-example/src/proto"
	flag "github.com/spf13/pflag"

	"google.golang.org/grpc"
)

var (
	topic string
)

func init() {
	flag.StringVar(&topic, "topic", "foo", "topic client subscribe to.")
}

func main() {
	flag.Parse()
	conn, err := grpc.Dial(":50005", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	client := pb.NewTopicServiceClient(conn)
	stream, err := client.Report(context.Background())
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}
	ctx := stream.Context()
	done := make(chan bool)
	req := pb.Request{Topic: topic}
	if err := stream.Send(&req); err != nil {
		log.Fatalf("can not send %v", err)
	}
	// if stream is finished it closes done channel
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			log.Printf("New content %v received", resp.Message)
		}
	}()

	<-ctx.Done()
	if err := ctx.Err(); err != nil {
		log.Println(err)
	}
	close(done)
}
