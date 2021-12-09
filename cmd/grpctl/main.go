package main

import (
	"context"
	"log"
	"os"

	"github.com/joshcarp/grpctl"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "grpctl",
		Short: "an intuitive grpc cli",
	}
	err := grpctl.BuildCommand(cmd, grpctl.WithArgs(os.Args), grpctl.WithReflection(os.Args))
	if err != nil {
		log.Print(err)
	}
	if err := grpctl.RunCommand(cmd, context.Background()); err != nil {
		log.Print(err)
	}
}
