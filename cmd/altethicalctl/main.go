package main

import (
	"github.com/joshcarp/grpctl"
	"github.com/joshcarp/altethical/backend/pkg/proto/altethical"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "altethical",
		Short: "a cli tool for altethical",
	}
	grpctl.Execute(cmd, altethical.File_api_proto)
}