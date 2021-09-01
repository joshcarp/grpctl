package grpctl

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func CommandFromFileDescriptors(config Config, descriptors ...protoreflect.FileDescriptor) []*cobra.Command {
	var cmds []*cobra.Command
	for _, desc := range descriptors {
		cmds = append(cmds, CommandFromFileDescriptor(config, desc)...)
	}
	return cmds
}

func CommandFromFileDescriptor(config Config, methods protoreflect.FileDescriptor) []*cobra.Command {
	var cmds []*cobra.Command
	for _, service := range descriptors.NewFileDescriptor(methods).Services() {
		cmds = append(cmds, CommandFromServiceDescriptor(config, service.ServiceDescriptor))
	}
	return cmds
}

func CommandFromServiceDescriptor(config Config, service protoreflect.ServiceDescriptor) *cobra.Command {
	servicedesc := descriptors.NewServiceDescriptor(service)
	serviceCmd := cobra.Command{
		Use:   servicedesc.Command(),
		ValidArgsFunction: cobra.NoFileCompletions,
		Short: fmt.Sprintf("%s as defined in %s", servicedesc.Command(), service.ParentFile().Path()),
	}
	for _, method := range servicedesc.Methods() {
		methodCmd := CommandFromMethodDescriptor(config, servicedesc, method)
		serviceCmd.AddCommand(&methodCmd)
	}
	return &serviceCmd
}

func CommandFromMethodDescriptor(config Config, service descriptors.ServiceDescriptor, method descriptors.MethodDescriptor) cobra.Command {
	datamap := make(descriptors.DataMap)
	for fieldnum := 0; fieldnum < method.Input().Fields().Len(); fieldnum++ {
		field := method.Input().Fields().Get(fieldnum)
		jsonName := field.JSONName()
		field.Default()
		field.Kind()
		datamap[jsonName] = &descriptors.DataValue{Kind: field.Kind(), Value: field.Default().Interface(), Proto: true}
	}
	var addr, inputdata, data string
	var plaintext bool
	methodcmd := cobra.Command{
		Use:               method.Command(),
		Short:             fmt.Sprintf("%s as defined in %s", method.Command(), method.ParentFile().Path()),
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(cmd *cobra.Command, args []string) error {
			servicecfg, err := config.GetService(service.Command())
			if err == nil {
				if cmd.Flag("plaintext").Changed {
					servicecfg.Plaintext = plaintext
				}
				if addr != "" {
					servicecfg.Addr = addr
				}
			}
			ctx := context.Background()
			user, err := config.GetUser(config.CurrentUser)
			if err == NotFoundError {
				fmt.Println("user not found")
			}
			for key, val := range user.Headers {
				ctx = metadata.AppendToOutgoingContext(ctx, key, val)
			}
			conn, err := setup(ctx, servicecfg.Plaintext, servicecfg.Addr)
			if err != nil {
				return err
			}
			switch data {
			case "":
				b, err := datamap.ToJson()
				if err != nil {
					return err
				}
				inputdata = string(b)
			default:
				inputdata = data
			}
			cobra.CheckErr(err)
			res, err := CallAPI(ctx, conn, method, inputdata)
			fmt.Println(res)
			return err
		},
	}
	methodcmd.Flags().StringVar(&data, "json-data", "", "")
	requiredFlags(&methodcmd, &plaintext, &addr)
	defaults, templ := MakeJsonTemplate(method.Input())
	cobra.CheckErr(methodcmd.RegisterFlagCompletionFunc("json-data", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{templ}, cobra.ShellCompDirectiveDefault
	}))
	for key, val := range datamap {
		methodcmd.Flags().Var(val, key, "")
		cobra.CheckErr(methodcmd.RegisterFlagCompletionFunc(key, func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return []string{fmt.Sprintf("%v", defaults[key])}, cobra.ShellCompDirectiveDefault
		}))
	}
	return methodcmd
}

func requiredFlags(cmd *cobra.Command, plaintext *bool, addr *string) {
	cmd.Flags().BoolVar(plaintext, "plaintext", false, "")
	cobra.CheckErr(cmd.RegisterFlagCompletionFunc("plaintext", cobra.NoFileCompletions))
	cmd.Flags().StringVar(addr, "addr", "", "")
	cobra.CheckErr(cmd.RegisterFlagCompletionFunc("addr", cobra.NoFileCompletions))
}

func flagCompletion(defaultVals descriptors.DataMap, flagstorer descriptors.DataMap, cmd *cobra.Command) {
	for key, val := range defaultVals {
		key := key
		val := val
		flagstorer[key] = &descriptors.DataValue{Value: val.Value, Empty: true}
		cmd.Flags().Var(flagstorer[key], key, "")
		cobra.CheckErr(cmd.RegisterFlagCompletionFunc(key, func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return []string{fmt.Sprintf("%v", val)}, cobra.ShellCompDirectiveDefault
		}))
	}
}