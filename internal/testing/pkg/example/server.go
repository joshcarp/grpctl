package example

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/googleapis/gax-go/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	"github.com/joshcarp/grpctl/internal/testing/proto/examplepb"
)

type FooServer struct {
	examplepb.UnimplementedFooAPIServer
}

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

/* Serve servers a servermock server and blocks until the server is running. Use context.WithCancel to stop the server */
func Serve(ctx context.Context, addr string, r ...func(*grpc.Server)) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return ServeLis(ctx, ln, r...)
}

func ServeLis(ctx context.Context, ln net.Listener, r ...func(*grpc.Server)) error {
	srv := grpc.NewServer()
	for _, rr := range r {
		rr(srv)
		reflection.Register(srv)
	}
	go func() {
		_ = srv.Serve(ln)
	}()
	go func() {
		<-ctx.Done()
		srv.Stop()
		ln.Close()
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
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	return ln.Addr().(*net.TCPAddr).Port, ServeLis(ctx, ln, r...)
}

func setup(ctx context.Context, plaintext bool, targetURL string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithInsecure(),
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
