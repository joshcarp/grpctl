package grpctl

import (
	"github.com/spf13/cobra"
)

func recusiveParentPreRun(cmd *cobra.Command, args []string) error {
	for cmd != nil {
		this := cmd
		if cmd.PersistentPreRunE != nil {
			err := cmd.PersistentPreRunE(this, args)
			if err != nil {
				return err
			}
		}
		if !this.HasParent() {
			break
		}
		cmd = this.Parent()
	}
	return nil
}
