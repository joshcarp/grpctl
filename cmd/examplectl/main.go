package main

import (
	"github.com/joshcarp/grpcexample/proto/examplepb"
	"github.com/joshcarp/grpctl"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "examplectl",
		Short: "a cli tool for examplectl",
	}
	grpctl.Execute(cmd, examplepb.File_api_proto)
}
