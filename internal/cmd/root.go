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
  stackit-acl <command> <project-id> <service> <resource-id> [flags]

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
			fmt.Println(Version)
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
	if len(args) != 3 {
		return fmt.Errorf("requires exactly 3 arguments: <project-id> <service> <resource-id>\n\nSupported services:\n%s", formatSupportedServices())
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

func runACLAction(args []string, act action) error {
	if err := stackit.CheckAvailable(); err != nil {
		return err
	}

	projectID := args[0]
	service := args[1]
	resourceID := args[2]

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

	var updatedACLs []string
	var preposition string

	if act == actionAdd {
		preposition = "to"
		if present {
			logWarn(fmt.Sprintf("IP %s is already in the ACL list. No changes needed.", cidrNotation))
			return nil
		}
		updatedACLs = acl.AppendCIDR(currentACLs, cidrNotation)
	} else {
		preposition = "from"
		if !present {
			logWarn(fmt.Sprintf("IP %s is not in the ACL list. Nothing to remove.", cidrNotation))
			return nil
		}
		updatedACLs = acl.RemoveCIDR(currentACLs, cidrNotation)
	}

	logInfo("Updated ACLs:")
	displayACLs := updatedACLs
	if act == actionRemove {
		displayACLs = currentACLs
	}
	fmt.Print(formatACLList(displayACLs, cidrNotation, act))

	if !assumeYes {
		prompt := fmt.Sprintf("Are you sure you want to %s %s %s the ACL of %s %s? (y/N)", act, cidrNotation, preposition, service, resourceLabel)
		if !confirmPrompt(prompt) {
			logWarn("Aborted.")
			return nil
		}
	}

	logStep(fmt.Sprintf("Updating ACLs for %s %s...", service, resourceLabel))
	if cfg.UpdateStrategy == services.PayloadStrategy {
		updatedPayload, err := acl.SetACLs(jsonData, cfg, updatedACLs)
		if err != nil {
			return err
		}
		if err = client.UpdateClusterACL(cfg, resourceID, updatedPayload); err != nil {
			return err
		}
	} else {
		if err = client.UpdateInstanceACL(cfg, resourceID, updatedACLs); err != nil {
			return err
		}
	}

	pastTense := "added"
	if act == actionRemove {
		pastTense = "removed"
	}
	logSuccess(fmt.Sprintf("Successfully %s %s %s the ACL of %s %s.", pastTense, cidrNotation, preposition, service, resourceLabel))
	return nil
}

func confirmPrompt(prompt string) bool {
	outPrompt(prompt)
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
	case 2:
		cfg, ok := services.GetByName(args[1])
		if !ok {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		reg, _ := cmd.Flags().GetString("region")
		instances, err := stackit.ListInstances(args[0], reg, cfg)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var completions []string
		for _, inst := range instances {
			if strings.HasPrefix(inst.ID, toComplete) {
				if inst.Name != "" {
					completions = append(completions, inst.ID+"\t"+inst.Name)
				} else {
					completions = append(completions, inst.ID)
				}
			}
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}
