package indexers

import (
	"errors"
	"log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// testing local.go
var _ = Describe("Tests for local.go", func() {
	Context("Default behavior of local.go, new()", func() {
		testcase := struct {
			indexerconfig IndexerConfig
		}{
			indexerconfig: IndexerConfig{Type: "elastic",
				Servers:            []string{""},
				Index:              "go-commons-test",
				InsecureSkipVerify: true,
				MetricsDirectory:   "",
			},
		}
		var localIndexer Local
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
				JobName:    "",
			},
		}
		var indexer Local

		It("No err is returned", func() {
			indexer.metricsDirectory = "placeholder"
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err).To(BeNil())
		})

		It("No err is returned", func() {
			indexer.metricsDirectory = "placeholder"
			testcase.opts.JobName = "placeholder"
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err).To(BeNil())
		})

		It("Err is returned metricsdirectory has fault", func() {
			indexer.metricsDirectory = "abc"
			testcase.opts.JobName = "placeholder"
			_, err := indexer.Index(testcase.documents, testcase.opts)

			Expect(err).To(BeEquivalentTo(errors.New("Error creating metrics file abc/placeholder-placeholder.json: open abc/placeholder-placeholder.json: no such file or directory")))
		})

		It("Err is returned by documents not processed", func() {
			testcase.documents = append(testcase.documents, make(chan string))
			indexer.metricsDirectory = "placeholder"
			_, err := indexer.Index(testcase.documents, testcase.opts)
			log.Println(err)
			Expect(err).To(BeEquivalentTo(errors.New("JSON encoding error: json: unsupported type: chan string")))
		})
	})
})
