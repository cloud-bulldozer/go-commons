# OCP Metadata CLI Tool

A command-line interface for retrieving OpenShift cluster metadata information.

## Overview

The `ocp-metadata` CLI tool provides easy access to comprehensive OpenShift cluster metadata including cluster information, node details, networking configuration, and more. It's built on top of the `ocp-metadata` Go package and provides a simple, focused interface for extracting cluster information.

## Features

- **Comprehensive Cluster Metadata**: Retrieve detailed cluster information including version, platform, nodes, and configuration
- **Multiple Output Formats**: Support for JSON, YAML, and table formats
- **Flexible Authentication**: Support for kubeconfig files and in-cluster authentication
- **Simple Interface**: Single command to get all cluster metadata

## Installation

### Build from Source

```bash
# From the go-commons project root
go build -o bin/ocp-metadata ./cmd/ocp-metadata

# Optionally install to your PATH
sudo cp bin/ocp-metadata /usr/local/bin/
```

## Usage

### Basic Commands

```bash
# Get help
ocp-metadata --help

# Get cluster metadata in table format (default)
ocp-metadata

# Get cluster metadata in JSON format
ocp-metadata --output json

# Get cluster metadata in YAML format
ocp-metadata --output yaml
```

### Authentication

The CLI tool supports multiple authentication methods:

1. **Kubeconfig file** (default: `~/.kube/config`):
   ```bash
   ocp-metadata --kubeconfig /path/to/kubeconfig
   ```

2. **In-cluster authentication** (when running inside a pod):
   ```bash
   ocp-metadata
   ```

### Output Formats

The tool supports three output formats:

- **table** (default): Human-readable table format
- **json**: JSON format for programmatic consumption
- **yaml**: YAML format for configuration files

```bash
# Table format (default)
ocp-metadata

# JSON format
ocp-metadata -o json

# YAML format
ocp-metadata -o yaml
```

## Example Output

### Cluster Metadata (Table Format)

```
Field                Value
-----                -----
Cluster Name         my-ocp-cluster
Platform             aws
Cluster Type         self-managed
OCP Version          4.14.0
OCP Major Version    4.14
K8s Version          v1.27.0+abc123
Region               us-east-1
SDN Type             OVNKubernetes
Total Nodes          6
Master Nodes         3 (m5.xlarge)
Worker Nodes         3 (m5.large)
Infra Nodes          0 ()
Other Nodes          0
Worker Arch          amd64
Control Plane Arch   amd64
FIPS                 false
Publish              External
IPSec                false
IPSec Mode           Disabled
```

### Cluster Metadata (JSON Format)

```json
{
  "platform": "aws",
  "clusterType": "self-managed",
  "ocpVersion": "4.14.0",
  "ocpMajorVersion": "4.14",
  "k8sVersion": "v1.27.0+abc123",
  "masterNodesType": "m5.xlarge",
  "workerNodesType": "m5.large",
  "masterNodesCount": 3,
  "workerNodesCount": 3,
  "infraNodesCount": 0,
  "otherNodesCount": 0,
  "totalNodes": 6,
  "sdnType": "OVNKubernetes",
  "clusterName": "my-ocp-cluster",
  "region": "us-east-1",
  "fips": false,
  "publish": "External",
  "workerArch": "amd64",
  "controlPlaneArch": "amd64",
  "ipsec": false,
  "ipsecMode": "Disabled"
}
```

## Requirements

- Go 1.23+
- Access to an OpenShift cluster
- Valid kubeconfig or in-cluster authentication

## Error Handling

The CLI tool provides meaningful error messages for common issues:

- **Authentication errors**: Issues with kubeconfig or cluster access
- **Resource not found**: When cluster resources are unavailable
- **Permission errors**: When service account lacks required permissions
- **Network errors**: When cluster is unreachable

## Contributing

This CLI tool is part of the go-commons project. Please refer to the main project README for contribution guidelines.

## License

Licensed under the Apache License, Version 2.0. See the LICENSE file for details.
