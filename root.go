package grpctl

import (
	"github.com/spf13/cobra"
)

type GrptlOption func(*cobra.Command) error

func PersistentFlags(cmd *cobra.Command, defaultHosts ...string) error {
	var plaintext bool
	var addr string
	var cfgFile string
	var defaultHost string
	cmd.PersistentFlags().BoolVarP(&plaintext, "plaintext", "p", false, "plaintext")
	err := cmd.RegisterFlagCompletionFunc("plaintext", cobra.NoFileCompletions)
	if err != nil {
		return err
	}
	if len(defaultHosts) > 0 {
		defaultHost = defaultHosts[0]
	}
	cmd.PersistentFlags().StringVarP(&addr, "address", "a", defaultHost, "address")
	err = cmd.RegisterFlagCompletionFunc("address", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{defaultHost}, cobra.ShellCompDirectiveNoFileComp
	})
	if err != nil {
		return err
	}
	cmd.PersistentFlags().StringArrayP("header", "H", []string{}, "")
	err = cmd.RegisterFlagCompletionFunc("header", cobra.NoFileCompletions)
	if err != nil {
		return err
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.grpctl.yaml)")
	return nil
}
