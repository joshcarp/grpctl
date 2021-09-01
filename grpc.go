package grpctl

import (
	"context"
	"crypto/x509"
	"fmt"

	"github.com/joshcarp/grpctl/internal/descriptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

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

func reflect(conn *grpc.ClientConn) (*descriptorpb.FileDescriptorSet, error) {
	client := reflectpb.NewServerReflectionClient(conn)
	methodClient, err := client.ServerReflectionInfo(context.Background())
	if err != nil {
		return nil, err
	}
	req := &reflectpb.ServerReflectionRequest{MessageRequest: &reflectpb.ServerReflectionRequest_ListServices{}}
	if err = methodClient.Send(req); err != nil {
		return nil, err
	}
	resp, err := methodClient.Recv()
	if err != nil {
		return nil, err
	}
	listResp := resp.GetListServicesResponse()
	if listResp == nil {
		return nil, fmt.Errorf("can't list services")
	}
	fds := &descriptorpb.FileDescriptorSet{}
	seen := map[string]bool{}
	for _, service := range listResp.GetService() {
		req = &reflectpb.ServerReflectionRequest{
			MessageRequest: &reflectpb.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: service.GetName(),
			},
		}
		if err = methodClient.Send(req); err != nil {
			return nil, err
		}
		resp, err = methodClient.Recv()
		if err != nil {
			return nil, fmt.Errorf("error listing methods on '%s': %w", service, err)
		}
		fdResp := resp.GetFileDescriptorResponse()
		for _, f := range fdResp.GetFileDescriptorProto() {
			a := &descriptorpb.FileDescriptorProto{}
			if err = proto.Unmarshal(f, a); err != nil {
				return nil, err
			}
			if seen[a.GetName()] {
				continue
			}
			seen[a.GetName()] = true
			fds.File = append(fds.File, a)
		}
	}
	return fds, nil
}

func ConvertToProtoReflectDesc(fds *descriptorpb.FileDescriptorSet) ([]protoreflect.FileDescriptor, error) {
	files, err := protodesc.NewFiles(fds)
	if err != nil {
		return nil, err
	}
	var reflectds []protoreflect.FileDescriptor
	for _, fd := range fds.File {
		reflectfile, err := protodesc.NewFile(fd, files)
		if err != nil {
			return nil, err
		}
		reflectds = append(reflectds, reflectfile)
	}
	return reflectds, nil
}

func CallAPI(ctx context.Context, cc *grpc.ClientConn, call protoreflect.MethodDescriptor, data string) (string, error) {
	fullmethod := descriptors.FullMethod(call)
	request := dynamicpb.NewMessage(call.Input())
	err := protojson.Unmarshal([]byte(data), request)
	if err != nil {
		return "", err
	}
	fmt.Println(request.String())
	response := dynamicpb.NewMessage(call.Output())
	err = cc.Invoke(ctx, fullmethod, request, response)
	if err != nil {
		return "", err
	}
	marshallerm, err := protojson.Marshal(response)
	return string(marshallerm), err
}
