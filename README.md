# stackit-acl

A CLI tool that automatically manages your external IP address in the ACL (Access Control List) of STACKIT cloud service instances and clusters.

## What it does

1. Fetches your external IP (from `https://ifconfig.schwarz`, falling back to `https://checkip.amazonaws.com` and `https://v4.ident.me`)
2. Converts it to CIDR notation (default `/32`, configurable via `--cidr`)
3. Fetches the current ACLs of the specified service instance/cluster
4. Adds or removes your IP from the ACL list (skips if already in desired state)
5. Updates the service with the modified ACL list

## Prerequisites

- The [STACKIT CLI](https://github.com/stackitcloud/stackit-cli) must be installed and authenticated

## Installation

Download the latest binary for your platform from the [GitHub releases](https://github.com/hiqs-gmbh/stackit-acl/releases) (checksums are published as `checksums.txt`):

| Platform | Archive |
|---|---|
| macOS (Apple Silicon) | `stackit-acl_X.Y.Z_darwin_arm64.tar.gz` |
| macOS (Intel) | `stackit-acl_X.Y.Z_darwin_amd64.tar.gz` |
| Linux (x86_64) | `stackit-acl_X.Y.Z_linux_amd64.tar.gz` |
| Linux (ARM64) | `stackit-acl_X.Y.Z_linux_arm64.tar.gz` |

Or install with Go:

```shell
go install github.com/hiqs-gmbh/stackit-acl@latest
```

To update an existing installation, run `stackit-acl self-update`.

Or build from source:

```shell
make build && mv bin/stackit-acl ~/.local/bin
```

## Shell Completion

`stackit-acl` provides shell completion via its built-in `completion` subcommand (powered by Cobra).

Completion is dynamic — it fetches data from the STACKIT API as you type:

1. **Project ID** — lists all projects you're a member of (with project names as descriptions)
2. **Service** — lists all supported services
3. **Resource ID(s)** — lists all instances/clusters of the chosen service in the project (with instance names as descriptions; already-selected IDs are excluded)

If the STACKIT CLI is not authenticated (e.g. your access token has expired), the completion shows a hint to run `stackit auth login` instead of failing silently. Other errors are shown with a short reason.

### zsh

Add this to your `~/.zshrc`:

```zsh
source <(stackit-acl completion zsh)
```

### bash

Add this to your `~/.bashrc` (or `~/.bash_profile`):

```bash
source <(stackit-acl completion bash)
```

Restart your shell or re-source your config file after making changes.

## Usage

```
stackit-acl <command> <project-id> <service> <resource-id> [resource-id...] [flags]
```

Multiple resource IDs can be passed to apply the same change to all of them at once (bulk mode). The tool fetches your external IP once, shows the planned changes for every resource, and asks for a single confirmation before updating. Resources that are already in the desired state are skipped; duplicates are ignored.

The `list` command is read-only: it shows the currently applied ACLs without fetching your external IP or asking for confirmation. Resource IDs are optional — when none is passed, the ACLs of all instances of the given service in the project are listed.

### Commands

| Command | Description |
|---|---|
| `add` | Add your external IP to the ACL |
| `list` | List the currently applied ACLs |
| `remove` | Remove your external IP from the ACL |
| `self-update` | Update the stackit-acl binary |

### Supported services

| Service | Example |
|---|---|
| mongodbflex | `stackit-acl add <PROJECT_ID> mongodbflex <INSTANCE_ID>` |
| postgresflex | `stackit-acl add <PROJECT_ID> postgresflex <INSTANCE_ID>` |
| sqlserverflex | `stackit-acl add <PROJECT_ID> sqlserverflex <INSTANCE_ID>` |
| redis | `stackit-acl add <PROJECT_ID> redis <INSTANCE_ID>` |
| valkey | `stackit-acl add <PROJECT_ID> valkey <INSTANCE_ID>` |
| opensearch | `stackit-acl add <PROJECT_ID> opensearch <INSTANCE_ID>` |
| rabbitmq | `stackit-acl add <PROJECT_ID> rabbitmq <INSTANCE_ID>` |
| mariadb | `stackit-acl add <PROJECT_ID> mariadb <INSTANCE_ID>` |
| logme | `stackit-acl add <PROJECT_ID> logme <INSTANCE_ID>` |
| ske | `stackit-acl add <PROJECT_ID> ske <CLUSTER_NAME>` |

### Flags

| Flag | Shorthand | Default | Description |
|---|---|---|---|
| `--region` | | | Target region for region-specific requests |
| `--assume-yes` | `-y` | `false` | Skip confirmation prompts |
| `--verbosity` | | `info` | `error`, `warning`, `info`, `debug` |
| `--cidr` | | `32` | CIDR prefix length (0-32) |
| `--version` | `-v` | | Show version |

### Self-update

`self-update` updates the installed binary from the [GitHub releases](https://github.com/hiqs-gmbh/stackit-acl/releases). It shows the current and the new version, asks for confirmation (unless `--assume-yes` is set), verifies the SHA-256 checksum of the download, and replaces the binary atomically:

| Flag | Default | Description |
|---|---|---|
| `--check` | `false` | Check for the latest release without installing it |
| `--version` | | Install a specific version (e.g. `0.4.0` or `v0.4.0`) instead of the latest |
| `--output` | | Install the binary at the given path (default: replace the running binary) |

### Examples

```shell
# Add your IP to a MongoDB Flex instance
stackit-acl add 00000000-0000-0000-0000-000000000000 mongodbflex abc-123-def

# Remove your IP from an SKE cluster
stackit-acl remove 00000000-0000-0000-0000-000000000000 ske my-cluster

# Add your IP to an SKE cluster with a /24 prefix
stackit-acl add 00000000-0000-0000-0000-000000000000 ske my-cluster --cidr 24

# Add your IP to a Redis instance in a specific region, skipping confirmation
stackit-acl add 00000000-0000-0000-0000-000000000000 redis abc-123-def --region eu01 -y

# Remove your IP from a PostgreSQL Flex instance
stackit-acl remove 00000000-0000-0000-0000-000000000000 postgresflex abc-123-def

# Add your IP to multiple PostgreSQL Flex instances at once
stackit-acl add 00000000-0000-0000-0000-000000000000 postgresflex abc-123-def abc-456-ghi

# Remove your IP from multiple Redis instances, skipping confirmation
stackit-acl remove 00000000-0000-0000-0000-000000000000 redis abc-123-def abc-456-ghi -y

# List the currently applied ACLs of a PostgreSQL Flex instance
stackit-acl list 00000000-0000-0000-0000-000000000000 postgresflex abc-123-def

# List the currently applied ACLs of all Redis instances in the project
stackit-acl list 00000000-0000-0000-0000-000000000000 redis

# Check for a newer version without installing it
stackit-acl self-update --check

# Update to a specific version, skipping confirmation
stackit-acl self-update --version v0.4.0 -y
```
