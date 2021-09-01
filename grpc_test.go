package grpctl

import (
	"context"
	"reflect"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestCallAPI(t *testing.T) {
	type args struct {
		ctx  context.Context
		cc   *grpc.ClientConn
		call protoreflect.MethodDescriptor
		data string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CallAPI(tt.args.ctx, tt.args.cc, tt.args.call, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CallAPI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CallAPI() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reflect(t *testing.T) {
	type args struct {
		conn *grpc.ClientConn
	}
	tests := []struct {
		name    string
		args    args
		want    *descriptorpb.FileDescriptorSet
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grpcreflect(tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("grpcreflect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("grpcreflect() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setup(t *testing.T) {
	type args struct {
		ctx       context.Context
		plaintext bool
		targetURL string
	}
	tests := []struct {
		name    string
		args    args
		want    *grpc.ClientConn
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := setup(tt.args.ctx, tt.args.plaintext, tt.args.targetURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("setup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setup() got = %v, want %v", got, tt.want)
			}
		})
	}
}
