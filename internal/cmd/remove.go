package cmd

import (
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:               "remove <project-id> <service> <resource-id> [resource-id...]",
	Short:             "Remove your external IP from the ACL",
	Long:              `Fetches your external IP address and removes it from the ACL list of the specified STACKIT service instances or clusters. Multiple resource IDs can be passed to update them all at once.`,
	Args:              validateArgs,
	ValidArgsFunction: completeArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runACLAction(args, actionRemove)
	},
}
