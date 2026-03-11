package main

import (
	"context"
	"github.com/joshcarp/grpctl"
	"log"
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
