package grpc

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/joshcarp/grpctl/internal/descriptors"
	"golang.org/x/net/http2"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

func CallUnary(ctx context.Context, addr string, method protoreflect.MethodDescriptor, inputData []byte) ([]byte, error) {
	dynamicRequest := dynamicpb.NewMessage(method.Input())
	err := protojson.Unmarshal(inputData, dynamicRequest)
	if err != nil {
		return nil, err
	}
	requestBytes, err := proto.Marshal(dynamicRequest)
	if err != nil {
		return nil, err
	}
	request := &emptypb.Empty{}
	if err := proto.Unmarshal(requestBytes, request); err != nil {
		return nil, err
	}
	connectReq := connect.NewRequest(request)
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		for key, val := range md {
			connectReq.Header().Set(key, val[0])
		}
	}
	fqnAddr := addr + descriptors.FullMethod(method)
	client := connect.NewClient[emptypb.Empty, emptypb.Empty](&http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}, fqnAddr, connect.WithGRPC())
	var registry protoregistry.Types
	if err := registry.RegisterMessage(dynamicpb.NewMessageType(method.Output())); err != nil {
		return nil, err
	}
	if err := registry.RegisterMessage(dynamicpb.NewMessageType(method.Input())); err != nil {
		return nil, err
	}
	response, err := client.CallUnary(ctx, connectReq)
	if err != nil {
		return nil, err
	}
	responseBytes, err := proto.Marshal(response.Msg)
	if err != nil {
		return nil, err
	}
	dynamicResponse := dynamicpb.NewMessage(method.Output())
	if err := proto.Unmarshal(responseBytes, dynamicResponse); err != nil {
		return nil, err
	}
	return protojson.MarshalOptions{Resolver: &registry, Multiline: true, Indent: " "}.Marshal(dynamicResponse)
}
