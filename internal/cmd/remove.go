package cmd

import (
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <service> <resource-group> <resource-id>",
	Short: "Remove your external IP from the ACL",
	Long:  `Fetches your external IP address and removes it from the ACL list of the specified STACKIT service instance or cluster.`,
	Args:  validateArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runACLAction(args, actionRemove)
	},
}
