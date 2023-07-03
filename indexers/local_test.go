package indexers

import (
	"errors"
	"log"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// testing local.go
var _ = Describe("Tests for local.go", func() {
	Context("Default behavior of local.go, new()", func() {
		type newtestcase struct {
			indexerconfig IndexerConfig
		}
		var testcase newtestcase
		var localIndexer Local
		BeforeEach(func() {
			testcase = newtestcase{
				indexerconfig: IndexerConfig{Type: "elastic",
					Servers:            []string{""},
					Index:              "go-commons-test",
					InsecureSkipVerify: true,
					MetricsDirectory:   "",
				},
			}
		})

		It("returns err no metrics directory", func() {
			err := localIndexer.new(testcase.indexerconfig)
			Expect(err).To(BeEquivalentTo(errors.New("directory name not specified")))
		})

		It("returns nil as error", func() {
			testcase.indexerconfig.MetricsDirectory = "placeholder"
			err := localIndexer.new(testcase.indexerconfig)
			Expect(err).To(BeNil())
		})
	})

	Context("Default behaviour of local.go, Index()", func() {
		var testcase indexMethodTestcase
		var indexer Local
		indexer.metricsDirectory = "placeholder"
		BeforeEach(func() {
			err := os.MkdirAll(indexer.metricsDirectory, 0744)
			if err != nil {
				log.Fatal(err)
			}
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

		It("No err is returned", func() {
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err).To(BeNil())
		})

		It("No err is returned", func() {
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err).To(BeNil())
		})

		It("Err is returned metricsdirectory has fault", func() {
			indexer.metricsDirectory = "abc"
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err).To(BeEquivalentTo(errors.New("Error creating metrics file abc/placeholder.json: open abc/placeholder.json: no such file or directory")))
		})

		It("Err is returned by documents not processed", func() {
			testcase.documents = append(testcase.documents, make(chan string))
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err).To(BeEquivalentTo(errors.New("JSON encoding error: json: unsupported type: chan string")))
		})
	})
})
