package cmd

import (
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:               "add <project-id> <service> <resource-id> [resource-id...]",
	Short:             "Add your external IP to the ACL",
	Long:              `Fetches your external IP address and appends it to the ACL list of the specified STACKIT service instances or clusters. Multiple resource IDs can be passed to update them all at once.`,
	Args:              validateArgs,
	ValidArgsFunction: completeArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runACLAction(args, actionAdd)
	},
}
