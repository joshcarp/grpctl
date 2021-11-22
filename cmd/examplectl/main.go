package main

import (
	"os"

	"github.com/joshcarp/grpcexample/proto/examplepb"
	"github.com/joshcarp/grpctl"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "examplectl",
		Short: "a cli tool for examplectl",
	}
	cobra.CheckErr(grpctl.Execute(cmd, os.Args, examplepb.File_api_proto))
}
