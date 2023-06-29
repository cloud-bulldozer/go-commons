package comparison

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"

	//"log"

	"net/http"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var searchFunc = func(o ...func(*esapi.SearchRequest)) (*esapi.Response, error) {
	var res *esapi.Response
	res = &esapi.Response{}
	res.StatusCode = 200
	body := ioutil.NopCloser(bytes.NewReader([]byte(`{"took":170,"timed_out":false,"_shards":{"total":443,"successful":443,"skipped":0,"failed":0},"hits":{"total":{"value":0,"relation":"eq"},"max_score":null,"hits":[]},"aggregations":{"stats":{"count":0,"min":null,"max":null,"avg":null,"sum":0.0}}}`)))
	res.Body = body
	defer res.Body.Close()
	return res, nil
}

var _ = Describe("Tests for elastic.go", func() {
	Context("Tests for NewComparator", func() {
		var index string
		var client elasticsearch.Client
		var expectedComparator Comparator
		BeforeEach(func() {
			index = "go-commons-test"
			expectedComparator = Comparator{client: client, index: index}
		})
		It("Test 1: Returns the expected comparator", func() {
			actualComparator := NewComparator(client, index)
			Expect(actualComparator).To(BeEquivalentTo(expectedComparator))
		})
	})

	Context("Tests for queryStringStats()", func() {
		var query string
		var field string
		var client *elasticsearch.Client
		var comparator Comparator
		var cfg elasticsearch.Config
		BeforeEach(func() {
			client, _ = elasticsearch.NewClient(cfg)
			query = "_all"
			field = ""
			cfg = elasticsearch.Config{
				Addresses: []string{
					"https://justAnotherRandomLinkForNoReason.com",
				},
				Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
			}
		})
		It("Test1 error 404", func() {
			client.Search = func(o ...func(*esapi.SearchRequest)) (*esapi.Response, error) {
				var res *esapi.Response
				res = &esapi.Response{}
				res.StatusCode = 404
				body := ioutil.NopCloser(bytes.NewReader([]byte(`{"error":{"root_cause":[{"type":"index_not_found_exception","reason":"no such index [placeholder]","index":"placeholder","resource.id":"placeholder","resource.type":"index_or_alias","index_uuid":"_na_"}],"type":"index_not_found_exception","reason":"no such index [placeholder]","index":"placeholder","resource.id":"placeholder","resource.type":"index_or_alias","index_uuid":"_na_"},"status":404}`)))
				res.Body = body
				defer res.Body.Close()
				return res, nil
			}
			comparator = NewComparator(*client, "placeholder")
			_, err := comparator.queryStringStats(query, field)
			Expect(err).To(BeEquivalentTo(errors.New("404 Not Found index_not_found_exception no such index [placeholder]")))
		})

		It("Test2 no error", func() {
			client.Search = searchFunc
			comparator = NewComparator(*client, "_all")
			_, err := comparator.queryStringStats(query, field)
			Expect(err).To(BeNil())
		})

		It("Test3 not a valid link", func() {
			comparator = NewComparator(*client, "_all")
			_, err := comparator.queryStringStats(query, field)
			Expect(err.Error()).To(ContainSubstring("dial tcp: lookup justAnotherRandomLinkForNoReason.com"))
		})

		It("Test4 non parsable json", func() {
			client.Search = func(o ...func(*esapi.SearchRequest)) (*esapi.Response, error) {
				var res *esapi.Response
				res = &esapi.Response{}
				res.StatusCode = 200
				body := ioutil.NopCloser(bytes.NewReader([]byte(`hi this a non parsable`)))
				res.Body = body
				defer res.Body.Close()
				return res, nil
			}
			comparator = NewComparator(*client, "_all")
			_, err := comparator.queryStringStats(query, field)
			Expect(err).To(BeEquivalentTo(errors.New("error parsing the response body: invalid character 'h' looking for beginning of value")))
		})

		It("Test5 non parsable json", func() {
			client.Search = func(o ...func(*esapi.SearchRequest)) (*esapi.Response, error) {
				var res *esapi.Response
				res = &esapi.Response{}
				res.StatusCode = 404
				body := ioutil.NopCloser(bytes.NewReader([]byte(`hi non parse`)))
				res.Body = body
				defer res.Body.Close()
				return res, nil
			}
			comparator = NewComparator(*client, "placeholder")
			_, err := comparator.queryStringStats(query, field)
			Expect(err).To(BeEquivalentTo(errors.New("error parsing the response body: invalid character 'h' looking for beginning of value")))
		})
	})

	Context("Tests on Compare()", func() {
		var client *elasticsearch.Client
		cfg := elasticsearch.Config{
			Addresses: []string{
				"https://justAnotherRandomLinkForNoReason.com",
			},
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}
		BeforeEach(func() {
			client, _ = elasticsearch.NewClient(cfg)
		})

		It("Test1 no error", func() {
			client.Search = searchFunc
			comparator := NewComparator(*client, "_all")
			_, err := comparator.Compare("placeholder", "placeholder", "max", 1.0, 1.0)
			Expect(err).To(BeNil())
		})

		It("Test2 negative value", func() {
			client.Search = searchFunc
			comparator := NewComparator(*client, "_all")
			_, err := comparator.Compare("placeholder", "placeholder", "min", -1.0, 1.0)
			log.Printf("%v", err)
			Expect(err).NotTo(BeNil())
		})

		It("Test3 negative tolerance", func() {
			client.Search = searchFunc
			comparator := NewComparator(*client, "_all")
			_, err := comparator.Compare("placeholder", "placeholder", "avg", 1.0, -1.0)
			log.Printf("%v", err)
			Expect(err).NotTo(BeNil())
		})

		It("Test4 negative tolerance", func() {
			client.Search = searchFunc
			comparator := NewComparator(*client, "_all")
			_, err := comparator.Compare("placeholder", "placeholder", "sum", 1.0, -1.0)
			Expect(err).NotTo(BeNil())
		})

		It("Test5 no error with sum stat", func() {
			comparator := NewComparator(*client, "_all")
			_, err := comparator.Compare("placeholder", "placeholder", "sum", 1.0, -1.0)
			Expect(err).NotTo(BeNil())
		})

	})

})
