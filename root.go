package grpctl

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Execute(cmd *cobra.Command, args []string, descriptors ...protoreflect.FileDescriptor) error {
	cmd.SetArgs(args[1:])
	for _, serviceCmds := range CommandFromFileDescriptors(descriptors...) {
		cmd.AddCommand(serviceCmds)
	}
	return cmd.Execute()
}

func ExecuteReflect(cmd *cobra.Command, args []string) (err error) {
	var plaintext bool
	var addr string
	cmd.PersistentFlags().BoolVar(&plaintext, "plaintext", false, "plaintext")
	cmd.PersistentFlags().StringVar(&addr, "addr", "", "address")
	cmd.PersistentFlags().StringArrayP("header", "H", []string{}, "")

	cmd.ValidArgsFunction = func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		fds, err2 := reflectfiledesc(args)
		if err2 != nil {
			err = err2
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var opts []string
		for _, serviceCmds := range CommandFromFileDescriptors(fds...) {
			opts = append(opts, serviceCmds.Name())
		}
		return opts, cobra.ShellCompDirectiveNoFileComp
	}
	err = cmd.RegisterFlagCompletionFunc("plaintext", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		return err
	}
	err = cmd.RegisterFlagCompletionFunc("addr", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"address"}, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		return err
	}
	fds, err := reflectfiledesc(args[1:])
	if err != nil {
		return err
	}
	return Execute(cmd, args, fds...)
}
