
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
	cmd.AddCommand(ConfigCommands(config))
	cobra.CheckErr(cmd.Execute())
}

func ExecuteReflect(cmd *cobra.Command) {
	config := globalFlags(cmd)
	cmd.AddCommand(ConfigCommands(config))
	for _, e := range AddCommand(config) {
		cmd.AddCommand(e)
	}
	for _, e := range config.Services {
		descriptor, err := e.ServiceDescriptor()
		cobra.CheckErr(err)
		cmd.AddCommand(CommandFromServiceDescriptor(config, descriptor))
	}
	cobra.CheckErr(cmd.Execute())
}
