package indexers

import (
	. "github.com/onsi/ginkgo"
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
		It("returns err no metrics directory", func() {
			var localIndexer Local
			err := localIndexer.new(testcase.indexerconfig)
			Expect(err).NotTo(BeNil())
		})

		It("returns nil as error", func() {
			var localIndexer Local
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

		It("No err is returned", func() {
			var indexer Local
			indexer.metricsDirectory = "placeholder"
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err).To(BeNil())
		})

		It("No err is returned", func() {
			var indexer Local
			indexer.metricsDirectory = "placeholder"
			testcase.opts.JobName = "placeholder"
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err).To(BeNil())
		})

		It("Err is returned metricsdirectory has fault", func() {
			var indexer Local
			indexer.metricsDirectory = "abc"
			testcase.opts.JobName = "placeholder"
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err).NotTo(BeNil())
		})

		It("Err is returned by documents not processed", func() {
			testcase.documents = append(testcase.documents, make(chan string))
			var indexer Local
			indexer.metricsDirectory = "placeholder"
			_, err := indexer.Index(testcase.documents, testcase.opts)
			Expect(err).NotTo(BeNil())
		})
	})
})
