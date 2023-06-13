package indexers

import (
	"fmt"
	"testing"
)

func TestElasticnew(t *testing.T) {
	var indexer Elastic
	indexer.index = "go-commons-test"
	tests := []struct {
		name          string
		indexerConfig IndexerConfig
		expectedError string
		wantErr       bool
	}{
		{"Test 1",
			IndexerConfig{Type: "elastic",
				Servers:            []string{},
				Index:              "",
				InsecureSkipVerify: true,
			},
			"",
			true,
		},
		{"Test 1",
			IndexerConfig{Type: "elastic",
				Servers:            []string{},
				Index:              "",
				InsecureSkipVerify: true,
			},
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := indexer.new(tt.indexerConfig)
			fmt.Printf("hi elastic test %v\n", err)

			if (err != nil) == tt.wantErr {
				//t.Errorf("NewIndexer() error: %v, wantErr %v\n", err, tt.wantErr)
				return
			}
			expectedError := tt.expectedError
			t.Errorf("Expected err: %v, Actual Error: %v", expectedError, err)

		})
	}

}
