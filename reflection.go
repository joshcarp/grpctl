package grpctl

import (
	"os"

	"github.com/spf13/cobra"
)

// ReflectionCommand returns the grpctl command that is used in the grpctl binary.
func ReflectionCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "grpctl",
		Short: "an intuitive grpc cli",
	}
	err := BuildCommand(cmd, WithArgs(os.Args), WithReflection(os.Args), WithCompletion())
	if err != nil {
		return nil, err
	}
	return cmd, nil
}
