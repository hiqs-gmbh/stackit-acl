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
stackit-acl <command> <service> <resource-group> <resource-id> [flags]
```

### Commands

| Command | Description |
|---|---|
| `add` | Add your external IP to the ACL |
| `remove` | Remove your external IP from the ACL |

### Supported services

| Service | Resource Group | Example |
|---|---|---|
| mongodbflex | instance | `stackit-acl add mongodbflex instance <INSTANCE_ID> -p <PROJECT_ID>` |
| postgresflex | instance | `stackit-acl add postgresflex instance <INSTANCE_ID> -p <PROJECT_ID>` |
| sqlserverflex | instance | `stackit-acl add sqlserverflex instance <INSTANCE_ID> -p <PROJECT_ID>` |
| redis | instance | `stackit-acl add redis instance <INSTANCE_ID> -p <PROJECT_ID>` |
| valkey | instance | `stackit-acl add valkey instance <INSTANCE_ID> -p <PROJECT_ID>` |
| opensearch | instance | `stackit-acl add opensearch instance <INSTANCE_ID> -p <PROJECT_ID>` |
| rabbitmq | instance | `stackit-acl add rabbitmq instance <INSTANCE_ID> -p <PROJECT_ID>` |
| mariadb | instance | `stackit-acl add mariadb instance <INSTANCE_ID> -p <PROJECT_ID>` |
| logme | instance | `stackit-acl add logme instance <INSTANCE_ID> -p <PROJECT_ID>` |
| ske | cluster | `stackit-acl add ske cluster <CLUSTER_NAME> -p <PROJECT_ID>` |

### Flags

| Flag | Shorthand | Default | Description |
|---|---|---|---|
| `--project-id` | `-p` | | Project ID (required) |
| `--region` | | | Target region for region-specific requests |
| `--assume-yes` | `-y` | `false` | Skip confirmation prompts |
| `--verbosity` | | `info` | `error`, `warning`, `info`, `debug` |
| `--cidr` | | `32` | CIDR prefix length (0-32) |
| `--version` | `-v` | | Show version |

### Examples

```shell
# Add your IP to a MongoDB Flex instance
stackit-acl add mongodbflex instance abc-123-def -p 00000000-0000-0000-0000-000000000000

# Remove your IP from an SKE cluster
stackit-acl remove ske cluster my-cluster -p 00000000-0000-0000-0000-000000000000

# Add your IP to an SKE cluster with a /24 prefix
stackit-acl add ske cluster my-cluster -p 00000000-0000-0000-0000-000000000000 --cidr 24

# Add your IP to a Redis instance in a specific region, skipping confirmation
stackit-acl add redis instance abc-123-def -p 00000000-0000-0000-0000-000000000000 --region eu01 -y

# Remove your IP from a PostgreSQL Flex instance
stackit-acl remove postgresflex instance abc-123-def -p 00000000-0000-0000-0000-000000000000
```
