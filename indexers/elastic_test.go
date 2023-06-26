package indexers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tests for elastic.go", func() {
	Context("Tests for new()", func() {
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
				Servers:            []string{},
				Index:              "go-commons-test",
				InsecureSkipVerify: true,
			},
			mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write(payload)
			})),
		}
		var indexer Elastic
		indexer.index = "go-commons-test"
		It("Returns error status bad request", func() {
			testcase.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			}))
			defer testcase.mockServer.Close()
			testcase.indexerConfig.Servers = []string{testcase.mockServer.URL}
			err := indexer.new(testcase.indexerConfig)
			Expect(err).To(BeEquivalentTo(errors.New("unexpected ES status code: 400")))
		})

		It("when no url is passed", func() {
			err := indexer.new(testcase.indexerConfig)
			testcase.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusGatewayTimeout)
			}))
			//using .Error() to convert to string as the error which is generated contains port and is dynamic
			Expect(err.Error()).To(ContainSubstring("connect: connection refused"))
		})

		It("Returns err not passing a valid URL in env variable", func() {
			testcase.indexerConfig.Servers = []string{}
			os.Setenv("ELASTICSEARCH_URL", "not a valid url:port")
			defer os.Unsetenv("ELASTICSEARCH_URL")
			defer testcase.mockServer.Close()
			err := indexer.new(testcase.indexerConfig)
			Expect(err).To(BeEquivalentTo(errors.New("error creating the ES client: cannot create client: cannot parse url: parse \"not a valid url:port\": first path segment in URL cannot contain colon")))
		})

		It("Returns err no index name", func() {
			defer testcase.mockServer.Close()
			testcase.indexerConfig.Servers = []string{testcase.mockServer.URL}
			testcase.indexerConfig.Index = ""
			err := indexer.new(testcase.indexerConfig)

			Expect(err).To(BeEquivalentTo(errors.New("index name not specified")))
		})

	})

	Context("Tests for Index()", func() {
		testcase := struct {
			documents []interface{}
			opts      IndexingOpts
		}{
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
		}
		var indexer Elastic

		It("No err returned", func() {
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err).To(BeNil())
		})

		// It("No err returned", func() {

		// 	testcase.documents = []interface{}{
		// 		map[string]interface{}{
		// 			"_index":   "dummy",
		// 			"_id":      "Vdc8g4IBYDIaJD7xG1yv",
		// 			"_version": 1,
		// 			"_score":   nil,
		// 			"_source": map[string]interface{}{
		// 				"timestamp":       "2022-08-09T15:32:10.784425",
		// 				"uuid":            "de4f28de-93de-4ab1-90dc-bdd67653b895",
		// 				"connection_time": 0.00347347604110837,
		// 			},
		// 			"fields": map[string]interface{}{
		// 				"timestamp": []string{
		// 					"2022-08-09T15:32:10.784Z"},
		// 			},
		// 			"sort": []int{
		// 				1660059130784,
		// 			},
		// 		},
		// 	}

		// 	_, err := indexer.Index(testcase.documents, testcase.opts)
		// 	Expect(err).To(BeNil())
		// })

		It("err returned docs not processed", func() {
			testcase.documents = append(testcase.documents, make(chan string))
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err.Error()).To(ContainSubstring("Cannot encode document"))
		})

	})
})
