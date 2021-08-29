
package grpctl

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc/metadata"

	"github.com/joshcarp/grpctl/internal/descriptors"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"
)

func ConfigCommands(config Config) *cobra.Command {
	configCmd := cobra.Command{
		Use:   "config",
		Short: "configure options in grpctl",
	}
	setcontext := &cobra.Command{
		Use:   "set-context",
		Short: "set context",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			newctx := args[0]
			var found bool
			for _, e := range config.Contexts {
				if e.Name == newctx {
					found = true
				}
			}
			if !found {
				log.Fatal("Context %s does not exist", newctx)
			}
			config.CurrentContext = newctx
			b, err := yaml.Marshal(config)
			cobra.CheckErr(err)
			err = os.WriteFile(config.ConfigFile, b, os.ModePerm)
			cobra.CheckErr(err)
		},
	}
	configCmd.AddCommand(setcontext)
	return &configCmd
}

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
		datamap[jsonName] = &descriptors.DataValue{Kind: field.Kind(), Value: field.Default().Interface()}
	}
	var addr string
	var plaintext bool
	var plaintextset bool
	var inputdata string
	var data string
	methodcmd := cobra.Command{
		Use:   method.Command(),
		Short: fmt.Sprintf("%s as defined in %s", method.Command(), method.ParentFile().Path()),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr2, plaintext2, ok := config.GetServiceConfig(service.Command())
			if ok {
				if !plaintextset {
					plaintext = plaintext2
				}
				if addr == "" {
					addr = addr2
				}
			}
			ctx := context.Background()
			for _, val := range config.GetCurrentContext().User.Headers {
				ctx = metadata.AppendToOutgoingContext(ctx, val.Key, val.Value)
			}
			conn, err := setup(ctx, plaintext, addr)
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
	requiredFlags(&methodcmd, &plaintext, &plaintextset, &addr)
	defaults, templ := MakeJsonTemplate(method.Input())
	err := methodcmd.RegisterFlagCompletionFunc("json-data", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{templ}, cobra.ShellCompDirectiveDefault
	})
	cobra.CheckErr(err)
	for key, val := range datamap {
		methodcmd.Flags().Var(val, key, "")
		methodcmd.RegisterFlagCompletionFunc(key, func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return []string{fmt.Sprintf("%v", defaults[key])}, cobra.ShellCompDirectiveDefault
		})
	}
	return methodcmd
}

func requiredFlags(cmd *cobra.Command, plaintext *bool, plaintextset *bool, addr *string) {
	cmd.Flags().BoolVar(plaintext, "plaintext", false, "")
	*plaintextset = cmd.Flag("plaintext").Changed
	err := cmd.RegisterFlagCompletionFunc("plaintext", cobra.NoFileCompletions)
	cobra.CheckErr(err)
	cmd.Flags().StringVar(addr, "addr", "", "")
	err = cmd.RegisterFlagCompletionFunc("addr", cobra.NoFileCompletions)
	cobra.CheckErr(err)
}
