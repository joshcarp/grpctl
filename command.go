package grpctl

import (
	"context"
	"fmt"

	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func CommandFromFileDescriptors(descriptors ...protoreflect.FileDescriptor) []*cobra.Command {
	var cmds []*cobra.Command
	for _, desc := range descriptors {
		cmds = append(cmds, CommandFromFileDescriptor(desc)...)
	}
	return cmds
}

func CommandFromFileDescriptor(methods protoreflect.FileDescriptor) []*cobra.Command {
	var cmds []*cobra.Command
	for _, service := range descriptors.NewFileDescriptor(methods).Services() {
		cmds = append(cmds, CommandFromServiceDescriptor(service.ServiceDescriptor))
	}
	return cmds
}

func CommandFromServiceDescriptor(service protoreflect.ServiceDescriptor) *cobra.Command {
	servicedesc := descriptors.NewServiceDescriptor(service)
	serviceCmd := cobra.Command{
		Use:   servicedesc.Command(),
		Short: fmt.Sprintf("%s as defined in %s", servicedesc.Command(), service.ParentFile().Path()),
	}
	for _, method := range servicedesc.Methods() {
		methodCmd, _ := CommandFromMethodDescriptor(method)
		serviceCmd.AddCommand(&methodCmd)
	}
	return &serviceCmd
}

func CommandFromMethodDescriptor(method descriptors.MethodDescriptor) (cobra.Command, error) {
	dataMap := make(descriptors.DataMap)
	for fieldNum := 0; fieldNum < method.Input().Fields().Len(); fieldNum++ {
		field := method.Input().Fields().Get(fieldNum)
		jsonName := field.JSONName()
		field.Default()
		field.Kind()
		dataMap[jsonName] = &descriptors.DataValue{Kind: field.Kind(), Value: field.Default().Interface(), Proto: true}
	}
	var addr, inputData, data string
	var plaintext bool
	methodCmd := cobra.Command{
		Use:   method.Command(),
		Short: fmt.Sprintf("%s as defined in %s", method.Command(), method.ParentFile().Path()),
		RunE: func(cmd *cobra.Command, args []string) error {
			//a, err := cmd.Flags().GetStringArray("header")
			//if err != nil {
			//	return err
			//}
			ctx := context.Background()
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
	requiredFlags(&methodCmd, &plaintext, &addr)
	defaults, templ := MakeJsonTemplate(method.Input())
	err := methodCmd.RegisterFlagCompletionFunc("json-data", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{templ}, cobra.ShellCompDirectiveDefault
	})
	if err != nil {
		return cobra.Command{}, err
	}
	for key, val := range dataMap {
		methodCmd.Flags().Var(val, key, "")
		err := methodCmd.RegisterFlagCompletionFunc(key, func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
			return []string{fmt.Sprintf("%v", defaults[key])}, cobra.ShellCompDirectiveDefault
		})
		if err != nil {
			return cobra.Command{}, err
		}
	}
	return methodCmd, nil
}

func requiredFlags(cmd *cobra.Command, plaintext *bool, addr *string) {
	cmd.Flags().BoolVar(plaintext, "plaintext", false, "")
	err := cmd.RegisterFlagCompletionFunc("plaintext", cobra.NoFileCompletions)
	cobra.CheckErr(err)
	cmd.Flags().StringVar(addr, "addr", "", "")
	err = cmd.RegisterFlagCompletionFunc("addr", cobra.NoFileCompletions)
	cobra.CheckErr(err)
}
