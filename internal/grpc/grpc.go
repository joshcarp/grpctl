package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/bufbuild/connect-go"
	reflectconnectv1 "github.com/joshcarp/grpctl/internal/reflection/gen/go/v1/grpc_reflection_v1alphaconnect"
	reflectconnect "github.com/joshcarp/grpctl/internal/reflection/gen/go/v1alpha1/grpc_reflection_v1alphaconnect"
	"golang.org/x/net/http2"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func plaintext(url string) bool {
	return strings.Contains(url, "http://")
}

func client(enablehttp1 bool, plaintext bool) *http.Client {
	if enablehttp1 {
		return http.DefaultClient
	}
	return &http.Client{
		Transport: &http2.Transport{
			AllowHTTP: true,
			DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
				if plaintext {
					return net.Dial(network, addr)
				}
				return tls.Dial(network, addr, cfg)
			},
		},
	}
}

func Reflect(ctx context.Context, baseurl string) (*descriptorpb.FileDescriptorSet, error) {
	fdset, err := ReflectV1alpha1(ctx, baseurl)
	if connect.CodeOf(err) == connect.CodeUnimplemented {
		return ReflectV1(ctx, baseurl)
	}
	return fdset, err
}

// nolint: dupl
func ReflectV1alpha1(ctx context.Context, baseurl string) (*descriptorpb.FileDescriptorSet, error) {
	client := reflectconnect.NewServerReflectionClient(client(false, plaintext(baseurl)), baseurl, connect.WithGRPC())
	stream := client.ServerReflectionInfo(ctx)
	req := &reflectpb.ServerReflectionRequest{MessageRequest: &reflectpb.ServerReflectionRequest_ListServices{}}
	if err := stream.Send(req); err != nil {
		return nil, err
	}
	resp, err := stream.Receive()
	if err != nil {
		return nil, err
	}
	listResp := resp.GetListServicesResponse()
	if listResp == nil {
		return nil, fmt.Errorf("can't list services")
	}
	fds := &descriptorpb.FileDescriptorSet{}
	seen := make(map[string]bool)
	for _, service := range listResp.GetService() {
		req = &reflectpb.ServerReflectionRequest{
			MessageRequest: &reflectpb.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: service.GetName(),
			},
		}
		if err = stream.Send(req); err != nil {
			return nil, err
		}
		resp, err = stream.Receive()
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

// nolint: dupl
func ReflectV1(ctx context.Context, baseurl string) (*descriptorpb.FileDescriptorSet, error) {
	client := reflectconnectv1.NewServerReflectionClient(client(false, plaintext(baseurl)), baseurl, connect.WithGRPC())
	stream := client.ServerReflectionInfo(ctx)
	req := &reflectpb.ServerReflectionRequest{MessageRequest: &reflectpb.ServerReflectionRequest_ListServices{}}
	if err := stream.Send(req); err != nil {
		return nil, err
	}
	resp, err := stream.Receive()
	if err != nil {
		return nil, err
	}
	listResp := resp.GetListServicesResponse()
	if listResp == nil {
		return nil, fmt.Errorf("can't list services")
	}
	fds := &descriptorpb.FileDescriptorSet{}
	seen := make(map[string]bool)
	for _, service := range listResp.GetService() {
		req = &reflectpb.ServerReflectionRequest{
			MessageRequest: &reflectpb.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: service.GetName(),
			},
		}
		if err = stream.Send(req); err != nil {
			return nil, err
		}
		resp, err = stream.Receive()
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
