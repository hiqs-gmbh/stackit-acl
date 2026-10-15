package cmd

import (
	"fmt"
	"strings"

	"github.com/hiqs-gmbh/stackit-acl/internal/services"
	"github.com/hiqs-gmbh/stackit-acl/internal/stackit"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:               "list <project-id> <service> [resource-id...]",
	Short:             "List the currently applied ACLs",
	Long: `Fetches and displays the currently applied ACLs of the specified STACKIT service instances or clusters.

If no resource ID is passed, the ACLs of all instances of the given service in the project are listed.`,
	Args:              validateListArgs,
	ValidArgsFunction: completeArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runList(args)
	},
}

func validateListArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("requires at least 2 arguments: <project-id> <service> [resource-id...]\n\nSupported services:\n%s", formatSupportedServices())
	}

	if _, ok := services.GetByName(args[1]); !ok {
		return fmt.Errorf("unsupported service: %s\n\nSupported services:\n%s", args[1], formatSupportedServices())
	}

	return nil
}

func runList(args []string) error {
	if err := stackit.CheckAvailable(); err != nil {
		return err
	}

	projectID := args[0]
	service := args[1]
	resourceIDs := uniqueResourceIDs(args[2:])

	cfg, _ := services.GetByName(service)
	client := stackit.New(projectID, region)

	if len(resourceIDs) == 0 {
		logStep(fmt.Sprintf("Listing %s resources...", service))
		instances, err := stackit.ListInstances(projectID, region, cfg)
		if err != nil {
			return err
		}
		for _, instance := range instances {
			resourceIDs = append(resourceIDs, instance.ID)
		}
		if len(resourceIDs) == 0 {
			logWarn(fmt.Sprintf("No %s resources found in the project.", service))
			return nil
		}
	}

	var failedLabels []string
	var failedErrs []error

	for _, resourceID := range resourceIDs {
		resourceLabel, currentACLs, _, err := fetchCurrentACLs(client, cfg, service, resourceID)
		if err != nil {
			failedLabels = append(failedLabels, fmt.Sprintf("%q", resourceID))
			failedErrs = append(failedErrs, err)
			continue
		}
		logInfo(fmt.Sprintf("ACLs for %s %s:", service, resourceLabel))
		logRaw(formatACLEntries(currentACLs))
	}

	if len(failedLabels) > 0 {
		var sb strings.Builder
		for i, label := range failedLabels {
			sb.WriteString("  " + service + " " + label + ": " + failedErrs[i].Error() + "\n")
		}
		return fmt.Errorf("failed to fetch the ACLs of %d of %d resources:\n%s", len(failedLabels), len(resourceIDs), sb.String())
	}

	return nil
}
