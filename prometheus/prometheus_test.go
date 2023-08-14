package prometheus

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tests for Prometheus", func() {

	Context("Tests for RoundTrip()", func() {
		var bat authTransport
		BeforeEach(func() {
			bat.username = "someRandomUsername"
			bat.password = "someRandomPassword"
			bat.token = "someRandomToken"
			bat.Transport = http.DefaultTransport
			count = 0
		})

		It("Test1 for default behaviour", func() {
			url := "https://example.com/api"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				fmt.Println("Failed to create request:", err)
				return
			}
			_, err = bat.RoundTrip(req)
			//Asserting no of times mocks are called
			Expect(count).To(BeEquivalentTo(0))
			Expect(err).To(BeNil())
		})
	})

	Context("Tests for NewClient()", func() {
		var url, username, password, token string
		var tlsSkipVerify bool
		BeforeEach(func() {
			url = ""
			username = ""
			password = ""
			token = ""
			tlsSkipVerify = false
			count = 0
		})
		It("Test1 empty parameters", func() {
			_, err := NewClient(url, token, username, password, tlsSkipVerify)
			//Asserting no of times mocks are called
			Expect(count).To(BeEquivalentTo(0))
			Expect(err.Error()).To(ContainSubstring("Get \"/api/v1/status/runtimeinfo\": unsupported protocol scheme \"\""))
		})

		It("Test2 passing not valid url", func() {
			url = "not a valid url:port"
			//Asserting no of times mocks are called
			Expect(count).To(BeEquivalentTo(0))
			_, err := NewClient(url, token, username, password, tlsSkipVerify)
			Expect(err.Error()).To(ContainSubstring("parse \"not a valid url:port\": first path segment in URL cannot contain colon"))
		})

	})

	Context("Tests for Query()", func() {
		var url, username, password, token string
		var tlsSkipVerify bool
		var pr *Prometheus
		BeforeEach(func() {
			pr, _ = NewClient(url, token, username, password, tlsSkipVerify)
			count = 0
		})

		It("Test1 empty url", func() {
			_, err := pr.Query("_all", time.Now())
			//Asserting no of times mocks are called
			Expect(count).To(BeEquivalentTo(0))
			Expect(err.Error()).To(ContainSubstring("Post \"/api/v1/query\": unsupported protocol scheme \"\""))
		})

		It("Test2 mock error to nil", func() {
			mockAPI := new(MockAPI)
			query := "your_query"
			start := time.Now()
			p := Prometheus{api: mockAPI}
			_, err := p.Query(query, start)
			//Asserting no of times mocks are called
			Expect(count).To(BeEquivalentTo(1))
			Expect(err).To(BeNil())
		})
	})

	Context("Tests for QueryRange()", func() {
		var url, username, password, token string
		var tlsSkipVerify bool
		var pr *Prometheus
		BeforeEach(func() {
			pr, _ = NewClient(url, token, username, password, tlsSkipVerify)
			count = 0
		})

		It("Test1 empty url", func() {
			_, err := pr.QueryRange("_all", time.Now(), time.Now().Add(time.Duration(10)), time.Duration(5))
			//Asserting no of times mocks are called
			Expect(count).To(BeEquivalentTo(0))
			Expect(err.Error()).To(ContainSubstring("Post \"/api/v1/query_range\": unsupported protocol scheme \"\""))
		})

	})

	Context("Tests for verifyConnection()", func() {
		var mockAPI *MockAPI
		var p Prometheus
		BeforeEach(func() {
			mockAPI = new(MockAPI)
			p = Prometheus{api: mockAPI}
			count = 0
		})
		It("Test1 mock to no nil", func() {
			err := p.verifyConnection()
			//Asserting no of times mocks are called
			Expect(count).To(BeEquivalentTo(1))
			Expect(err).To(BeNil())

		})
	})
})
