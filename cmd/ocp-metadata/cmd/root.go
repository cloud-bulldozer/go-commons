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

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	ocpmetadata "github.com/cloud-bulldozer/go-commons/v2/ocp-metadata"
)

var (
	kubeconfig   string
	outputFormat string
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

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (default is $HOME/.kube/config)")
	rootCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (json|yaml|table)")
}

// getRestConfig returns a Kubernetes REST config
func getRestConfig() (*rest.Config, error) {
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
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
	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	case "yaml":
		encoder := yaml.NewEncoder(os.Stdout)
		defer encoder.Close()
		return encoder.Encode(data)
	case "table":
		return outputTable(data)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// outputTable handles table format output for different data types
func outputTable(data interface{}) error {
	switch v := data.(type) {
	case ocpmetadata.ClusterMetadata:
		return printClusterMetadataTable(v)
	case map[string]interface{}:
		return printMapTable(v)
	case string:
		fmt.Println(v)
		return nil
	default:
		// Fall back to JSON for complex types
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	}
}

// printClusterMetadataTable prints cluster metadata in table format
func printClusterMetadataTable(metadata ocpmetadata.ClusterMetadata) error {
	fmt.Printf("%-20s %s\n", "Field", "Value")
	fmt.Printf("%-20s %s\n", "-----", "-----")
	fmt.Printf("%-20s %s\n", "Cluster Name", metadata.ClusterName)
	fmt.Printf("%-20s %s\n", "Platform", metadata.Platform)
	fmt.Printf("%-20s %s\n", "Cluster Type", metadata.ClusterType)
	fmt.Printf("%-20s %s\n", "OCP Version", metadata.OCPVersion)
	fmt.Printf("%-20s %s\n", "OCP Major Version", metadata.OCPMajorVersion)
	fmt.Printf("%-20s %s\n", "K8s Version", metadata.K8SVersion)
	fmt.Printf("%-20s %s\n", "Region", metadata.Region)
	fmt.Printf("%-20s %s\n", "SDN Type", metadata.SDNType)
	fmt.Printf("%-20s %d\n", "Total Nodes", metadata.TotalNodes)
	fmt.Printf("%-20s %d (%s)\n", "Master Nodes", metadata.MasterNodesCount, metadata.MasterNodesType)
	fmt.Printf("%-20s %d (%s)\n", "Worker Nodes", metadata.WorkerNodesCount, metadata.WorkerNodesType)
	fmt.Printf("%-20s %d (%s)\n", "Infra Nodes", metadata.InfraNodesCount, metadata.InfraNodesType)
	fmt.Printf("%-20s %d\n", "Other Nodes", metadata.OtherNodesCount)
	fmt.Printf("%-20s %s\n", "Worker Arch", metadata.WorkerArch)
	fmt.Printf("%-20s %s\n", "Control Plane Arch", metadata.ControlPlaneArch)
	fmt.Printf("%-20s %t\n", "FIPS", metadata.Fips)
	fmt.Printf("%-20s %s\n", "Publish", metadata.Publish)
	fmt.Printf("%-20s %t\n", "IPSec", metadata.Ipsec)
	fmt.Printf("%-20s %s\n", "IPSec Mode", metadata.IpsecMode)
	return nil
}

// printMapTable prints a map in table format
func printMapTable(data map[string]interface{}) error {
	fmt.Printf("%-20s %s\n", "Field", "Value")
	fmt.Printf("%-20s %s\n", "-----", "-----")
	for key, value := range data {
		fmt.Printf("%-20s %v\n", key, value)
	}
	return nil
}
