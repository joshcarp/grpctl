package main

import (
	"github.com/joshcarp/grpctl"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "grpctl",
		Short: "A brief description of your application",
	}
	grpctl.ExecuteReflect(cmd)
}
