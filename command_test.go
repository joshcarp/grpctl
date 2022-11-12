package grpctl

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"testing"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"google.golang.org/grpc/metadata"

	"github.com/joshcarp/grpctl/internal/testing/pkg/example"
	"github.com/joshcarp/grpctl/internal/testing/proto/examplepb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/spf13/cobra"
)

func TestBuildCommand(t *testing.T) {
	t.Parallel()
	port, err := example.ServeRand(
		context.Background(),
		func(server *grpc.Server) {
			examplepb.RegisterFooAPIServer(server, &example.FooServer{})
		})
	require.NoError(t, err)
	addr := fmt.Sprintf("localhost:%d", port)
	tests := []struct {
		name    string
		args    []string
		want    string
		opts    func([]string) []CommandOption
		json    string
		wantErr bool
	}{
		{
			name: "basic",
			args: []string{
				"grpctl",
				"--address=http://" + addr,
				"FooAPI",
				"Hello",
				"--message",
				"blah",
			},
			opts: func(args []string) []CommandOption {
				return []CommandOption{
					WithArgs(args),
					WithReflection(args),
				}
			},
			json: fmt.Sprintf("{\n \"message\": \"Incoming Message: blah \\n "+
				"Metadata: map[:authority:[%s] accept-encoding:[identity] "+
				"content-type:[application/grpc+proto] grpc-accept-encoding:[gzip] "+
				"user-agent:[grpc-go-connect/1.1.0 (%s)]]\"\n}", addr, runtime.Version()),
		},
		{
			name: "completion_enabled",
			args: []string{
				"root",
			},
			opts: func(args []string) []CommandOption {
				return []CommandOption{
					WithArgs(args),
					WithReflection(args),
					WithCompletion(),
				}
			},
			want: `Usage:
  root [command]

Available Commands:
  completion  Generate completion script
  help        Help about any command

Flags:
  -a, --address string       Address in form 'host:port'
  -H, --header stringArray   Header in form 'key: value'
  -h, --help                 help for root
  -p, --protocol string      protocol to use: [connect, grpc, grpcweb] (default "grpc")

Use "root [command] --help" for more information about a command.
`,
		},
		{
			name: "__complete_empty_string",
			args: []string{"grpctl", "__complete", "--address=http://" + addr, ""},
			opts: func(args []string) []CommandOption {
				return []CommandOption{
					WithArgs(args),
					WithReflection(args),
				}
			},
			want: `BarAPI	BarAPI as defined in api.proto
FooAPI	FooAPI as defined in api.proto
completion	Generate the autocompletion script for the specified shell
help	Help about any command
:4
`,
		},
		{
			name: "__complete_empty",
			args: []string{"grpctl", "__complete", "--address=http://" + addr},
			opts: func(args []string) []CommandOption {
				return []CommandOption{
					WithArgs(args),
					WithReflection(args),
				}
			},
			want: `:0
`,
		},
		{
			name: "__complete_BarAPI",
			args: []string{"grpctl", "__complete", "--address=http://" + addr, "BarAPI", ""},
			opts: func(args []string) []CommandOption {
				return []CommandOption{
					WithArgs(args),
					WithReflection(args),
				}
			},
			want: `ListBars	ListBars as defined in api.proto
:4
`,
		},
		{
			name: "header",
			args: []string{
				"grpctl",
				"--address=http://" + addr,
				"-H=Foo:Bar",
				"FooAPI",
				"Hello",
				"--message",
				"blah",
			},
			opts: func(args []string) []CommandOption {
				return []CommandOption{
					WithArgs(args),
					WithReflection(args),
				}
			},
			json: fmt.Sprintf("{\n \"message\": \"Incoming Message: blah \\n "+
				"Metadata: map[:authority:[%s] accept-encoding:[identity] "+
				"content-type:[application/grpc+proto] foo:[Bar] grpc-accept-encoding:[gzip] "+
				"user-agent:[grpc-go-connect/1.1.0 (%s)]]\"\n}", addr, runtime.Version()),
		},
		{
			name: "headers",
			args: []string{
				"grpctl",
				"--address=http://" + addr,
				"-H=Foo:Bar",
				"-H=Foo2:Bar2",
				"FooAPI",
				"Hello",
				"--message",
				"blah",
			},
			opts: func(args []string) []CommandOption {
				return []CommandOption{
					WithArgs(args),
					WithReflection(args),
				}
			},
			json: fmt.Sprintf("{\n \"message\":  \"Incoming Message: blah \\n "+
				"Metadata: map[:authority:[%s] accept-encoding:[identity] content-type:[application/grpc+proto] "+
				"foo:[Bar] foo2:[Bar2] grpc-accept-encoding:[gzip] "+
				"user-agent:[grpc-go-connect/1.1.0 (%s)]]\"\n}", addr, runtime.Version()),
		},
		{
			name: "WithContextFunc-No-Change",
			args: []string{
				"grpctl",
				"--address=http://" + addr,
				"-H=Foo:Bar",
				"-H=Foo2:Bar2",
				"FooAPI",
				"Hello",
				"--message",
				"blah",
			},
			opts: func(args []string) []CommandOption {
				return []CommandOption{
					WithContextFunc(func(ctx context.Context, cmd *cobra.Command) (context.Context, error) {
						return ctx, nil
					}),
					WithArgs(args),
					WithReflection(args),
				}
			},
			json: fmt.Sprintf("{\n \"message\":  \"Incoming Message: blah \\n "+
				"Metadata: map[:authority:[%s] accept-encoding:[identity] content-type:[application/grpc+proto] "+
				"foo:[Bar] foo2:[Bar2] grpc-accept-encoding:[gzip] "+
				"user-agent:[grpc-go-connect/1.1.0 (%s)]]\"\n}", addr, runtime.Version()),
		},
		{
			name: "WithContextFunc-No-Change",
			args: []string{
				"grpctl",
				"--address=http://" + addr,
				"-H=Foo:Bar",
				"-H=Foo2:Bar2",
				"FooAPI",
				"Hello",
				"--message",
				"blah",
			},
			opts: func(args []string) []CommandOption {
				return []CommandOption{
					WithContextFunc(func(ctx context.Context, cmd *cobra.Command) (context.Context, error) {
						return ctx, nil
					}),
					WithArgs(args),
					WithReflection(args),
				}
			},
			json: fmt.Sprintf("{\n \"message\":  \"Incoming Message: blah \\n "+
				"Metadata: map[:authority:[%s] accept-encoding:[identity] content-type:[application/grpc+proto] "+
				"foo:[Bar] foo2:[Bar2] grpc-accept-encoding:[gzip] "+
				"user-agent:[grpc-go-connect/1.1.0 (%s)]]\"\n}", addr, runtime.Version()),
		},
		{
			name: "WithContextFunc",
			args: []string{
				"grpctl",
				"--address=http://" + addr,
				"-H=Foo:Bar",
				"-H=Foo2:Bar2",
				"FooAPI",
				"Hello",
				"--message",
				"blah",
			},
			opts: func(args []string) []CommandOption {
				return []CommandOption{
					WithContextFunc(func(ctx context.Context, _ *cobra.Command) (context.Context, error) {
						return metadata.AppendToOutgoingContext(ctx, "fookey", "fooval"), nil
					}),
					WithArgs(args),
					WithReflection(args),
				}
			},
			json: fmt.Sprintf("{\n \"message\": \"Incoming Message: blah \\n "+
				"Metadata: map[:authority:[%s] accept-encoding:[identity] content-type:[application/grpc+proto] "+
				"foo:[Bar] foo2:[Bar2] fookey:[fooval] grpc-accept-encoding:[gzip] "+
				"user-agent:[grpc-go-connect/1.1.0 (%s)]]\"\n}", addr, runtime.Version()),
		},
		{
			name: "WithDescriptorContextFuncSimple",
			args: []string{
				"grpctl",
				"--address=http://" + addr,
				"-H=Foo:Bar",
				"-H=Foo2:Bar2",
				"FooAPI",
				"Hello",
				"--message",
				"blah",
			},
			opts: func(args []string) []CommandOption {
				return []CommandOption{
					WithContextDescriptorFunc(func(ctx context.Context, _ *cobra.Command, _ protoreflect.MethodDescriptor) (context.Context, error) {
						return metadata.AppendToOutgoingContext(ctx, "fookey", "fooval"), nil
					}),
					WithArgs(args),
					WithReflection(args),
				}
			},
			json: fmt.Sprintf("{\n \"message\":  \"Incoming Message: "+
				"blah \\n Metadata: map[:authority:[%s] accept-encoding:[identity] "+
				"content-type:[application/grpc+proto] foo:[Bar] foo2:[Bar2] fookey:[fooval] "+
				"grpc-accept-encoding:[gzip] user-agent:[grpc-go-connect/1.1.0 (%s)]]\"\n}", addr, runtime.Version()),
		},
		{
			name: "WithDescriptorContextFuncMethodDescriptorsUsed",
			args: []string{
				"grpctl",
				"--address=http://" + addr,
				"-H=Foo:Bar",
				"-H=Foo2:Bar2",
				"FooAPI",
				"Hello",
				"--message",
				"blah",
			},
			opts: func(args []string) []CommandOption {
				return []CommandOption{
					WithContextDescriptorFunc(func(ctx context.Context, _ *cobra.Command, descriptor protoreflect.MethodDescriptor) (context.Context, error) {
						serviceDesc := descriptor.Parent()
						service, ok := serviceDesc.(protoreflect.ServiceDescriptor)
						require.True(t, ok)
						b := proto.GetExtension(service.Options(), annotations.E_DefaultHost)
						bstr, _ := b.(string)
						return metadata.AppendToOutgoingContext(ctx, "fookey", bstr), nil
					}),
					WithArgs(args),
					WithReflection(args),
				}
			},
			json: fmt.Sprintf(
				"{\n \"message\":  \"Incoming Message: "+
					"blah \\n Metadata: map[:authority:[%s] "+
					"accept-encoding:[identity] content-type:[application/grpc+proto] "+
					"foo:[Bar] foo2:[Bar2] fookey:[] grpc-accept-encoding:[gzip] "+
					"user-agent:[grpc-go-connect/1.1.0 (%s)]]\"\n}", addr, runtime.Version()),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := &cobra.Command{
				Use: "root",
			}
			var b bytes.Buffer
			cmd.SetOut(&b)
			if err := BuildCommand(cmd, tt.opts(tt.args)...); (err != nil) != tt.wantErr {
				t.Errorf("ExecuteReflect() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := cmd.ExecuteContext(context.Background()); err != nil {
				t.Errorf("ExecuteReflect() error = %v, wantErr %v", err, tt.wantErr)
			}
			bs := b.String()
			if tt.json != "" {
				require.JSONEq(t, tt.json, bs)
				return
			}
			require.Equal(t, tt.want, bs)
		})
	}
}

func TestRunCommand(t *testing.T) {
	t.Parallel()
	type contextkey struct{}
	tests := []struct {
		name    string
		args    *cobra.Command
		wantErr bool
	}{
		{
			name: "",
			args: &cobra.Command{
				Use: "foobar",
				PreRunE: func(cmd *cobra.Command, args []string) error {
					cmd.Root().SetContext(context.WithValue(cmd.Root().Context(), contextkey{}, "bar"))
					return nil
				},
				RunE: func(cmd *cobra.Command, args []string) error {
					require.Equal(t, "bar", cmd.Root().Context().Value(contextkey{}))
					return nil
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.args.ExecuteContext(context.Background())
			require.NoError(t, err)
		})
	}
}
