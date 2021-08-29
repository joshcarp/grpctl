package grpctl

import (
	"context"
	"fmt"
	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
)

func AddCommand(config Config) []*cobra.Command {
	var addr string
	var plaintext bool
	var plaintextset bool
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a services to grpctl",
		Run: func(cmd *cobra.Command, args []string) {
			conn, err := setup(context.Background(), plaintext, addr)
			cobra.CheckErr(err)
			fds, err := reflect(conn)
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
	requiredFlags(addCmd, &plaintext, &plaintextset, &addr)
	list := &cobra.Command{
		Use:   "list",
		Short: "list service",
		Run: func(cmd *cobra.Command, args []string) {
			for _, service := range config.Services {
				for _, method := range service.Methods {
					fmt.Println(service.Name, method.Name)
				}

			}
		},
	}
	return []*cobra.Command{addCmd, list}
}
