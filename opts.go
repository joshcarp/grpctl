package grpctl

import (
	"context"

	"github.com/spf13/cobra"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func RunCommand(cmd *cobra.Command, ctx context.Context) error {
	customCtx := customContext{
		ctx: &ctx,
	}
	return cmd.ExecuteContext(context.WithValue(context.Background(), configKey, &customCtx))
}

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

// WithContextFunc will modify the context  before the main command is run but not in the completion stage.
func WithContextFunc(f func(context.Context, *cobra.Command) (context.Context, error)) GrptlOption {
	return func(cmd *cobra.Command) error {
		existingPreRun := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if existingPreRun != nil {
				err := existingPreRun(cmd, args)
				if err != nil {
					return err
				}
			}
			custCtx, ctx, err := getContext(cmd)
			if err != nil {
				return err
			}
			ctx, err = f(ctx, cmd)
			if err != nil {
				return err
			}
			custCtx.setContext(ctx)
			return nil
		}
		return nil
	}
}

// WithContextDescriptorsFunc will modify the context  before the main command is run but not in the completion stage.
func WithContextDescriptorsFunc(f func(context.Context, *cobra.Command, protoreflect.MethodDescriptor) (context.Context, error)) GrptlOption {
	return func(cmd *cobra.Command) error {
		existingPreRun := cmd.PersistentPreRunE
		cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if existingPreRun != nil {
				err := existingPreRun(cmd, args)
				if err != nil {
					return err
				}
			}
			custCtx, ctx, err := getContext(cmd)
			if err != nil {
				return err
			}
			a := ctx.Value(methodDescriptorKey)
			method, ok := a.(protoreflect.MethodDescriptor)
			if !ok {
				return ContextError
			}
			ctx, err = f(ctx, cmd, method)
			if err != nil {
				return err
			}
			custCtx.setContext(ctx)
			return nil
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
