//go:generate go run .
package main

import (
	"log"

	"github.com/joshcarp/grpctl"

	"github.com/spf13/cobra/doc"
)

func main() {
	cmd, err := grpctl.ReflectionCommand()
	if err != nil {
		log.Fatal(err)
	}
	err = doc.GenMarkdownTree(cmd, ".")
	if err != nil {
		log.Fatal(err)
	}
}
