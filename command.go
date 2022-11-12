package grpctl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/joshcarp/grpctl/internal/grpc"
	"google.golang.org/grpc/metadata"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"

	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// CommandOption are options to customize the grpctl cobra command.
type CommandOption func(*cobra.Command) error

func WithStdin(stdin io.Reader) func(cmd *cobra.Command) error {
	return func(cmd *cobra.Command) error {
		cmd.SetIn(stdin)
		return nil
	}
}

// BuildCommand builds a grpctl command from a list of GrpctlOption.
func BuildCommand(cmd *cobra.Command, opts ...CommandOption) error {
	for _, f := range opts {
		err := f(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func persistentFlags(cmd *cobra.Command, defaultHosts ...string) error {
	var addr, cfgFile, defaultHost, protocol string
	var http1enabled bool
	cmd.PersistentFlags().BoolVar(&http1enabled, "http1", false, "use http1.1 instead of http2")
	cmd.PersistentFlags().StringVarP(&protocol, "protocol", "p", "grpc", "protocol to use: [connect, grpc, grpcweb]")
	err := cmd.RegisterFlagCompletionFunc("protocol", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"grpc", "connect", "grpcweb"}, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		return err
	}
	if len(defaultHosts) > 0 {
		defaultHost = defaultHosts[0]
	}
	cmd.PersistentFlags().StringVarP(&addr, "address", "a", defaultHost, "Address in form 'scheme://host:port'")
	if len(defaultHosts) > 0 {
		err = cmd.RegisterFlagCompletionFunc("address", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return defaultHosts, cobra.ShellCompDirectiveNoFileComp
		})
	}

	if err != nil {
		return err
	}
	cmd.PersistentFlags().StringArrayP("header", "H", []string{}, "Header in form 'key: value'")
	err = cmd.RegisterFlagCompletionFunc("header", cobra.NoFileCompletions)
	if err != nil {
		return err
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.grpctl.yaml)")
	cmd.PersistentFlags().Lookup("config").Hidden = true
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
	seen := map[string]bool{}
	for _, service := range descriptors.ServicesFromFileDescriptor(methods) {
		command := descriptors.Command(service)
		if seen[command] {
			return fmt.Errorf("duplicate service name: %s in %s", command, methods.Name())
		}
		seen[command] = true
		err := CommandFromServiceDescriptor(cmd, service)
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
	command := descriptors.Command(service)
	serviceCmd := cobra.Command{
		Use:   command,
		Short: fmt.Sprintf("%s as defined in %s", command, service.ParentFile().Path()),
	}
	for _, method := range descriptors.MethodsFromServiceDescriptor(service) {
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
func CommandFromMethodDescriptor(cmd *cobra.Command, method protoreflect.MethodDescriptor) error {
	dataMap := make(descriptors.DataMap)
	for fieldNum := 0; fieldNum < method.Input().Fields().Len(); fieldNum++ {
		field := method.Input().Fields().Get(fieldNum)
		jsonName := field.JSONName()
		field.Default()
		field.Kind()
		dataMap[jsonName] = &descriptors.DataValue{Kind: field.Kind(), Value: field.Default().Interface(), Proto: true}
	}
	var inputData, data string
	methodCmdName := descriptors.Command(method)
	methodCmd := cobra.Command{
		Use:   methodCmdName,
		Short: fmt.Sprintf("%s as defined in %s", methodCmdName, method.ParentFile().Path()),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cmd.Root().SetContext(context.WithValue(cmd.Root().Context(), methodDescriptorKey{}, method))
			return recusiveParentPreRun(cmd.Parent(), args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			protocol, err := cmd.Flags().GetString("protocol")
			if err != nil {
				return err
			}
			http1, err := cmd.Flags().GetBool("http1")
			if err != nil {
				return err
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
			for _, header := range headers {
				keyval := strings.Split(header, ":")
				if len(keyval) != 2 {
					return fmt.Errorf("headers need to be in form -H=Foo:Bar")
				}
				cmd.Root().SetContext(metadata.AppendToOutgoingContext(cmd.Root().Context(), keyval[0], strings.TrimLeft(keyval[1], " ")))
			}
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
			if method.IsStreamingClient() || method.IsStreamingServer() {
				return handleStreaming(cmd, method, addr, protocol, http1)
			}
			return handleUnary(cmd, addr, method, inputData, protocol, http1)
		},
	}
	methodCmd.Flags().StringVar(&data, "json-data", "", "JSON data input that will be used as a request")
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

func handleUnary(cmd *cobra.Command, addr string, method protoreflect.MethodDescriptor, inputData string, protocol string, http1 bool) error {
	marshallerm, err := grpc.CallUnary(cmd.Root().Context(), addr, method, []byte(inputData), protocol, http1)
	if err != nil {
		return err
	}
	res := string(marshallerm)
	if err != nil {
		_, err = cmd.OutOrStderr().Write([]byte(err.Error()))
		if err != nil {
			return err
		}
		return err
	}
	_, err = cmd.OutOrStdout().Write([]byte(res))
	return err
}

func handleStreaming(cmd *cobra.Command, method protoreflect.MethodDescriptor, addr, protocol string, http1 bool) (err error) {
	inputJSON, outputJSON := make(chan []byte), make(chan []byte)
	go func() {
		reterr := grpc.CallStreaming(cmd.Root().Context(), addr, method, protocol, http1, inputJSON, outputJSON)
		if reterr != nil {
			err = reterr
			return
		}
	}()
	b, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return err
	}
	msgArr := make([]map[string]any, 0)
	if err := json.Unmarshal(b, &msgArr); err != nil {
		return err
	}
	for _, msg := range msgArr {
		byteMsg, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		inputJSON <- byteMsg
	}
	close(inputJSON)
	for marshallerm := range outputJSON {
		res := string(marshallerm)
		if err != nil {
			_, err = cmd.OutOrStderr().Write([]byte(err.Error()))
			if err != nil {
				return err
			}
			return err
		}
		_, err = cmd.OutOrStdout().Write([]byte(res))
		if err != nil {
			return err
		}
	}
	return nil
}
