package indexers

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tests for elastic.go", func() {
	Context("Tests for new()", func() {
		var testcase newMethodTestcase
		var indexer Elastic
		BeforeEach(func() {
			testcase = newMethodTestcase{
				indexerConfig: IndexerConfig{Type: "elastic",
					Servers:            []string{},
					Index:              "go-commons-test",
					InsecureSkipVerify: true,
				},
				mockServer: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, err := w.Write(payload)
					if err != nil {
						log.Printf("Error while sending payload to http mock server: %v", err)
					}
				})),
			}

			indexer.index = "go-commons-test"
		})

		It("Returns error status bad request", func() {
			testcase.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			}))
			defer testcase.mockServer.Close()
			testcase.indexerConfig.Servers = []string{testcase.mockServer.URL}
			_, err := NewElasticIndexer(testcase.indexerConfig)
			Expect(err).To(BeEquivalentTo(errors.New("unexpected ES status code: 400")))
		})

		It("when no url is passed", func() {
			_, err := NewElasticIndexer(testcase.indexerConfig)
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
			_, err := NewElasticIndexer(testcase.indexerConfig)
			Expect(err).To(BeEquivalentTo(errors.New("error creating the ES client: cannot create client: cannot parse url: parse \"not a valid url:port\": first path segment in URL cannot contain colon")))
		})

		It("Returns err no index name", func() {
			defer testcase.mockServer.Close()
			testcase.indexerConfig.Servers = []string{testcase.mockServer.URL}
			testcase.indexerConfig.Index = ""
			_, err := NewElasticIndexer(testcase.indexerConfig)
			Expect(err).To(BeEquivalentTo(errors.New("index name not specified")))
		})

	})

	Context("Tests for Index()", func() {
		var testcase indexMethodTestcase
		var indexer Elastic
		BeforeEach(func() {
			testcase = indexMethodTestcase{
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
				},
			}
		})

		It("No err returned", func() {
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err).To(BeNil())
		})

		It("Test empty list of docs", func() {
			_, err := indexer.Index([]interface{}{}, testcase.opts)
			Expect(err).To(BeNil())
		})

		It("Redundant list of docs", func() {
			lastDoc := testcase.documents[len(testcase.documents)-1]
			testcase.documents = append(testcase.documents, lastDoc)
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err).To(BeNil())
		})

		It("err returned docs not processed", func() {
			testcase.documents = append(testcase.documents, make(chan string))
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err.Error()).To(ContainSubstring("Cannot encode document"))
		})

	})
})
