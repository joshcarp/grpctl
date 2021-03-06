// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package examplepb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// FooAPIClient is the client API for FooAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FooAPIClient interface {
	Hello(ctx context.Context, in *ExampleRequest, opts ...grpc.CallOption) (*ExampleResponse, error)
}

type fooAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewFooAPIClient(cc grpc.ClientConnInterface) FooAPIClient {
	return &fooAPIClient{cc}
}

func (c *fooAPIClient) Hello(ctx context.Context, in *ExampleRequest, opts ...grpc.CallOption) (*ExampleResponse, error) {
	out := new(ExampleResponse)
	err := c.cc.Invoke(ctx, "/example.FooAPI/Hello", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FooAPIServer is the server API for FooAPI service.
// All implementations must embed UnimplementedFooAPIServer
// for forward compatibility
type FooAPIServer interface {
	Hello(context.Context, *ExampleRequest) (*ExampleResponse, error)
	mustEmbedUnimplementedFooAPIServer()
}

// UnimplementedFooAPIServer must be embedded to have forward compatible implementations.
type UnimplementedFooAPIServer struct {
}

func (UnimplementedFooAPIServer) Hello(context.Context, *ExampleRequest) (*ExampleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Hello not implemented")
}
func (UnimplementedFooAPIServer) mustEmbedUnimplementedFooAPIServer() {}

// UnsafeFooAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FooAPIServer will
// result in compilation errors.
type UnsafeFooAPIServer interface {
	mustEmbedUnimplementedFooAPIServer()
}

func RegisterFooAPIServer(s grpc.ServiceRegistrar, srv FooAPIServer) {
	s.RegisterService(&FooAPI_ServiceDesc, srv)
}

func _FooAPI_Hello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExampleRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FooAPIServer).Hello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.FooAPI/Hello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FooAPIServer).Hello(ctx, req.(*ExampleRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// FooAPI_ServiceDesc is the grpc.ServiceDesc for FooAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FooAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "example.FooAPI",
	HandlerType: (*FooAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Hello",
			Handler:    _FooAPI_Hello_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api.proto",
}

// BarAPIClient is the client API for BarAPI service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BarAPIClient interface {
	ListBars(ctx context.Context, in *BarRequest, opts ...grpc.CallOption) (*BarResponse, error)
}

type barAPIClient struct {
	cc grpc.ClientConnInterface
}

func NewBarAPIClient(cc grpc.ClientConnInterface) BarAPIClient {
	return &barAPIClient{cc}
}

func (c *barAPIClient) ListBars(ctx context.Context, in *BarRequest, opts ...grpc.CallOption) (*BarResponse, error) {
	out := new(BarResponse)
	err := c.cc.Invoke(ctx, "/example.BarAPI/ListBars", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BarAPIServer is the server API for BarAPI service.
// All implementations must embed UnimplementedBarAPIServer
// for forward compatibility
type BarAPIServer interface {
	ListBars(context.Context, *BarRequest) (*BarResponse, error)
	mustEmbedUnimplementedBarAPIServer()
}

// UnimplementedBarAPIServer must be embedded to have forward compatible implementations.
type UnimplementedBarAPIServer struct {
}

func (UnimplementedBarAPIServer) ListBars(context.Context, *BarRequest) (*BarResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListBars not implemented")
}
func (UnimplementedBarAPIServer) mustEmbedUnimplementedBarAPIServer() {}

// UnsafeBarAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BarAPIServer will
// result in compilation errors.
type UnsafeBarAPIServer interface {
	mustEmbedUnimplementedBarAPIServer()
}

func RegisterBarAPIServer(s grpc.ServiceRegistrar, srv BarAPIServer) {
	s.RegisterService(&BarAPI_ServiceDesc, srv)
}

func _BarAPI_ListBars_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BarRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BarAPIServer).ListBars(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/example.BarAPI/ListBars",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BarAPIServer).ListBars(ctx, req.(*BarRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BarAPI_ServiceDesc is the grpc.ServiceDesc for BarAPI service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BarAPI_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "example.BarAPI",
	HandlerType: (*BarAPIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListBars",
			Handler:    _BarAPI_ListBars_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api.proto",
}
