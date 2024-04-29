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
	"fmt"
)

// NewIndexer creates a new Indexer with the specified IndexerConfig
func NewIndexer(indexerConfig IndexerConfig) (*Indexer, error) {
	var indexer Indexer
	var err error
	switch indexerConfig.Type {
	case LocalIndexer:
		indexer, err = NewLocalIndexer(indexerConfig)
	case ElasticIndexer:
		indexer, err = NewElasticIndexer(indexerConfig)
	case OpenSearchIndexer:
		indexer, err = NewOpenSearchIndexer(indexerConfig)
	default:
		return &indexer, fmt.Errorf("Indexer not found: %s", indexerConfig.Type)
	}
	return &indexer, err
}
