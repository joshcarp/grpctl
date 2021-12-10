package main

import (
	"context"
	"log"

	"github.com/joshcarp/grpctl"
)

func main() {
	cmd, err := grpctl.ReflectionCommand()
	if err != nil {
		log.Fatal(err)
	}
	if err := grpctl.RunCommand(cmd, context.Background()); err != nil {
		log.Fatal(err)
	}
}
