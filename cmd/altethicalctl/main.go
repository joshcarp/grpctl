package main

import (
	"github.com/joshcarp/grpcexample/proto/examplepb"
	"github.com/joshcarp/grpctl"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "altethical",
		Short: "a cli tool for altethical",
	}
	grpctl.Execute(cmd, examplepb.File_api_proto)
}
