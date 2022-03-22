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
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		log.Fatal(err)
	}
}
