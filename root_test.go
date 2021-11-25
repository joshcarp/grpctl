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
			want: fmt.Sprintf("{\"message\":\"Incoming Message: blah \\n Metadata: map[:authority:[%s] content-type:[application/grpc] user-agent:[grpc-go/1.40.0]]\"}", addr),
		},
		{
			name: "__complete_empty_string",
			args: []string{"grpctl", "__complete", "--addr=" + addr, "--plaintext=true", ""},
			want: `BarAPI	BarAPI as defined in api.proto
FooAPI	FooAPI as defined in api.proto
ServerReflection	ServerReflection as defined in reflection/grpc_reflection_v1alpha/reflection.proto
completion	generate the autocompletion script for the specified shell
help	Help about any command
FooAPI
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
		{
			name: "__complete_BarAPI",
			args: []string{"grpctl", "__complete", "--addr=" + addr, "--plaintext=true", "BarAPI", ""},
			want: `ListBars	ListBars as defined in api.proto
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
			bs := b.String()
			require.Equal(t, tt.want, bs)
		})
	}
}
