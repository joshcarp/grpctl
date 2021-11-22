package descriptors

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type FileDescriptor struct {
	protoreflect.FileDescriptor
}

func NewFileDescriptor(f protoreflect.FileDescriptor) FileDescriptor {
	return FileDescriptor{FileDescriptor: f}
}

type ServiceDescriptor struct {
	protoreflect.ServiceDescriptor
}

func NewServiceDescriptor(f protoreflect.ServiceDescriptor) ServiceDescriptor {
	return ServiceDescriptor{ServiceDescriptor: f}
}

type MethodDescriptor struct {
	protoreflect.MethodDescriptor
}

func (c FileDescriptor) Services() []ServiceDescriptor {
	var objs []ServiceDescriptor
	for i := 0; i < c.FileDescriptor.Services().Len(); i++ {
		service := c.FileDescriptor.Services().Get(i)
		newService := ServiceDescriptor{ServiceDescriptor: service}
		objs = append(objs, newService)
	}
	return objs
}

func (c ServiceDescriptor) Methods() []MethodDescriptor {
	var objs []MethodDescriptor
	for j := 0; j < c.ServiceDescriptor.Methods().Len(); j++ {
		method := c.ServiceDescriptor.Methods().Get(j)
		objs = append(objs, MethodDescriptor{MethodDescriptor: method})
	}
	return objs
}

func (c ServiceDescriptor) Command() string {
	return string(c.Name())
}

func (c MethodDescriptor) Command() string {
	return string(c.Name())
}

func FullMethod(c protoreflect.MethodDescriptor) string {
	return fmt.Sprintf("/%s/%s", c.Parent().FullName(), c.Name())
}
