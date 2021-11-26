package main

import (
	"os"

	"google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/joshcarp/grpctl"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "examplectl",
		Short: "a cli tool for examplectl",
	}
	cobra.CheckErr(grpctl.Execute(cmd, os.Args,
		secretmanager.File_google_cloud_secretmanager_v1_service_proto,
		secretmanager.File_google_cloud_secretmanager_v1_resources_proto,
	))
}
