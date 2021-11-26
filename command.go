package grpctl

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"

	"google.golang.org/grpc/metadata"

	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func CommandFromFileDescriptors(cmd *cobra.Command, descriptors ...protoreflect.FileDescriptor) error {
	for _, desc := range descriptors {
		err := CommandFromFileDescriptor(cmd, desc)
		if err != nil {
			return err
		}
	}
	return nil
}

func CommandFromFileDescriptor(cmd *cobra.Command, methods protoreflect.FileDescriptor) error {
	for _, service := range descriptors.NewFileDescriptor(methods).Services() {
		err := CommandFromServiceDescriptor(cmd, service.ServiceDescriptor)
		if err != nil {
			return err
		}
	}
	return nil
}

func CommandFromServiceDescriptor(cmd *cobra.Command, service protoreflect.ServiceDescriptor) error {
	servicedesc := descriptors.NewServiceDescriptor(service)
	serviceCmd := cobra.Command{
		Use:   servicedesc.Command(),
		Short: fmt.Sprintf("%s as defined in %s", servicedesc.Command(), service.ParentFile().Path()),
	}
	for _, method := range servicedesc.Methods() {
		err := CommandFromMethodDescriptor(&serviceCmd, method)
		if err != nil {
			return err
		}
	}
	cmd.AddCommand(&serviceCmd)
	defaulthost := proto.GetExtension(service.Options(), annotations.E_DefaultHost)
	serviceCmd.Parent().ResetFlags()
	err := PersistentFlags(serviceCmd.Parent(), fmt.Sprintf("%v", defaulthost))
	if err != nil {
		return err
	}
	return nil
}

func CommandFromMethodDescriptor(cmd *cobra.Command, method descriptors.MethodDescriptor) error {
	dataMap := make(descriptors.DataMap)
	for fieldNum := 0; fieldNum < method.Input().Fields().Len(); fieldNum++ {
		field := method.Input().Fields().Get(fieldNum)
		jsonName := field.JSONName()
		field.Default()
		field.Kind()
		dataMap[jsonName] = &descriptors.DataValue{Kind: field.Kind(), Value: field.Default().Interface(), Proto: true}
	}
	var inputData, data string
	methodCmd := cobra.Command{
		Use:   method.Command(),
		Short: fmt.Sprintf("%s as defined in %s", method.Command(), method.ParentFile().Path()),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			a, err := cmd.Flags().GetStringArray("header")
			if err != nil {
				return err
			}
			addr, err := cmd.Flags().GetString("addr")
			if err != nil {
				return err
			}
			if addr == "" {
				return nil
			}
			plaintext, err := cmd.Flags().GetBool("plaintext")
			if err != nil {
				return err
			}

			for _, e := range a {
				arr := strings.Split(e, ":")
				if len(arr) != 2 {
					return fmt.Errorf("headers need to be in form -H=Foo:Bar")
				}
				ctx = metadata.AppendToOutgoingContext(ctx, arr[0], strings.TrimLeft(arr[1], " "))
			}

			conn, err := setup(ctx, plaintext, addr)
			if err != nil {
				return err
			}
			switch data {
			case "":
				b, err := dataMap.ToJson()
				if err != nil {
					return err
				}
				inputData = string(b)
			default:
				inputData = data
			}
			res, err := CallAPI(ctx, conn, method, inputData)
			if err != nil {
				_, _ = cmd.OutOrStderr().Write([]byte(err.Error()))
				return err
			}
			_, _ = cmd.OutOrStdout().Write([]byte(res))
			return nil
		},
	}
	methodCmd.Flags().StringVar(&data, "json-data", "", "")
	defaults, templ := MakeJsonTemplate(method.Input())
	err := methodCmd.RegisterFlagCompletionFunc("json-data", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{templ}, cobra.ShellCompDirectiveDefault
	})
	if err != nil {
		return err
	}
	for key, val := range dataMap {
		methodCmd.Flags().Var(val, key, "")
		err := methodCmd.RegisterFlagCompletionFunc(key, func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return []string{fmt.Sprintf("%v", defaults[key])}, cobra.ShellCompDirectiveDefault
		})
		if err != nil {
			return err
		}
	}
	cmd.AddCommand(&methodCmd)
	return nil
}
