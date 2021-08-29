package grpctl

import (
	"context"
	"fmt"
	"github.com/joshcarp/grpctl/internal/descriptors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func AddCommand(config Config) []*cobra.Command {
	var addr string
	var plaintext bool
	var plaintextset bool
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a services to grpctl",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			conn, err := setup(context.Background(), plaintext, addr)
			cobra.CheckErr(err)
			fds, err := reflect(conn)
			for _, desc := range fds{
				for _, service := range descriptors.NewFileDescriptor(desc).Services(){
					config.Services = append(config.Services, service)

				}
			}
			cobra.CheckErr(err)
		},
	}
	requiredFlags(addCmd, &plaintext, &plaintextset, &addr)

	list := &cobra.Command{
		Use:   "list",
		Short: "List all the saved services",
		Run: func(cmd *cobra.Command, args []string) {
			b, err := yaml.Marshal(config)
			cobra.CheckErr(err)
			fmt.Println(string(b))
		},
	}
	setcontext := &cobra.Command{
		Use:   "remove",
		Short: "remove a service from saved services",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			newctx := args[0]
			var found bool
			for _, e := range config.Contexts {
				if e.Name == newctx {
					found = true
				}
			}
			if !found {
				log.Fatal("Context %s does not exist", newctx)
			}
			config.CurrentContext = newctx
			b, err := yaml.Marshal(config)
			cobra.CheckErr(err)
			err = os.WriteFile(config.ConfigFile, b, os.ModePerm)
			cobra.CheckErr(err)
		},
	}
	return []*cobra.Command{addCmd, list, setcontext}
}
