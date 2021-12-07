package grpctl

import (
	"context"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"testing"
)

func TestSetCommandContext(t *testing.T) {
	cmd := &cobra.Command{
		Use: "Foobar",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			err := setCommandContext(cmd, metadata.AppendToOutgoingContext(context.Background(), "pre", "run"), args)
			require.NoError(t, err)

		},
	}
	child := &cobra.Command{
		Use: "foobar",
		Run: func(cmd *cobra.Command, args []string) {
			_, added, ok := metadata.FromOutgoingContextRaw(cmd.Context())
			require.True(t, ok)
			require.Equal(t, [][]string{{"pre", "run"}}, added)
		},
	}
	child.Flags().String("foobar", "", "")
	cmd.AddCommand(child)
	cmd.SetArgs([]string{"foobar", "--foobar=blah"})
	require.NoError(t, cmd.Execute())
}

func TestWithContextFunc(t *testing.T) {
	cmd := &cobra.Command{
		Use: "Foobar",
	}
	child := &cobra.Command{
		Use: "foobar",
		Run: func(cmd *cobra.Command, args []string) {
			_, added, ok := metadata.FromOutgoingContextRaw(cmd.Context())
			require.True(t, ok)
			require.Equal(t, [][]string{{"pre", "run"}}, added)
		},
	}
	child.Flags().String("foobar", "", "")
	cmd.AddCommand(child)
	cmd.SetArgs([]string{"foobar", "--foobar=blah"})

	err := BuildCommand(cmd, WithContextFunc(func(ctx context.Context) (context.Context, error) {
		return metadata.AppendToOutgoingContext(ctx, "pre", "run"), nil
	}))
	require.NoError(t, err)
	require.NoError(t, cmd.Execute())
}
