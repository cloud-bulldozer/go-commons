package indexers

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Testing factory.go
var _ = Describe("Factory.go Unit Tests: NewIndexer()", func() {
	var testcase newMethodTestcase
	BeforeEach(func() {
		testcase = newMethodTestcase{
			indexerConfig: IndexerConfig{Type: "elastic",
				Servers:            []string{""},
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
	})

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

			Expect(err).To(BeEquivalentTo(errors.New("unexpected ES status code: 502")))
		})

		It("returns indexer and err unknown indexer", func() {
			defer testcase.mockServer.Close()
			testcase.indexerConfig.Servers = []string{testcase.mockServer.URL}
			testcase.indexerConfig.Type = "Unknown"
			_, err := NewIndexer(testcase.indexerConfig)
			Expect(err).To(BeEquivalentTo(errors.New("Indexer not found: Unknown")))
		})

	})
})

// Unit Test to call opensearch.new()
var _ = Describe("Factory.go Unit Tests: NewIndexer()", func() {
	var testcase newMethodTestcase
	BeforeEach(func() {
		testcase = newMethodTestcase{
			indexerConfig: IndexerConfig{Type: "opensearch",
				Servers:            []string{""},
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
	})

	Context("Default behaviour of NewIndexer()", func() {
		It("returns indexer and nil", func() {
			defer testcase.mockServer.Close()
			testcase.indexerConfig.Servers = []string{testcase.mockServer.URL}
			_, err := NewIndexer(testcase.indexerConfig)
			Expect(err).To(BeNil())
		})

	})
})
