package indexers

import (
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
)

var payload []byte
var _ = BeforeSuite(func() {
	payload = []byte(`{
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

})

type newMethodTestcase struct {
	indexerConfig IndexerConfig
	mockServer    *httptest.Server
}

type indexMethodTestcase struct {
	documents []interface{}
	opts      IndexingOpts
}
