package grpc

import (
	"context"
	"errors"
	"io"

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

func CallUnary(ctx context.Context, addr string, method protoreflect.MethodDescriptor, inputData []byte, protocol string, http1 bool) ([]byte, error) {
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
	client := connect.NewClient[emptypb.Empty, emptypb.Empty](client(http1), fqnAddr, clientOpts...)
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

func ParseMessage(inputJSON []byte, messageDesc protoreflect.MessageDescriptor) (*emptypb.Empty, error) {
	dynamicRequest := dynamicpb.NewMessage(messageDesc)
	err := protojson.Unmarshal(inputJSON, dynamicRequest)
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
	return request, nil
}

func Send(inputJSON chan []byte, messageDescriptor protoreflect.MessageDescriptor, f func(*emptypb.Empty) error) error {
	for inputs := range inputJSON {
		request, err := ParseMessage(inputs, messageDescriptor)
		if err != nil {
			return err
		}
		err = f(request)
		if err != nil {
			return err
		}
	}
	return nil
}

func Receive(outputJSON chan []byte, method protoreflect.MethodDescriptor, f func() (*emptypb.Empty, error)) error {
	defer close(outputJSON)
	for {
		msg, err := f()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if msg == nil {
			break
		}
		responseBytes, err := proto.Marshal(msg)
		if err != nil {
			return nil
		}
		dynamicResponse := dynamicpb.NewMessage(method.Output())
		if err := proto.Unmarshal(responseBytes, dynamicResponse); err != nil {
			return err
		}
		reg, err := registry(method)
		if err != nil {
			return err
		}
		b, err := protojson.MarshalOptions{Resolver: &reg, Multiline: true, Indent: " "}.Marshal(dynamicResponse)
		if err != nil {
			return err
		}
		outputJSON <- b
	}
	return nil
}

func CallStreaming(ctx context.Context, addr string, method protoreflect.MethodDescriptor, protocol string, http1 bool, inputJSON, outputJSON chan []byte) error {
	client := getClient(addr, method, protocol, http1)
	if method.IsStreamingClient() && method.IsStreamingServer() { //nolint:gocritic
		stream := client.CallBidiStream(ctx)
		if err := Send(inputJSON, method.Input(), stream.Send); err != nil {
			return err
		}
		if err := Receive(outputJSON, method, stream.Receive); err != nil {
			return err
		}
	} else if method.IsStreamingClient() {
		stream := client.CallClientStream(ctx)
		if err := Send(inputJSON, method.Input(), stream.Send); err != nil {
			return err
		}
		err := Receive(outputJSON, method, func() (*emptypb.Empty, error) {
			resp, err := stream.CloseAndReceive()
			if err != nil {
				return nil, err
			}
			return resp.Msg, err
		})
		if err != nil {
			return err
		}
	} else if method.IsStreamingServer() {
		req, err := ParseMessage(<-inputJSON, method.Input())
		if err != nil {
			return err
		}
		stream, err := client.CallServerStream(ctx, connect.NewRequest(req))
		if err != nil {
			return err
		}
		err = Receive(outputJSON, method, func() (*emptypb.Empty, error) {
			if stream.Receive() {
				return stream.Msg(), nil
			}
			return nil, nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func registry(method protoreflect.MethodDescriptor) (protoregistry.Types, error) {
	var registry protoregistry.Types
	if err := registry.RegisterMessage(dynamicpb.NewMessageType(method.Output())); err != nil {
		return protoregistry.Types{}, err
	}
	if err := registry.RegisterMessage(dynamicpb.NewMessageType(method.Input())); err != nil {
		return protoregistry.Types{}, err
	}
	return registry, nil
}

func getClient(addr string, method protoreflect.MethodDescriptor, protocol string, http1 bool) *connect.Client[emptypb.Empty, emptypb.Empty] {
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
	return connect.NewClient[emptypb.Empty, emptypb.Empty](client(http1), fqnAddr, clientOpts...)
}
