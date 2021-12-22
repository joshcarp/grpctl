package main

import (
	"log"
	"net"

	"github.com/joshcarp/grpctl/internal/testing/pkg/example"
	"github.com/joshcarp/grpctl/internal/testing/proto/examplepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}
	srv := grpc.NewServer()
	examplepb.RegisterFooAPIServer(srv, example.FooServer{})
	examplepb.RegisterBarAPIServer(srv, example.BarServer{})
	reflection.Register(srv)
	err = srv.Serve(ln)
	if err != nil {
		log.Fatal(err)
	}
}
