package grpc

import (
	"context"

	"github.com/bufbuild/connect-go"
	"github.com/joshcarp/grpctl/internal/descriptors"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

func CallUnary(ctx context.Context, addr string, method protoreflect.MethodDescriptor, inputData []byte, protocol string) ([]byte, error) {
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
	var clientOpts []connect.ClientOption
	switch protocol {
	case "grpc":
		clientOpts = append(clientOpts, connect.WithGRPC())
	case "grpcweb":
		clientOpts = append(clientOpts, connect.WithGRPCWeb())
	case "connect":
	default:
	}
	client := connect.NewClient[emptypb.Empty, emptypb.Empty](client(), fqnAddr, clientOpts...)
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
