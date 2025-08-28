// Copyright 2023 The go-commons Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	ocpmetadata "github.com/cloud-bulldozer/go-commons/v2/ocp-metadata"
)

var (
	kubeconfig string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ocp-metadata",
	Short: "A CLI tool for retrieving OpenShift cluster metadata",
	Long: `ocp-metadata is a CLI tool that retrieves comprehensive OpenShift cluster metadata
including cluster information, node details, networking configuration, and more.

The tool supports multiple output formats (json, yaml, table) and retrieves:
- Cluster name, platform, and type
- OCP and Kubernetes versions
- Node information (masters, workers, infra)
- Network configuration (SDN type, IPSec)
- Architecture information
- FIPS and publish configuration`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize the metadata client
		restConfig, err := getRestConfig()
		if err != nil {
			return fmt.Errorf("failed to get Kubernetes config: %w", err)
		}

		metadata, err := ocpmetadata.NewMetadata(restConfig)
		if err != nil {
			return fmt.Errorf("failed to initialize metadata client: %w", err)
		}

		// Get cluster metadata
		clusterMetadata, err := metadata.GetClusterMetadata()
		if err != nil {
			return fmt.Errorf("failed to get cluster metadata: %w", err)
		}

		return outputData(clusterMetadata)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (default is $HOME/.kube/config)")
}

// getRestConfig returns a Kubernetes REST config
func getRestConfig() (*rest.Config, error) {
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			if home := homedir.HomeDir(); home != "" {
				kubeconfig = filepath.Join(home, ".kube", "config")
			}
		}
	}

	// Try in-cluster config first
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}

	// Fall back to kubeconfig file
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

// outputData formats and prints data based on the output format
func outputData(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
