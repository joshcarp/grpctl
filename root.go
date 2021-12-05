package grpctl

import (
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Execute(cmd *cobra.Command, args []string, descriptors ...protoreflect.FileDescriptor) (*cobra.Command, error) {
	var err error
	cmd.SetArgs(args[1:])
	if err = PersistentFlags(cmd, ""); err != nil {
		return nil, err
	}
	err = CommandFromFileDescriptors(cmd, descriptors...)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func PersistentFlags(cmd *cobra.Command, defaultHosts ...string) error {
	var plaintext bool
	var addr string
	var cfgFile string
	var defaultHost string
	cmd.PersistentFlags().BoolVar(&plaintext, "plaintext", false, "plaintext")
	err := cmd.RegisterFlagCompletionFunc("plaintext", cobra.NoFileCompletions)
	if err != nil {
		return err
	}
	if len(defaultHosts) > 0 {
		defaultHost = defaultHosts[0]
	}
	cmd.PersistentFlags().StringVar(&addr, "address", defaultHost, "address")
	err = cmd.RegisterFlagCompletionFunc("address", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{defaultHost}, cobra.ShellCompDirectiveNoFileComp
	})
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
		fds, err2 := reflectFileDesc(args)
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

	fds, err := reflectFileDesc(args[1:])
	if err != nil {
		return err
	}
	finalcmd, err := Execute(cmd, args, fds...)
	if err != nil {
		return err
	}
	return finalcmd.Execute()
}
