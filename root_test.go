package grpctl

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/joshcarp/grpcexample/pkg/example"
	"github.com/joshcarp/grpcexample/proto/examplepb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/spf13/cobra"
)

func TestExecuteReflect(t *testing.T) {
	port, err := example.ServeRand(context.Background(), func(server *grpc.Server) { examplepb.RegisterFooAPIServer(server, &example.FooServer{}) })
	require.NoError(t, err)
	addr := fmt.Sprintf("localhost:%d", port)
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			args: []string{"grpctl", "--addr=" + addr, "--plaintext=true", "FooAPI", "Hello", "--message", "blah"},
			want: `{"message":"Hello blah"}`,
		},
		{
			name: "__complete_empty_string",
			args: []string{"grpctl", "__complete", "--addr=" + addr, "--plaintext=true", ""},
			want: `FooAPI
BarAPI
ServerReflection
:4
`,
		},
		{
			name: "__complete_empty",
			args: []string{"grpctl", "__complete", "--addr=" + addr, "--plaintext=true"},
			want: `true
false
:4
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			var b bytes.Buffer
			cmd.SetOut(&b)
			if err := ExecuteReflect(cmd, tt.args); (err != nil) != tt.wantErr {
				t.Errorf("ExecuteReflect() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, tt.want, b.String())
		})
	}
}
