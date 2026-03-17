package indexers

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tests for tsdb.go", func() {
	Context("NewTSDBIndexer()", func() {
		It("returns error when metricsDirectory is empty", func() {
			_, err := NewTSDBIndexer(IndexerConfig{Type: TSDBIndexer})
			Expect(err).To(MatchError("metricsDirectory not specified for TSDB indexer"))
		})

		It("creates indexer and directory successfully", func() {
			dir := filepath.Join(os.TempDir(), "tsdb-test-new")
			defer os.RemoveAll(dir)
			indexer, err := NewTSDBIndexer(IndexerConfig{
				Type:             TSDBIndexer,
				MetricsDirectory: dir,
			})
			Expect(err).To(BeNil())
			Expect(indexer).NotTo(BeNil())
			_, err = os.Stat(dir)
			Expect(err).To(BeNil())
		})
	})

	Context("Index() with prometheus-style metrics", func() {
		var indexer *TSDB
		var dir string

		BeforeEach(func() {
			dir = filepath.Join(os.TempDir(), "tsdb-test-prom")
			var err error
			indexer, err = NewTSDBIndexer(IndexerConfig{
				Type:             TSDBIndexer,
				MetricsDirectory: dir,
			})
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			os.RemoveAll(dir)
		})

		It("returns error on empty document list", func() {
			_, err := indexer.Index([]interface{}{}, IndexingOpts{MetricName: "test"})
			Expect(err).To(MatchError("empty document list in test"))
		})

		It("writes a TSDB block for prometheus-style documents", func() {
			now := time.Now().UTC()
			docs := []interface{}{
				map[string]interface{}{
					"timestamp":  now.Format(time.RFC3339Nano),
					"labels":     map[string]interface{}{"instance": "node1"},
					"value":      42.5,
					"uuid":       "test-uuid",
					"metricName": "cpuUsage",
					"jobName":    "test-job",
				},
				map[string]interface{}{
					"timestamp":  now.Add(30 * time.Second).Format(time.RFC3339Nano),
					"labels":     map[string]interface{}{"instance": "node1"},
					"value":      43.0,
					"uuid":       "test-uuid",
					"metricName": "cpuUsage",
					"jobName":    "test-job",
				},
			}
			resp, err := indexer.Index(docs, IndexingOpts{MetricName: "cpuUsage"})
			Expect(err).To(BeNil())
			Expect(resp).To(ContainSubstring("wrote block with 2 samples"))

			// Verify a TSDB block directory was created
			verifyBlockExists(dir)
		})

		It("skips documents with zero timestamp", func() {
			docs := []interface{}{
				map[string]interface{}{
					"value":      1.0,
					"metricName": "test",
					"labels":     map[string]interface{}{},
				},
			}
			resp, err := indexer.Index(docs, IndexingOpts{MetricName: "test"})
			Expect(err).To(BeNil())
			Expect(resp).To(ContainSubstring("no valid samples"))
		})
	})

	Context("Index() with runtime measurement documents", func() {
		var indexer *TSDB
		var dir string

		BeforeEach(func() {
			dir = filepath.Join(os.TempDir(), "tsdb-test-measurement")
			var err error
			indexer, err = NewTSDBIndexer(IndexerConfig{
				Type:             TSDBIndexer,
				MetricsDirectory: dir,
			})
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			os.RemoveAll(dir)
		})

		It("decomposes measurement documents into per-field samples", func() {
			now := time.Now().UTC()
			docs := []interface{}{
				map[string]interface{}{
					"timestamp":              now.Format(time.RFC3339Nano),
					"schedulingLatency":      100.0,
					"initializedLatency":     200.0,
					"containersReadyLatency": 300.0,
					"podReadyLatency":        400.0,
					"metricName":             "podLatencyMeasurement",
					"namespace":              "default",
					"podName":                "test-pod-1",
					"nodeName":               "node1",
					"uuid":                   "test-uuid",
					"jobName":                "create-pods",
				},
			}
			resp, err := indexer.Index(docs, IndexingOpts{MetricName: "podLatencyMeasurement"})
			Expect(err).To(BeNil())
			// 4 latency fields + jobIteration-like fields if any
			Expect(resp).To(ContainSubstring("wrote block with"))
			verifyBlockExists(dir)
		})

		It("handles latency quantiles documents", func() {
			now := time.Now().UTC()
			docs := []interface{}{
				map[string]interface{}{
					"timestamp":    now.Format(time.RFC3339Nano),
					"quantileName": "Ready",
					"P99":          500.0,
					"P95":          400.0,
					"P50":          200.0,
					"min":          50.0,
					"max":          600.0,
					"avg":          250.0,
					"metricName":   "podLatencyQuantilesMeasurement",
					"uuid":         "test-uuid",
					"jobName":      "create-pods",
				},
			}
			resp, err := indexer.Index(docs, IndexingOpts{MetricName: "podLatencyQuantilesMeasurement"})
			Expect(err).To(BeNil())
			// 6 numeric fields: P99, P95, P50, min, max, avg
			Expect(resp).To(ContainSubstring("wrote block with 6 samples"))
			verifyBlockExists(dir)
		})

		It("skips documents with no numeric fields", func() {
			now := time.Now().UTC()
			docs := []interface{}{
				map[string]interface{}{
					"timestamp":  now.Format(time.RFC3339Nano),
					"metricName": "test",
					"name":       "something",
				},
			}
			resp, err := indexer.Index(docs, IndexingOpts{MetricName: "test"})
			Expect(err).To(BeNil())
			Expect(resp).To(ContainSubstring("no valid samples"))
		})
	})

	Context("extractSamples()", func() {
		It("uses opts.MetricName as __name__ label for prom-style docs", func() {
			now := time.Now().UTC()
			doc := map[string]interface{}{
				"timestamp": now.Format(time.RFC3339Nano),
				"value":     1.0,
				"labels":    map[string]interface{}{},
			}
			samples := extractSamples(doc, IndexingOpts{MetricName: "myMetric"})
			Expect(samples).To(HaveLen(1))
			Expect(samples[0].labels.Get("__name__")).To(Equal("myMetric"))
		})

		It("sets field label for measurement-style docs", func() {
			now := time.Now().UTC()
			doc := map[string]interface{}{
				"timestamp":         now.Format(time.RFC3339Nano),
				"schedulingLatency": 100.0,
				"namespace":         "default",
			}
			samples := extractSamples(doc, IndexingOpts{MetricName: "podLatency"})
			Expect(samples).To(HaveLen(1))
			Expect(samples[0].labels.Get("__name__")).To(Equal("podLatency"))
			Expect(samples[0].labels.Get("field")).To(Equal("schedulingLatency"))
			Expect(samples[0].labels.Get("namespace")).To(Equal("default"))
			Expect(samples[0].value).To(Equal(100.0))
		})

		It("carries string fields as labels on measurement samples", func() {
			now := time.Now().UTC()
			doc := map[string]interface{}{
				"timestamp":    now.Format(time.RFC3339Nano),
				"P99":          500.0,
				"P50":          200.0,
				"quantileName": "Ready",
				"uuid":         "abc",
				"jobName":      "test",
			}
			samples := extractSamples(doc, IndexingOpts{MetricName: "quantiles"})
			Expect(samples).To(HaveLen(2))
			for _, s := range samples {
				Expect(s.labels.Get("quantileName")).To(Equal("Ready"))
				Expect(s.labels.Get("uuid")).To(Equal("abc"))
				Expect(s.labels.Get("jobName")).To(Equal("test"))
			}
		})
	})

	Context("Factory integration", func() {
		It("NewIndexer creates TSDB indexer", func() {
			dir := filepath.Join(os.TempDir(), "tsdb-test-factory")
			defer os.RemoveAll(dir)
			indexer, err := NewIndexer(IndexerConfig{
				Type:             TSDBIndexer,
				MetricsDirectory: dir,
			})
			Expect(err).To(BeNil())
			Expect(indexer).NotTo(BeNil())
		})

		It("returns error for unknown indexer", func() {
			_, err := NewIndexer(IndexerConfig{Type: "unknown"})
			Expect(err).To(BeEquivalentTo(errors.New("Indexer not found: unknown")))
		})
	})
})

func verifyBlockExists(dir string) {
	entries, err := os.ReadDir(dir)
	Expect(err).To(BeNil())
	blockFound := false
	for _, entry := range entries {
		if entry.IsDir() {
			metaPath := filepath.Join(dir, entry.Name(), "meta.json")
			if _, err := os.Stat(metaPath); err == nil {
				blockFound = true
				content, err := os.ReadFile(metaPath)
				Expect(err).To(BeNil())
				var meta map[string]interface{}
				Expect(json.Unmarshal(content, &meta)).To(BeNil())
			}
		}
	}
	Expect(blockFound).To(BeTrue())
}
