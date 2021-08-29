/*
Copyright © 2021 Joshua Carpeggiani josh@joshcarp.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
	cmd.AddCommand(cmd)
	cobra.CheckErr(cmd.Execute())
}
