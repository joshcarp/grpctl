package grpctl

import (
	"context"

	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
)

func AddCommand(config Config) *cobra.Command {
	var addr string
	var plaintext bool
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add services to grpctl",
		Run: func(cmd *cobra.Command, args []string) {
			conn, err := setup(context.Background(), plaintext, addr)
			cobra.CheckErr(err)
			fds, err := grpcreflect(conn)
			cobra.CheckErr(err)
			reflectfds, err := ConvertToProtoReflectDesc(fds)
			cobra.CheckErr(err)
			for _, desc := range reflectfds {
				for _, service := range descriptors.NewFileDescriptor(desc).Services() {
					config.Services = append(config.Services,
						NewService(fds, service.ServiceDescriptor, addr, plaintext))
				}
			}
			cobra.CheckErr(config.Save())
		},
	}
	addCmd.Flags().BoolVar(&plaintext, "plaintext", false, "")
	addCmd.Flags().StringVar(&addr, "addr", "", "")
	return addCmd
}
