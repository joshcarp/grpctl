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
	service := &cobra.Command{Use: "service"}
	service.AddCommand(AddCommand(config))
	service.AddCommand(GetServiceCommands(config)...)
	cmd.AddCommand(service)
	user := &cobra.Command{Use: "user"}
	user.AddCommand(GetSetUser(config))
	user.AddCommand(GetUserCommands(config)...)
	cmd.AddCommand(user)
	cobra.CheckErr(cmd.Execute())
}

func GetSetUser(config Config) *cobra.Command {
	return &cobra.Command{
		Use:       "set",
		Short:     "set the current user",
		ValidArgs: config.Users.Names(),
		Args:      cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config, err := config.SetUser(args[0])
			cobra.CheckErr(err)
			cobra.CheckErr(config.Save())
		},
	}
}
