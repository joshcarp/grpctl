package descriptors

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func FullMethod(c protoreflect.MethodDescriptor) string {
	return fmt.Sprintf("/%s/%s", c.Parent().FullName(), c.Name())
}

func Command(descriptor protoreflect.Descriptor) string {
	return string(descriptor.Name())
}

func ServicesFromFileDescriptor(c protoreflect.FileDescriptor) []protoreflect.ServiceDescriptor {
	var objs []protoreflect.ServiceDescriptor
	for i := 0; i < c.Services().Len(); i++ {
		service := c.Services().Get(i)
		objs = append(objs, service)
	}
	return objs
}

func MethodsFromServiceDescriptor(c protoreflect.ServiceDescriptor) []protoreflect.MethodDescriptor {
	var objs []protoreflect.MethodDescriptor
	for j := 0; j < c.Methods().Len(); j++ {
		method := c.Methods().Get(j)
		objs = append(objs, method)
	}
	return objs
}
