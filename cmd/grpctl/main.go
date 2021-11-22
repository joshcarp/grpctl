package main

import (
	"log"
	"os"

	"github.com/joshcarp/grpctl"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "grpctl",
		Short: "A brief description of your application",
	}
	err := grpctl.ExecuteReflect(cmd, os.Args)
	if err != nil {
		log.Print(err)
	}
}
