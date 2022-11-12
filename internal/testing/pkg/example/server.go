package example

import (
	"context"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc/reflection"

	"github.com/googleapis/gax-go/v2"
	"github.com/joshcarp/grpctl/internal/testing/proto/examplepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type FooServer struct {
	examplepb.UnimplementedFooAPIServer
}

type Logger func(format string, args ...interface{})

func (f FooServer) Hello(ctx context.Context, example *examplepb.ExampleRequest) (*examplepb.ExampleResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	return &examplepb.ExampleResponse{
		Message: fmt.Sprintf("Incoming Message: %s \n Metadata: %s", example.Message, md),
	}, nil
}

type BarServer struct {
	examplepb.UnimplementedBarAPIServer
}

func (f BarServer) ListBars(ctx context.Context, example *examplepb.BarRequest) (*examplepb.BarResponse, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	return &examplepb.BarResponse{
		Message: fmt.Sprintf("Incoming Message: %s \n Metadata: %s", example.Message, md),
	}, nil
}

func ServeLis(ctx context.Context, log Logger, ln net.Listener, r ...func(*grpc.Server)) (err error) {
	srv := grpc.NewServer()
	for _, rr := range r {
		rr(srv)
		reflection.Register(srv)
	}
	go func() {
		err := srv.Serve(ln)
		log("error serving: %v", err)
	}()
	go func() {
		<-ctx.Done()
		srv.Stop()
		err := ln.Close()
		log("error closing: %v", err)
	}()

	bo := gax.Backoff{
		Initial:    time.Second,
		Multiplier: 2,
		Max:        10 * time.Second,
	}
	for {
		_, err := setup(context.Background(), true, fmt.Sprintf("localhost:%d", ln.Addr().(*net.TCPAddr).Port))
		if err != nil {
			if err := gax.Sleep(ctx, bo.Pause()); err != nil {
				return err
			}
			continue
		}
		return nil
	}
}

func ServeRand(ctx context.Context, r ...func(*grpc.Server)) (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	tcpAddr, _ := ln.Addr().(*net.TCPAddr)
	return tcpAddr.Port, ServeLis(ctx, log.Printf, ln, r...)
}

func setup(ctx context.Context, plaintext bool, targetURL string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(), //nolint
	}
	if !plaintext {
		cp, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		opts = []grpc.DialOption{
			grpc.WithBlock(),
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(cp, "")),
		}
	}
	cc, err := grpc.DialContext(ctx, targetURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("%v: failed to connect to server", err)
	}
	return cc, nil
}
