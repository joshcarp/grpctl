package grpctl

import (
	"context"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func BuildCommand(cmd *cobra.Command, opts ...GrptlOption) error {
	for _, f := range opts {
		err := f(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func WithFileDescriptors(descriptors ...protoreflect.FileDescriptor) GrptlOption {
	return func(cmd *cobra.Command) error {
		err := CommandFromFileDescriptors(cmd, descriptors...)
		if err != nil {
			return err
		}
		return nil
	}
}

// WithContext must be run on the root command before anything is added to it
func WithContext(ctx context.Context) GrptlOption {
	return func(cmd *cobra.Command) error {
		var err error
		_, err = cmd.ExecuteContextC(ctx)
		if err != nil {
			panic(err)
		}
		return nil
	}
}

func WithArgs(args []string) GrptlOption {
	return func(cmd *cobra.Command) error {
		cmd.SetArgs(args[1:])
		return nil
	}
}

func WithReflection(args []string) GrptlOption {
	return func(cmd *cobra.Command) error {
		var err error
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
		if err = PersistentFlags(cmd, ""); err != nil {
			return err
		}
		err = CommandFromFileDescriptors(cmd, fds...)
		if err != nil {
			return err
		}
		return nil
	}
}
