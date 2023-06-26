package indexers

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Testing factory.go
var _ = Describe("Factory.go Unit Tests: NewIndexer()", func() {
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
	testcase := struct {
		indexerConfig IndexerConfig
		mockServer    *httptest.Server
	}{
		indexerConfig: IndexerConfig{Type: "elastic",
			Servers:            []string{""},
			Index:              "go-commons-test",
			InsecureSkipVerify: true,
		},
		mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(payload)
		})),
	}
	Context("Default behaviour of NewIndexer()", func() {
		It("returns indexer and nil", func() {
			defer testcase.mockServer.Close()
			testcase.indexerConfig.Servers = []string{testcase.mockServer.URL}
			_, err := NewIndexer(testcase.indexerConfig)
			Expect(err).To(BeNil())
		})

		It("returns indexer and err status bad gateway", func() {
			testcase.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			}))
			defer testcase.mockServer.Close()
			testcase.indexerConfig.Servers = []string{testcase.mockServer.URL}
			_, err := NewIndexer(testcase.indexerConfig)
			Expect(err).NotTo(BeNil())
		})

		It("returns indexer and err unknown indexer", func() {
			defer testcase.mockServer.Close()
			testcase.indexerConfig.Servers = []string{testcase.mockServer.URL}
			testcase.indexerConfig.Type = "Unknown"
			_, err := NewIndexer(testcase.indexerConfig)
			Expect(err).NotTo(BeNil())
		})

	})
})

// Unit Test to call opensearch.new()
var _ = Describe("Factory.go Unit Tests: NewIndexer()", func() {
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
	testcase := struct {
		indexerConfig IndexerConfig
		mockServer    *httptest.Server
	}{
		indexerConfig: IndexerConfig{Type: "opensearch",
			Servers:            []string{""},
			Index:              "go-commons-test",
			InsecureSkipVerify: true,
		},
		mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(payload)
		})),
	}
	Context("Default behaviour of NewIndexer()", func() {

		It("returns indexer and nil", func() {
			defer testcase.mockServer.Close()
			testcase.indexerConfig.Servers = []string{testcase.mockServer.URL}
			_, err := NewIndexer(testcase.indexerConfig)
			Expect(err).To(BeNil())
		})

	})
})
