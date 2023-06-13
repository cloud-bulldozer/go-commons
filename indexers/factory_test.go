package indexers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewIndexer(t *testing.T) {
	tests := []struct {
		name          string
		indexerConfig IndexerConfig
		wantIndexer   Indexer
		wantErr       bool
		mockServer    *httptest.Server
	}{
		//testcase1 runs as intended without any error
		{"Test 1",
			IndexerConfig{Type: "elastic",
				Servers:            []string{""},
				Index:              "go-commons-test",
				InsecureSkipVerify: true,
			},
			&Elastic{"go-commons-test"},
			false,
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
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
				  }`))
			})),
		},

		//test 2 creates error in factory at line 34
		{"Test 2",
			IndexerConfig{Type: "Unknown",
				Servers:            []string{""},
				Index:              "go-commons-test",
				InsecureSkipVerify: true,
			},
			&Elastic{"go-commons-test"},
			true,
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
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
				  }`))
			})), //mockServerEnds
		}, //testcase ends

		//test3 creates error in factory at
		{"Test 3",
			IndexerConfig{Type: "elastic",
				Servers:            []string{"placeholderserver"},
				Index:              "go-commons-test",
				InsecureSkipVerify: true,
			},
			&Elastic{"go-commons-test"},
			true,
			httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte(`{
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
				  }`))
			})),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer tt.mockServer.Close()
			tt.indexerConfig.Servers = []string{tt.mockServer.URL}
			got, err := NewIndexer(tt.indexerConfig)
			fmt.Printf("ho %v, %v\n", *got, err)
			if (err != nil) == tt.wantErr {
				return
			}

			want := tt.wantIndexer
			if !reflect.DeepEqual(*got, want) {
				t.Errorf("NewIndexer() error: got %v, want %v\n", got, want)
				return
			}
		})
	}
}
