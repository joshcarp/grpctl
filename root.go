package grpctl

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Execute(cmd *cobra.Command, args []string, descriptors ...protoreflect.FileDescriptor) error {
	var err error
	cmd.SetArgs(args[1:])
	if err = PersistentFlags(cmd, ""); err != nil {
		return err
	}
	err = CommandFromFileDescriptors(cmd, descriptors...)
	if err != nil {
		return err
	}
	return cmd.Execute()
}

func PersistentFlags(cmd *cobra.Command, defaultHost string) error {
	var plaintext bool
	var addr string
	var cfgFile string
	cmd.PersistentFlags().BoolVar(&plaintext, "plaintext", false, "plaintext")
	err := cmd.RegisterFlagCompletionFunc("plaintext", cobra.NoFileCompletions)
	if err != nil {
		return err
	}
	cmd.PersistentFlags().StringVar(&addr, "addr", defaultHost, "address")
	err = cmd.RegisterFlagCompletionFunc("addr", cobra.NoFileCompletions)
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

func ExecuteReflect(cmd *cobra.Command, args []string) (err error) {
	cmd.ValidArgsFunction = func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		fds, err2 := reflectfiledesc(args)
		if err2 != nil {
			err = err2
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var opts []string
		 err2 = CommandFromFileDescriptors(cmd, fds...)
		if err2 != nil {
			err = err2
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return opts, cobra.ShellCompDirectiveNoFileComp
	}

	fds, err := reflectfiledesc(args[1:])
	if err != nil {
		return err
	}
	return Execute(cmd, args, fds...)
}
