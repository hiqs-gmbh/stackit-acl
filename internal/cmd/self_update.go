package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/hiqs-gmbh/stackit-acl/internal/selfupdate"

	"github.com/spf13/cobra"
)

var (
	selfUpdateCheck   bool
	selfUpdateVersion string
	selfUpdateOutput  string
)

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update the stackit-acl binary",
	Long: `Update the stackit-acl binary from the GitHub releases.

Fetches the latest release (or the version given via --version), downloads the
archive for the current platform, verifies its SHA-256 checksum, and replaces
the running binary (or installs it at the path given via --output).`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSelfUpdate()
	},
}

func init() {
	flags := selfUpdateCmd.Flags()
	flags.BoolVar(&selfUpdateCheck, "check", false, "Check for the latest release without installing it")
	flags.StringVar(&selfUpdateVersion, "version", "", "Install the given version instead of the latest (e.g. 0.4.0 or v0.4.0)")
	flags.StringVar(&selfUpdateOutput, "output", "", "Install the binary at the given path (default: replace the running binary)")
}

func runSelfUpdate() error {
	releases := selfupdate.NewReleases()

	release, err := fetchRelease(releases)
	if err != nil {
		return err
	}

	asset, err := selfupdate.FindAsset(release)
	if err != nil {
		return err
	}

	if Version != "dev" && selfupdate.Normalize(Version) == selfupdate.Normalize(release.TagName) {
		logInfo(fmt.Sprintf("stackit-acl is up to date (%s).", currentVersionLabel()))
		return nil
	}

	target := selfUpdateOutput
	if target == "" {
		executable, exeErr := os.Executable()
		if exeErr != nil {
			return fmt.Errorf("locate the running binary: %w", exeErr)
		}
		target = executable
	}

	newVersion := selfupdate.Tag(release.TagName)
	logInfo(fmt.Sprintf("Current version: %s", currentVersionLabel()))
	logInfo(fmt.Sprintf("New version: %s", newVersion))

	if selfUpdateCheck {
		logInfo(fmt.Sprintf("Without --check this would install %s at %s.", newVersion, target))
		return nil
	}

	if err := selfupdate.VerifyAccess(target); err != nil {
		return err
	}

	if !assumeYes {
		prompt := fmt.Sprintf("Are you sure you want to update stackit-acl from %s to %s at %s? (y/N)",
			currentVersionLabel(), newVersion, target)
		if !confirmPrompt(prompt) {
			logWarn("Aborted.")
			return nil
		}
	}

	logStep(fmt.Sprintf("Downloading %s...", asset.Name))
	archive, err := releases.Download(asset.BrowserDownloadURL)
	if err != nil {
		return err
	}

	logStep("Verifying checksum...")
	if err := verifyArchive(releases, release, asset, archive); err != nil {
		return err
	}

	logStep("Installing...")
	binary, err := selfupdate.ExtractBinary(archive)
	if err != nil {
		return err
	}
	if err := selfupdate.Replace(target, binary); err != nil {
		return err
	}

	logSuccess(fmt.Sprintf("Successfully updated stackit-acl from %s to %s.", currentVersionLabel(), newVersion))
	logInfo("The new version is used by the next invocation of stackit-acl.")
	return nil
}

func fetchRelease(releases *selfupdate.Releases) (*selfupdate.Release, error) {
	if selfUpdateVersion == "" {
		logStep("Checking for the latest release...")
		release, err := releases.Latest()
		if err != nil {
			return nil, fmt.Errorf("check for the latest release: %w", err)
		}
		return release, nil
	}

	tag := selfupdate.Tag(selfUpdateVersion)
	logStep(fmt.Sprintf("Checking for release %s...", tag))
	release, err := releases.ReleaseByTag(tag)
	if err != nil {
		if errors.Is(err, selfupdate.ErrReleaseNotFound) {
			return nil, fmt.Errorf("version %s not found", tag)
		}
		return nil, fmt.Errorf("check for release %s: %w", tag, err)
	}
	return release, nil
}

func verifyArchive(releases *selfupdate.Releases, release *selfupdate.Release, asset *selfupdate.Asset, archive []byte) error {
	if digest, ok := selfupdate.DigestBytes(asset.Digest); ok {
		if !selfupdate.VerifyChecksum(archive, digest) {
			return fmt.Errorf("checksum mismatch for %s", asset.Name)
		}
		return nil
	}

	checksumsAsset := selfupdate.FindAssetByName(release, selfupdate.ChecksumsName)
	if checksumsAsset == nil {
		return fmt.Errorf("no checksum available for %s", asset.Name)
	}
	checksums, err := releases.Download(checksumsAsset.BrowserDownloadURL)
	if err != nil {
		return err
	}
	want, err := selfupdate.ParseChecksums(checksums, asset.Name)
	if err != nil {
		return err
	}
	if !selfupdate.VerifyChecksum(archive, want) {
		return fmt.Errorf("checksum mismatch for %s", asset.Name)
	}
	return nil
}
