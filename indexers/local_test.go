package indexers

import (
	"fmt"
	"testing"
)

func TestLocalnew(t *testing.T) {
	tests := []struct {
		name          string
		indexerConfig IndexerConfig
		wantErr       bool
	}{
		{
			"Test 1",
			IndexerConfig{Type: "elastic",
				Servers:            []string{""},
				Index:              "go-commons-test",
				InsecureSkipVerify: true,
				MetricsDirectory:   "",
			},
			true,
		},

		{
			"Test 2",
			IndexerConfig{Type: "elastic",
				Servers:            []string{""},
				Index:              "go-commons-test",
				InsecureSkipVerify: true,
				MetricsDirectory:   "placeholder",
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var indexer Local
			err := indexer.new(tt.indexerConfig)

			if (err != nil) != tt.wantErr {
				fmt.Printf("Expected Error %v, Actual Error %v", tt.wantErr, err)
				return
			}
		})
	}
}

func TestLocalIndex(t *testing.T) {
	tests := []struct {
		name          string
		documents     []interface{}
		opts          IndexingOpts
		expectedError string
		wantErr       bool
	}{
		{name: "test1",
			documents: []interface{}{
				"example document",
				42,
				3.14,
				false,
				struct {
					Name string
					Age  int
				}{
					Name: "John Doe",
					Age:  25,
				},
				map[string]interface{}{
					"key1": "value1",
					"key2": 123,
					"key3": true,
				}},
			opts: IndexingOpts{
				MetricName: "placeholder",
				JobName:    "",
			},
			expectedError: "",
			wantErr:       false,
		},

		{name: "test2",
			documents: []interface{}{
				"example document",
				42,
				3.14,
				false,
				struct {
					Name string
					Age  int
				}{
					Name: "John Doe",
					Age:  25,
				},
				map[string]interface{}{
					"key1": "value1",
					"key2": 123,
					"key3": true,
				}},
			opts: IndexingOpts{
				MetricName: "placeholder",
				JobName:    "placeholder",
			},
			expectedError: "",
			wantErr:       false,
		},

		{name: "test3",
			documents: []interface{}{
				"example document",
				42,
				3.14,
				false,
				struct {
					Name string
					Age  int
				}{
					Name: "John Doe",
					Age:  25,
				},
				map[string]interface{}{
					"key1": "value1",
					"key2": 123,
					"key3": true,
				}},
			opts: IndexingOpts{
				MetricName: "placeholder",
				JobName:    "placeholder",
			},
			expectedError: "",
			wantErr:       true,
		},

		{name: "test4",
			documents: []interface{}{
				make(chan string),
				"example document",
				42,
				3.14,
				false,
				struct {
					Name string
					Age  int
				}{
					Name: "John Doe",
					Age:  25,
				},
				map[string]interface{}{
					"key1": "value1",
					"key2": 123,
					"key3": true,
				}},
			opts: IndexingOpts{
				MetricName: "placeholder",
				JobName:    "placeholder",
			},
			expectedError: "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var indexer Local
			indexer.metricsDirectory = "placeholder"
			if tt.name == "test3" {
				indexer.metricsDirectory = "abc"
			}
			_, err := indexer.Index(tt.documents, tt.opts)

			if (err != nil) != tt.wantErr {
				fmt.Printf("Expected Error %v, Actual Error %v", tt.wantErr, err)
				return
			}
		})
	}
}
