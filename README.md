# stackit-acl

A CLI tool that automatically adds your external IP address to the ACL (Access Control List) of STACKIT cloud service instances and clusters.

## What it does

1. Fetches your external IP from `https://ifconfig.schwarz`
2. Converts it to CIDR notation (default `/32`, configurable via `--cidr`)
3. Fetches the current ACLs of the specified service instance/cluster
4. Appends your IP (skips if already present)
5. Updates the service with the combined ACL list

## Prerequisites

- The [STACKIT CLI](https://github.com/stackitcloud/stackit-cli) must be installed and authenticated

## Installation

```shell
make build
# Binary is at ./bin/stackit-acl
```

## Usage

```
stackit-acl <service> <resource-group> <resource-id> [flags]
```

### Supported services

| Service | Resource Group | Example |
|---|---|---|
| mongodbflex | instance | `stackit-acl mongodbflex instance <INSTANCE_ID> -p <PROJECT_ID>` |
| postgresflex | instance | `stackit-acl postgresflex instance <INSTANCE_ID> -p <PROJECT_ID>` |
| sqlserverflex | instance | `stackit-acl sqlserverflex instance <INSTANCE_ID> -p <PROJECT_ID>` |
| redis | instance | `stackit-acl redis instance <INSTANCE_ID> -p <PROJECT_ID>` |
| valkey | instance | `stackit-acl valkey instance <INSTANCE_ID> -p <PROJECT_ID>` |
| opensearch | instance | `stackit-acl opensearch instance <INSTANCE_ID> -p <PROJECT_ID>` |
| rabbitmq | instance | `stackit-acl rabbitmq instance <INSTANCE_ID> -p <PROJECT_ID>` |
| mariadb | instance | `stackit-acl mariadb instance <INSTANCE_ID> -p <PROJECT_ID>` |
| logme | instance | `stackit-acl logme instance <INSTANCE_ID> -p <PROJECT_ID>` |
| ske | cluster | `stackit-acl ske cluster <CLUSTER_NAME> -p <PROJECT_ID>` |

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
stackit-acl mongodbflex instance abc-123-def -p 00000000-0000-0000-0000-000000000000

# Add your IP to an SKE cluster with a /24 prefix
stackit-acl ske cluster my-cluster -p 00000000-0000-0000-0000-000000000000 --cidr 24

# Add your IP to a Redis instance in a specific region, skipping confirmation
stackit-acl redis instance abc-123-def -p 00000000-0000-0000-0000-000000000000 --region eu01 -y
```
