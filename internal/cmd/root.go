package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"stackit-acl/internal/acl"
	"stackit-acl/internal/ip"
	"stackit-acl/internal/services"
	"stackit-acl/internal/stackit"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

var (
	projectID string
	region    string
	assumeYes bool
	verbosity string
	cidr      int
	showVer   bool
)

var rootCmd = &cobra.Command{
	Use:   "stackit-acl <service> <resource-group> <resource-id>",
	Short: "Automatically adds your external IP to STACKIT service ACLs",
	Long: `Automatically adds your external IP to STACKIT service ACLs.

This tool fetches your external IP address and appends it to the ACL list
of the specified STACKIT service instance or cluster.

The stackit CLI must be installed and authenticated.

Usage:
  stackit-acl <service> <resource-group> <resource-id> [flags]

Supported services:
  mongodbflex instance <INSTANCE_ID>
  postgresflex instance <INSTANCE_ID>
  sqlserverflex instance <INSTANCE_ID>
  redis instance <INSTANCE_ID>
  valkey instance <INSTANCE_ID>
  opensearch instance <INSTANCE_ID>
  rabbitmq instance <INSTANCE_ID>
  mariadb instance <INSTANCE_ID>
  logme instance <INSTANCE_ID>
  ske cluster <CLUSTER_NAME>`,
	Args: validateArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVarP(&projectID, "project-id", "p", "", "Project ID (required)")
	rootCmd.Flags().StringVar(&region, "region", "", "Target region for region-specific requests")
	rootCmd.Flags().BoolVarP(&assumeYes, "assume-yes", "y", false, "If set, skips all confirmation prompts")
	rootCmd.Flags().StringVar(&verbosity, "verbosity", "info", "Verbosity of the CLI (one of: [error, warning, info, debug])")
	rootCmd.Flags().IntVar(&cidr, "cidr", 32, "CIDR prefix length for the IP address (0-32)")
	rootCmd.Flags().BoolVarP(&showVer, "version", "v", false, "Show version")

}

func Execute() error {
	return rootCmd.Execute()
}

func validateArgs(cmd *cobra.Command, args []string) error {
	if showVer {
		return nil
	}
	if projectID == "" {
		return fmt.Errorf("required flag(s) \"project-id\" not set")
	}
	if len(args) != 3 {
		return fmt.Errorf("requires exactly 3 arguments: <service> <resource-group> <resource-id>\n\nSupported services:\n%s", formatSupportedServices())
	}

	service := args[0]
	resourceGroup := args[1]

	if _, ok := services.Get(service, resourceGroup); !ok {
		return fmt.Errorf("unsupported service/resource-group: %s %s\n\nSupported services:\n%s", service, resourceGroup, formatSupportedServices())
	}

	if cidr < 0 || cidr > 32 {
		return fmt.Errorf("--cidr must be between 0 and 32, got %d", cidr)
	}

	return nil
}

func formatSupportedServices() string {
	supported := services.Supported()
	var sb strings.Builder
	for _, s := range supported {
		sb.WriteString("  " + s + "\n")
	}
	return sb.String()
}

func run(cmd *cobra.Command, args []string) error {
	if showVer {
		fmt.Println(Version)
		return nil
	}

	if err := stackit.CheckAvailable(); err != nil {
		return err
	}

	service := args[0]
	resourceGroup := args[1]
	resourceID := args[2]

	cfg, _ := services.Get(service, resourceGroup)

	client := stackit.New(projectID, region)

	log(verbosity, "info", "Fetching your external IP...")
	externalIP, err := ip.Fetch()
	if err != nil {
		return err
	}
	log(verbosity, "info", fmt.Sprintf("Your external IP: %s", externalIP))

	cidrNotation := acl.ToCIDR(externalIP, cidr)
	log(verbosity, "info", fmt.Sprintf("Using CIDR: %s", cidrNotation))

	var jsonData []byte
	if cfg.UpdateStrategy == services.PayloadStrategy {
		log(verbosity, "info", fmt.Sprintf("Generating payload for %s %s %q...", service, resourceGroup, resourceID))
		jsonData, err = client.GeneratePayload(cfg, resourceID)
	} else {
		log(verbosity, "info", fmt.Sprintf("Fetching current ACLs for %s %s %q...", service, resourceGroup, resourceID))
		jsonData, err = client.DescribeInstance(cfg, resourceID)
	}
	if err != nil {
		return err
	}

	currentACLs, err := acl.ExtractACLs(jsonData, cfg)
	if err != nil {
		return err
	}

	if len(currentACLs) > 0 {
		log(verbosity, "info", fmt.Sprintf("Current ACLs: %s", strings.Join(currentACLs, ", ")))
	} else {
		log(verbosity, "info", "Current ACLs: (none)")
	}

	for _, c := range currentACLs {
		if c == cidrNotation {
			log(verbosity, "info", fmt.Sprintf("IP %s is already in the ACL list. No changes needed.", cidrNotation))
			return nil
		}
	}

	updatedACLs := acl.AppendCIDR(currentACLs, cidrNotation)
	log(verbosity, "info", fmt.Sprintf("Updated ACLs: %s", strings.Join(updatedACLs, ", ")))

	if !assumeYes {
		prompt := fmt.Sprintf("Are you sure you want to add %s to the ACL of %s %s %q? (y/N)", cidrNotation, service, resourceGroup, resourceID)
		if !confirmPrompt(prompt) {
			log(verbosity, "info", "Aborted.")
			return nil
		}
	}

	log(verbosity, "info", fmt.Sprintf("Updating ACLs for %s %s %q...", service, resourceGroup, resourceID))
	if cfg.UpdateStrategy == services.PayloadStrategy {
		updatedPayload, err := acl.SetACLs(jsonData, cfg, updatedACLs)
		if err != nil {
			return err
		}
		err = client.UpdateClusterACL(cfg, resourceID, updatedPayload)
		if err != nil {
			return err
		}
	} else {
		err = client.UpdateInstanceACL(cfg, resourceID, updatedACLs)
		if err != nil {
			return err
		}
	}

	log(verbosity, "info", fmt.Sprintf("Successfully added %s to the ACL of %s %s %q.", cidrNotation, service, resourceGroup, resourceID))
	return nil
}

var verbosityLevels = map[string]int{
	"error":   0,
	"warning": 1,
	"info":    2,
	"debug":   3,
}

func log(currentVerbosity, msgLevel, msg string) {
	cl, ok := verbosityLevels[currentVerbosity]
	if !ok {
		cl = 2
	}
	ml, ok := verbosityLevels[msgLevel]
	if !ok {
		ml = 2
	}
	if ml <= cl {
		fmt.Println(msg)
	}
}

func confirmPrompt(prompt string) bool {
	fmt.Print(prompt + " ")
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
