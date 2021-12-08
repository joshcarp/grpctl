package grpctl

import (
	"bytes"
	"context"

	"github.com/spf13/cobra"
)

// TODO: delete this when one can modify the cobra context: https://github.com/spf13/cobra/pull/1551
func setCommandContext(cmd *cobra.Command, ctx context.Context, args []string) error {
	PersistentPreRun := cmd.PersistentPreRun
	PersistentPreRunE := cmd.PersistentPreRunE
	PreRun := cmd.PreRun
	PreRunE := cmd.PreRunE
	Run := cmd.Run
	RunE := cmd.RunE
	PostRun := cmd.PostRun
	PostRunE := cmd.PostRunE
	PersistentPostRun := cmd.PersistentPostRun
	PersistentPostRunE := cmd.PersistentPostRunE
	out := cmd.OutOrStdout()
	cmd.PersistentPreRun = nil
	cmd.PersistentPreRunE = nil
	cmd.PreRun = nil
	cmd.PreRunE = nil
	cmd.Run = nil
	cmd.RunE = nil
	cmd.PostRun = nil
	cmd.PostRunE = nil
	cmd.PersistentPostRun = nil
	cmd.PersistentPostRunE = nil
	cmd.DisableFlagParsing = true
	cmd.SetArgs([]string{})
	cmd.SetOut(&bytes.Buffer{})
	err := cmd.ExecuteContext(ctx)
	if err != nil {
		return err
	}
	cmd.PersistentPreRun = PersistentPreRun
	cmd.PersistentPreRunE = PersistentPreRunE
	cmd.PreRun = PreRun
	cmd.PreRunE = PreRunE
	cmd.Run = Run
	cmd.RunE = RunE
	cmd.PostRun = PostRun
	cmd.PostRunE = PostRunE
	cmd.PersistentPostRun = PersistentPostRun
	cmd.PersistentPostRunE = PersistentPostRunE
	cmd.DisableFlagParsing = false
	cmd.SetArgs(args)
	cmd.SetOut(out)
	return nil
}

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
