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

package indexers

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

// Local indexer instance
type Local struct {
	metricsDirectory string
}

// NewLocalIndexer returns a new Local Indexer
func NewLocalIndexer(indexerConfig IndexerConfig) (*Local, error) {
	var localIndexer Local
	if indexerConfig.MetricsDirectory == "" {
		return &localIndexer, fmt.Errorf("directory name not specified")
	}
	localIndexer.metricsDirectory = indexerConfig.MetricsDirectory
	err := os.MkdirAll(localIndexer.metricsDirectory, 0744)
	return &localIndexer, err
}

// Index uses generates a local file with the given name and metrics
func (l *Local) Index(documents []interface{}, opts IndexingOpts) (string, error) {
	if len(documents) == 0 {
		return "", fmt.Errorf("Empty document list in %v", opts.MetricName)
	}
	if opts.MetricName == "" {
		return "", fmt.Errorf("MetricName shouldn't be empty")
	}
	metricName := fmt.Sprintf("%s.json", opts.MetricName)
	filename := path.Join(l.metricsDirectory, metricName)
	if content, err := os.ReadFile(filename); err == nil {
		var existingDocs []interface{}
		if err := json.Unmarshal(content, &existingDocs); err != nil {
			return "", fmt.Errorf("JSON decoding error in %s: %s", filename, err)
		}
		documents = append(existingDocs, documents...)
	}

	content, err := json.Marshal(documents)
	if err != nil {
		return "", fmt.Errorf("JSON encoding error: %s", err)
	}
	if err := os.WriteFile(filename, content, 0644); err != nil {
		return "", fmt.Errorf("Error writing metrics file %s: %s", filename, err)
	}
	return fmt.Sprintf("File %s now contains %d documents", filename, len(documents)), nil
}
