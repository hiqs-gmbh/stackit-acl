package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hiqs-gmbh/stackit-acl/internal/acl"
	"github.com/hiqs-gmbh/stackit-acl/internal/ip"
	"github.com/hiqs-gmbh/stackit-acl/internal/services"
	"github.com/hiqs-gmbh/stackit-acl/internal/stackit"

	"github.com/spf13/cobra"
)

const Version = "0.2.1"

var (
	region    string
	assumeYes bool
	verbosity string
	cidr      int
	showVer   bool
)

var rootCmd = &cobra.Command{
	Use:   "stackit-acl",
	Short: "Automatically manages your external IP in STACKIT service ACLs",
	Long: `Automatically manages your external IP in STACKIT service ACLs.

This tool fetches your external IP address and adds or removes it from the
ACL list of the specified STACKIT service instance or cluster.

The stackit CLI must be installed and authenticated.

Usage:
  stackit-acl <command> <project-id> <service> <resource-id> [resource-id...] [flags]

Commands:
  add     Add your external IP to the ACL
  remove  Remove your external IP from the ACL

Supported services:
  mongodbflex
  postgresflex
  sqlserverflex
  redis
  valkey
  opensearch
  rabbitmq
  mariadb
  logme
  ske`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		setupLogger(verbosity)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if showVer {
			logInfo(Version)
			return nil
		}
		return cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&region, "region", "", "Target region for region-specific requests")
	rootCmd.PersistentFlags().BoolVarP(&assumeYes, "assume-yes", "y", false, "If set, skips all confirmation prompts")
	rootCmd.PersistentFlags().StringVar(&verbosity, "verbosity", "info", "Verbosity of the CLI (one of: [error, warning, info, debug])")
	rootCmd.PersistentFlags().IntVar(&cidr, "cidr", 32, "CIDR prefix length for the IP address (0-32)")
	rootCmd.Flags().BoolVarP(&showVer, "version", "v", false, "Show version")

	rootCmd.AddCommand(addCmd, removeCmd)
}

func Execute() error {
	setupLogger("info")
	return rootCmd.Execute()
}

func validateArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("requires at least 3 arguments: <project-id> <service> <resource-id> [resource-id...]\n\nSupported services:\n%s", formatSupportedServices())
	}

	service := args[1]

	if _, ok := services.GetByName(service); !ok {
		return fmt.Errorf("unsupported service: %s\n\nSupported services:\n%s", service, formatSupportedServices())
	}

	if cidr < 0 || cidr > 32 {
		return fmt.Errorf("--cidr must be between 0 and 32, got %d", cidr)
	}

	return nil
}

func formatSupportedServices() string {
	names := services.ServiceNames()
	var sb strings.Builder
	for _, s := range names {
		sb.WriteString("  " + s + "\n")
	}
	return sb.String()
}

func formatResourceLabel(name, id string) string {
	if name != "" && name != id {
		return fmt.Sprintf("%q (%q)", name, id)
	}
	return fmt.Sprintf("%q", id)
}

type action string

const (
	actionAdd    action = "add"
	actionRemove action = "remove"
)

type plannedChange struct {
	resourceID    string
	resourceLabel string
	currentACLs   []string
	updatedACLs   []string
	jsonData      []byte
}

func runACLAction(args []string, act action) error {
	if err := stackit.CheckAvailable(); err != nil {
		return err
	}

	projectID := args[0]
	service := args[1]
	resourceIDs := uniqueResourceIDs(args[2:])

	cfg, _ := services.GetByName(service)
	client := stackit.New(projectID, region)

	logStep("Fetching your external IP...")
	externalIP, err := ip.Fetch()
	if err != nil {
		return err
	}
	logInfo(fmt.Sprintf("Your external IP: %s", externalIP))

	cidrNotation := acl.ToCIDR(externalIP, cidr)
	logInfo(fmt.Sprintf("Using CIDR: %s", cidrNotation))

	preposition := "to"
	pastTense := "added"
	if act == actionRemove {
		preposition = "from"
		pastTense = "removed"
	}

	var changes []plannedChange

	for _, resourceID := range resourceIDs {
		var jsonData []byte
		if cfg.UpdateStrategy == services.PayloadStrategy {
			logStep(fmt.Sprintf("Generating payload for %s %q...", service, resourceID))
			jsonData, err = client.GeneratePayload(cfg, resourceID)
		} else {
			logStep(fmt.Sprintf("Fetching current ACLs for %s %q...", service, resourceID))
			jsonData, err = client.DescribeInstance(cfg, resourceID)
		}
		if err != nil {
			return err
		}

		resourceName := acl.ExtractName(jsonData, cfg)
		resourceLabel := formatResourceLabel(resourceName, resourceID)

		currentACLs, err := acl.ExtractACLs(jsonData, cfg)
		if err != nil {
			return err
		}

		present := acl.Contains(currentACLs, cidrNotation)

		if act == actionAdd && present {
			logWarn(fmt.Sprintf("IP %s is already in the ACL list of %s %s. No changes needed.", cidrNotation, service, resourceLabel))
			continue
		}
		if act == actionRemove && !present {
			logWarn(fmt.Sprintf("IP %s is not in the ACL list of %s %s. Nothing to remove.", cidrNotation, service, resourceLabel))
			continue
		}

		var updatedACLs []string
		if act == actionAdd {
			updatedACLs = acl.AppendCIDR(currentACLs, cidrNotation)
		} else {
			updatedACLs = acl.RemoveCIDR(currentACLs, cidrNotation)
		}

		changes = append(changes, plannedChange{
			resourceID:    resourceID,
			resourceLabel: resourceLabel,
			currentACLs:   currentACLs,
			updatedACLs:   updatedACLs,
			jsonData:      jsonData,
		})
	}

	if len(changes) == 0 {
		if len(resourceIDs) > 1 {
			logWarn("No ACL changes needed for any resource.")
		}
		return nil
	}

	for _, ch := range changes {
		logInfo(fmt.Sprintf("Updated ACLs for %s %s:", service, ch.resourceLabel))
		displayACLs := ch.updatedACLs
		if act == actionRemove {
			displayACLs = ch.currentACLs
		}
		logRaw(formatACLList(displayACLs, cidrNotation, act))
	}

	if !assumeYes {
		target := fmt.Sprintf("%s %s", service, changes[0].resourceLabel)
		if len(changes) > 1 {
			target = fmt.Sprintf("%d %s resources", len(changes), service)
		}
		prompt := fmt.Sprintf("Are you sure you want to %s %s %s the ACL of %s? (y/N)", act, cidrNotation, preposition, target)
		if !confirmPrompt(prompt) {
			logWarn("Aborted.")
			return nil
		}
	}

	var failedLabels []string
	var failedErrs []error

	for _, ch := range changes {
		logStep(fmt.Sprintf("Updating ACLs for %s %s...", service, ch.resourceLabel))

		var applyErr error
		if cfg.UpdateStrategy == services.PayloadStrategy {
			var updatedPayload []byte
			updatedPayload, applyErr = acl.SetACLs(ch.jsonData, cfg, ch.updatedACLs)
			if applyErr == nil {
				applyErr = client.UpdateClusterACL(cfg, ch.resourceID, updatedPayload)
			}
		} else {
			applyErr = client.UpdateInstanceACL(cfg, ch.resourceID, ch.updatedACLs)
		}
		if applyErr != nil {
			failedLabels = append(failedLabels, ch.resourceLabel)
			failedErrs = append(failedErrs, applyErr)
			continue
		}

		logSuccess(fmt.Sprintf("Successfully %s %s %s the ACL of %s %s.", pastTense, cidrNotation, preposition, service, ch.resourceLabel))
	}

	if len(failedLabels) > 0 {
		var sb strings.Builder
		for i, label := range failedLabels {
			sb.WriteString("  " + service + " " + label + ": " + failedErrs[i].Error() + "\n")
		}
		return fmt.Errorf("failed to update the ACL of %d of %d resources:\n%s", len(failedLabels), len(changes), sb.String())
	}

	return nil
}

func uniqueResourceIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			logWarn(fmt.Sprintf("Ignoring duplicate resource ID %q.", id))
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func confirmPrompt(prompt string) bool {
	logPrompt(prompt)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func completeArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		projects, err := stackit.ListProjects()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var completions []string
		for _, p := range projects {
			if strings.HasPrefix(p.ID, toComplete) {
				if p.Name != "" {
					completions = append(completions, p.ID+"\t"+p.Name)
				} else {
					completions = append(completions, p.ID)
				}
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	case 1:
		var completions []string
		for _, name := range services.ServiceNames() {
			if strings.HasPrefix(name, toComplete) {
				completions = append(completions, name)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	default:
		cfg, ok := services.GetByName(args[1])
		if !ok {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		reg, _ := cmd.Flags().GetString("region")
		instances, err := stackit.ListInstances(args[0], reg, cfg)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		selected := make(map[string]bool, len(args)-2)
		for _, id := range args[2:] {
			selected[id] = true
		}
		var completions []string
		for _, inst := range instances {
			if selected[inst.ID] || !strings.HasPrefix(inst.ID, toComplete) {
				continue
			}
			if inst.Name != "" {
				completions = append(completions, inst.ID+"\t"+inst.Name)
			} else {
				completions = append(completions, inst.ID)
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}
}
