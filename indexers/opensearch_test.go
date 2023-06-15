package indexers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestOpenSearchnew(t *testing.T) {
	var indexer OpenSearch
	payload := []byte(`{
		"name" : "0bcd132328f2f0c8ee451d471960750e",
		"cluster_name" : "415909267177:perfscale-dev",
		"cluster_uuid" : "Xz2IU4etSieAeaO2j-QCUw",
		"version" : {
		  "number" : "7.10.2",
		  "build_type" : "tar",
		  "build_hash" : "unknown",
		  "build_date" : "2023-03-22T14:16:51.874273Z",
		  "build_snapshot" : false,
		  "lucene_version" : "9.3.0",
		  "minimum_wire_compatibility_version" : "7.10.0",
		  "minimum_index_compatibility_version" : "7.0.0"
		},
		"tagline" : "The OpenSearch Project: https://opensearch.org/"
	  }`)
	indexer.index = "go-commons-test"
	tests := []struct {
		name          string
		indexerConfig IndexerConfig
		expectedError string
		wantErr       bool
		mockServer    *httptest.Server
	}{
		{"Test 1",
			IndexerConfig{Type: "openSearch",
				Servers:            []string{},
				Index:              "",
				InsecureSkipVerify: true,
			},
			"",
			true,
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(payload)
			})),
		},

		{"Test 2",
			IndexerConfig{Type: "opensearch",
				Servers:            []string{},
				Index:              "go-commons",
				InsecureSkipVerify: false,
			},
			"",
			true,
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			})),
		},

		{"Test 3",
			IndexerConfig{Type: "opensearch",
				Servers:            []string{},
				Index:              "go-commons",
				InsecureSkipVerify: false,
			},
			"",
			true,
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusGatewayTimeout)
			})),
		},

		{"Test 4",
			IndexerConfig{Type: "opensearch",
				Servers:            []string{},
				Index:              "go-commons",
				InsecureSkipVerify: true,
			},
			"",
			true,
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(payload)
			})),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name != "Test 3" {
				defer tt.mockServer.Close()
				tt.indexerConfig.Servers = []string{tt.mockServer.URL}
			}

			if tt.name == "Test 4" {
				tt.indexerConfig.Servers = []string{}
				os.Setenv("ELASTICSEARCH_URL", "not a valid url:port")
				defer os.Unsetenv("ELASTICSEARCH_URL")

			}
			err := indexer.new(tt.indexerConfig)

			if (err != nil) == tt.wantErr {
				return
			}
			expectedError := tt.expectedError
			t.Errorf("Expected err: %v, Actual Error: %v", expectedError, err)

		})
	}

}

func TestOpenSearchIndex(t *testing.T) {
	var indexer OpenSearch
	indexer.index = "abc"
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
				JobName:    "placeholder",
			},
			expectedError: "",
			wantErr:       false,
		},

		{name: "test2",
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

		{name: "test3",
			documents: []interface{}{
				"example document",
				42,
			},
			opts: IndexingOpts{
				MetricName: "placeholder",
				JobName:    "placeholder",
			},
			expectedError: "",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexerConfig := IndexerConfig{
				Type:               "opensearch",
				Servers:            []string{"https://search-perfscale-dev-chmf5l4sh66lvxbnadi4bznl3a.us-west-2.es.amazonaws.com"},
				Index:              "go-commons-test",
				InsecureSkipVerify: true,
			}

			if tt.name == "test3" {
				indexer.new(indexerConfig)
			}

			_, err := indexer.Index(tt.documents, tt.opts)

			if (err != nil) == tt.wantErr {
				return
			}
			expectedError := tt.expectedError
			t.Errorf("Expected err: %v, Actual Error: %v", expectedError, err)

		})
	}

}
