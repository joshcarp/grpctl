package grpctl

import (
	"context"
	"fmt"
	"strings"

	"github.com/joshcarp/grpctl/internal/grpc"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"

	"google.golang.org/grpc/metadata"

	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// GrptlOption are options to customize the grpctl cobra command.
type GrptlOption func(*cobra.Command) error

// RunCommand runs the command with a context. It is important that the cobra command is run through this
// and not through a raw cmd.Execute() because it needs a custom context to be added by the grpctl library.
func RunCommand(cmd *cobra.Command, ctx context.Context) error {
	customCtx := customContext{
		ctx: &ctx,
	}
	return cmd.ExecuteContext(context.WithValue(context.Background(), configKey{}, &customCtx))
}

// BuildCommand builds a grpctl command from a list of GrpctlOption.
func BuildCommand(cmd *cobra.Command, opts ...GrptlOption) error {
	for _, f := range opts {
		err := f(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func persistentFlags(cmd *cobra.Command, defaultHosts ...string) error {
	var plaintext bool
	var addr string
	var cfgFile string
	var defaultHost string
	cmd.PersistentFlags().BoolVarP(&plaintext, "plaintext", "p", false, "plaintext")
	err := cmd.RegisterFlagCompletionFunc("plaintext", cobra.NoFileCompletions)
	if err != nil {
		return err
	}
	if len(defaultHosts) > 0 {
		defaultHost = defaultHosts[0]
	}
	cmd.PersistentFlags().StringVarP(&addr, "address", "a", defaultHost, "address")
	if len(defaultHosts) > 0 {
		err = cmd.RegisterFlagCompletionFunc("address", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return defaultHosts, cobra.ShellCompDirectiveNoFileComp
		})
	}

	if err != nil {
		return err
	}
	cmd.PersistentFlags().StringArrayP("header", "H", []string{}, "")
	err = cmd.RegisterFlagCompletionFunc("header", cobra.NoFileCompletions)
	if err != nil {
		return err
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.grpctl.yaml)")
	return nil
}

// CommandFromFileDescriptors adds commands to cmd from FileDescriptors.
func CommandFromFileDescriptors(cmd *cobra.Command, descriptors ...protoreflect.FileDescriptor) error {
	for _, desc := range descriptors {
		err := CommandFromFileDescriptor(cmd, desc)
		if err != nil {
			return err
		}
	}
	return nil
}

// CommandFromFileDescriptor adds commands to cmd from a single FileDescriptor.
func CommandFromFileDescriptor(cmd *cobra.Command, methods protoreflect.FileDescriptor) error {
	for _, service := range descriptors.NewFileDescriptor(methods).Services() {
		err := CommandFromServiceDescriptor(cmd, service.ServiceDescriptor)
		if err != nil {
			return err
		}
	}
	return nil
}

// CommandFromServiceDescriptor adds commands to cmd from a ServiceDescriptor.
// Commands added through this will have two levels: the ServiceDescriptor name as level 1 commands
// And the MethodDescriptors as level 2 commands.
func CommandFromServiceDescriptor(cmd *cobra.Command, service protoreflect.ServiceDescriptor) error {
	serviceDesc := descriptors.NewServiceDescriptor(service)
	serviceCmd := cobra.Command{
		Use:   serviceDesc.Command(),
		Short: fmt.Sprintf("%s as defined in %s", serviceDesc.Command(), service.ParentFile().Path()),
	}
	for _, method := range serviceDesc.Methods() {
		err := CommandFromMethodDescriptor(&serviceCmd, method)
		if err != nil {
			return err
		}
	}
	cmd.AddCommand(&serviceCmd)
	defaulthost := proto.GetExtension(service.Options(), annotations.E_DefaultHost)
	serviceCmd.Parent().ResetFlags()
	if defaulthost != "" {
		return persistentFlags(serviceCmd.Parent(), fmt.Sprintf("%v:443", defaulthost))
	}
	return persistentFlags(serviceCmd.Parent())
}

// CommandFromMethodDescriptor adds commands to cmd from a MethodDescriptor.
// Commands added through this will have one level from the MethodDescriptors name.
func CommandFromMethodDescriptor(cmd *cobra.Command, method descriptors.MethodDescriptor) error {
	dataMap := make(descriptors.DataMap)
	if method.IsStreamingClient() || method.IsStreamingServer() {
		return nil
	}
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			custCtx, ctx, ok := getContext(cmd)
			if !ok {
				return nil
			}
			custCtx.setContext(context.WithValue(ctx, methodDescriptorKey{}, method))
			return recusiveParentPreRun(cmd.Parent(), args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_, ctx, ok := getContext(cmd)
			if !ok {
				ctx = cmd.Context()
			}
			headers, err := cmd.Flags().GetStringArray("header")
			if err != nil {
				return err
			}
			addr, err := cmd.Flags().GetString("address")
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
			for _, header := range headers {
				keyval := strings.Split(header, ":")
				if len(keyval) != 2 {
					return fmt.Errorf("headers need to be in form -H=Foo:Bar")
				}
				ctx = metadata.AppendToOutgoingContext(ctx, keyval[0], strings.TrimLeft(keyval[1], " "))
			}
			conn, err := grpc.Setup(ctx, plaintext, addr)
			if err != nil {
				return err
			}
			switch data {
			case "":
				b, err := dataMap.ToJSON()
				if err != nil {
					return err
				}
				inputData = string(b)
			default:
				inputData = data
			}
			res, err := grpc.CallAPI(ctx, conn, method, []byte(inputData))
			if err != nil {
				_, err = cmd.OutOrStderr().Write([]byte(err.Error()))
				if err != nil {
					return err
				}
				return err
			}
			_, err = cmd.OutOrStdout().Write([]byte(res))
			return err
		},
	}
	methodCmd.Flags().StringVar(&data, "json-data", "", "")
	defaults, templ := descriptors.MakeJSONTemplate(method.Input())
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
	methodCmd.ValidArgsFunction = cobra.NoFileCompletions
	cmd.AddCommand(&methodCmd)
	return nil
}
