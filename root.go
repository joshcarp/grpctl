
package grpctl

import (
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/spf13/cobra"
)

func globalFlags(cmd *cobra.Command) Config {
	var cfgFile string
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.grpctl.yaml)")
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	return initConfig(cfgFile)
}

func Execute(cmd *cobra.Command, descriptors ...protoreflect.FileDescriptor) {
	config := globalFlags(cmd)
	for _, serviceCmds := range CommandFromFileDescriptors(config, descriptors...) {
		cmd.AddCommand(serviceCmds)
	}
	cobra.CheckErr(cmd.Execute())
}

func ExecuteReflect(cmd *cobra.Command) {
	config := globalFlags(cmd)
	for _, e := range config.Services {
		descriptor, err := e.ServiceDescriptor()
		cobra.CheckErr(err)
		cmd.AddCommand(CommandFromServiceDescriptor(config, descriptor))
	}
	cmd.AddCommand(AddCommand(config))
	cmd.AddCommand(ConfigCommands(config))
	cmd.AddCommand(GetContextCommand(config))
	cmd.AddCommand(GetServiceCommand(config))
	cmd.AddCommand(GetUserCommand(config))
	cobra.CheckErr(cmd.Execute())
}
