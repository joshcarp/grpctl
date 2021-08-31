package grpctl

import (
	"context"
	"fmt"
	"google.golang.org/grpc/metadata"
	"log"

	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func ConfigCommands(config Config) (*cobra.Command, *cobra.Command) {
	setcontext := &cobra.Command{
		Use:   "set",
		Short: "set current context",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			newctx := args[0]
			var found bool
			for _, e := range config.ListContext() {
				if e.Name == newctx {
					found = true
				}
			}
			if !found {
				log.Fatalf("Context %s does not exist", newctx)
			}
			config.CurrentContext = newctx
			cobra.CheckErr(config.Save())
		},
	}
	current := &cobra.Command{
		Use:   "current",
		Short: "get current context",
		Run: func(cmd *cobra.Command, args []string) {
			for _, e := range config.ListContext() {
				if e.Name == config.CurrentContext {
					fmt.Println(e)

				}
			}
		},
	}
	return setcontext, current
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
		datamap[jsonName] = &descriptors.DataValue{Kind: field.Kind(), Value: field.Default().Interface(), Proto: true}
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
			getContext, err := config.GetContext(config.CurrentContext)
			cobra.CheckErr(err)
			servicecfg, err := config.GetService(service.Command())

			if err == nil {
				environment, err := servicecfg.GetEnvironment(getContext.EnvironmentName)
				cobra.CheckErr(err)
				if !plaintextset {
					environment.Plaintext = plaintext
					servicecfg, err = servicecfg.UpdateEnvironment(environment)
					cobra.CheckErr(err)
				}
				if addr == "" {
					environment.Addr = addr
					servicecfg, err = servicecfg.UpdateEnvironment(environment)
					cobra.CheckErr(err)
				}
			}
			ctx := context.Background()
			user, err := config.GetUser(getContext.UserName)
			if err != nil {
				cobra.CheckErr(err)
			}
			for key, val := range user.Headers {
				ctx = metadata.AppendToOutgoingContext(ctx, key, val)
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
