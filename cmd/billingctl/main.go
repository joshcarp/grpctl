package main

import (
	"context"
	"log"
	"os"

	"github.com/joshcarp/grpctl"
	"github.com/spf13/cobra"
	"google.golang.org/genproto/googleapis/cloud/billing/v1"
)

// Example call: billingctl \
// -H="Authorization: Bearer $(gcloud auth application-default print-access-token)" CloudBilling ListBillingAccounts
func main() {
	cmd := &cobra.Command{
		Use:   "billingctl",
		Short: "an example cli tool for the gcp billing api",
	}
	err := grpctl.BuildCommand(cmd,
		grpctl.WithArgs(os.Args),
		grpctl.WithFileDescriptors(
			billing.File_google_cloud_billing_v1_cloud_billing_proto,
			billing.File_google_cloud_billing_v1_cloud_catalog_proto,
		),
	)
	if err != nil {
		log.Print(err)
	}
	if err := grpctl.RunCommand(cmd, context.Background()); err != nil {
		log.Print(err)
	}
}
