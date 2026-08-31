# stackit-acl

A CLI tool that automatically manages your external IP address in the ACL (Access Control List) of STACKIT cloud service instances and clusters.

## What it does

1. Fetches your external IP from `https://ifconfig.schwarz`
2. Converts it to CIDR notation (default `/32`, configurable via `--cidr`)
3. Fetches the current ACLs of the specified service instance/cluster
4. Adds or removes your IP from the ACL list (skips if already in desired state)
5. Updates the service with the modified ACL list

## Prerequisites

- The [STACKIT CLI](https://github.com/stackitcloud/stackit-cli) must be installed and authenticated

## Installation

```shell
make build
# Binary is at ./bin/stackit-acl
```

## Shell Completion

`stackit-acl` provides shell completion via its built-in `completion` subcommand (powered by Cobra).

Completion is dynamic — it fetches data from the STACKIT API as you type:

1. **Project ID** — lists all projects you're a member of (with project names as descriptions)
2. **Service** — lists all supported services
3. **Resource ID** — lists all instances/clusters of the chosen service in the project (with instance names as descriptions)

### zsh

Add this to your `~/.zshrc`:

```zsh
source <(stackit-acl completion zsh)
```

Alternatively, install the completion script to a directory on your `fpath`:

```zsh
mkdir -p ~/.zsh/completions
stackit-acl completion zsh > ~/.zsh/completions/_stackit-acl
```

Then ensure these are in your `~/.zshrc` (before `compinit`):

```zsh
fpath+=~/.zsh/completions
autoload -Uz compinit && compinit
```

### bash

Add this to your `~/.bashrc` (or `~/.bash_profile`):

```bash
source <(stackit-acl completion bash)
```

Restart your shell or re-source your config file after making changes.

## Usage

```
stackit-acl <command> <project-id> <service> <resource-id> [flags]
```

### Commands

| Command | Description |
|---|---|
| `add` | Add your external IP to the ACL |
| `remove` | Remove your external IP from the ACL |

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
```
