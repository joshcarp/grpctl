package grpctl

import (
	"context"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// WithContextFunc will add commands to the cobra command through the file descriptors provided.
func WithFileDescriptors(descriptors ...protoreflect.FileDescriptor) CommandOption {
	return func(cmd *cobra.Command) error {
		err := CommandFromFileDescriptors(cmd, descriptors...)
		if err != nil {
			return err
		}
		return nil
	}
}

// WithContextFunc will modify the context  before the main command is run but not in the completion stage.
func WithContextFunc(f func(context.Context, *cobra.Command) (context.Context, error)) CommandOption {
	return func(cmd *cobra.Command) error {
		existingPreRun := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if existingPreRun != nil {
				err := existingPreRun(cmd, args)
				if err != nil {
					return err
				}
			}
			custCtx, ctx, ok := getContext(cmd)
			if !ok {
				return nil
			}
			ctx, err := f(ctx, cmd)
			if err != nil {
				return err
			}
			custCtx.setContext(ctx)
			return nil
		}
		return nil
	}
}

// WithContextDescriptorFunc will modify the context  before the main command is run but not in the completion stage.
func WithContextDescriptorFunc(f func(context.Context, *cobra.Command, protoreflect.MethodDescriptor) (context.Context, error)) CommandOption {
	return func(cmd *cobra.Command) error {
		existingPreRun := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if existingPreRun != nil {
				err := existingPreRun(cmd, args)
				if err != nil {
					return err
				}
			}
			custCtx, ctx, ok := getContext(cmd)
			if !ok {
				return nil
			}
			a := ctx.Value(methodDescriptorKey{})
			method, ok := a.(protoreflect.MethodDescriptor)
			if !ok {
				return nil
			}
			ctx, err := f(ctx, cmd, method)
			if err != nil {
				return err
			}
			custCtx.setContext(ctx)
			return nil
		}
		return nil
	}
}

// WithArgs will set the args of the command as args[1:].
func WithArgs(args []string) CommandOption {
	return func(cmd *cobra.Command) error {
		cmd.SetArgs(args[1:])
		return nil
	}
}

// WithReflection will enable grpc reflection on the command. Use this as an alternative to WithFileDescriptors.
func WithReflection(args []string) CommandOption {
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
		if err = persistentFlags(cmd, ""); err != nil {
			return err
		}
		err = CommandFromFileDescriptors(cmd, fds...)
		if err != nil {
			return err
		}
		return nil
	}
}
