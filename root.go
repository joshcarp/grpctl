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
	"fmt"
	"os"
	"path"

	"google.golang.org/protobuf/reflect/protoreflect"

	"gopkg.in/yaml.v3"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

func Execute(cmd *cobra.Command, descriptors ...protoreflect.FileDescriptor) {
	var cfgFile string
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.grpctl.yaml)")
	config := initConfig(cfgFile)
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	for _, serviceCmds := range CommandFromFileDescriptor(config, descriptors...) {
		cmd.AddCommand(serviceCmds)
	}
}

func initConfig(cfgFile string) Config {
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
	if cfgFile == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			cobra.CheckErr(err)
		}
		cfgFile = path.Join(home, ".grpctl.yaml")
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			data := Config{ConfigFile: cfgFile}
			b, err := yaml.Marshal(data)
			if err != nil {
				cobra.CheckErr(err)
			}
			err = os.WriteFile(cfgFile, b, os.ModePerm)
			if err != nil {
				cobra.CheckErr(err)
			}
		}
	}
	b, err := os.ReadFile(cfgFile)
	if err != nil {
		cobra.CheckErr(err)
	}
	var config Config
	err = yaml.Unmarshal(b, &config)
	cobra.CheckErr(err)
	return config
}
