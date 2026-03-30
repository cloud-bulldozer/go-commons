// Copyright 2024 The go-commons Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package indexers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	gokitlog "github.com/go-kit/log"
	log "github.com/sirupsen/logrus"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/tsdb"
)

// TSDB indexer creates native Prometheus TSDB blocks from indexed documents.
// Each Index() call writes a complete TSDB block to the metrics directory.
// It handles two document formats:
//   - Prometheus-style metrics: documents with "value" (number) and optional "labels" (map)
//   - Runtime measurements: documents with multiple numeric fields (e.g. latencies),
//     each numeric field becomes a separate time series with a "field" label.
type TSDB struct {
	metricsDirectory string
}

type tsdbSample struct {
	labels    labels.Labels
	timestamp int64
	value     float64
}

// Fields to skip when decomposing measurement documents into samples.
var measurementSkipKeys = map[string]bool{
	"timestamp":  true,
	"metricName": true,
	"metadata":   true,
}

// NewTSDBIndexer returns a new TSDB indexer that writes Prometheus TSDB blocks
func NewTSDBIndexer(indexerConfig IndexerConfig) (*TSDB, error) {
	if indexerConfig.MetricsDirectory == "" {
		return nil, fmt.Errorf("metricsDirectory not specified for TSDB indexer")
	}
	if err := os.MkdirAll(indexerConfig.MetricsDirectory, 0744); err != nil {
		return nil, fmt.Errorf("error creating metrics directory: %v", err)
	}
	return &TSDB{
		metricsDirectory: indexerConfig.MetricsDirectory,
	}, nil
}

// Index converts documents to TSDB samples and writes them as a TSDB block.
func (t *TSDB) Index(documents []interface{}, opts IndexingOpts) (string, error) {
	if len(documents) == 0 {
		return "", fmt.Errorf("empty document list in %s", opts.MetricName)
	}

	var samples []tsdbSample
	for _, doc := range documents {
		jsonBytes, err := json.Marshal(doc)
		if err != nil {
			log.Warnf("TSDB indexer: error marshalling document: %v", err)
			continue
		}
		var docMap map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &docMap); err != nil {
			log.Warnf("TSDB indexer: error unmarshalling document: %v", err)
			continue
		}
		samples = append(samples, extractSamples(docMap, opts)...)
	}

	if len(samples) == 0 {
		return "", fmt.Errorf("TSDB indexer: no valid samples for %s", opts.MetricName)
	}

	if err := t.writeBlock(samples); err != nil {
		return "", err
	}

	return fmt.Sprintf("TSDB indexer: wrote block with %d samples for %s", len(samples), opts.MetricName), nil
}

// extractSamples converts a document map into one or more TSDB samples.
// Prometheus-style docs (with "value" and optional "labels" map) produce a single sample.
// Measurement-style docs produce one sample per numeric field.
func extractSamples(doc map[string]interface{}, opts IndexingOpts) []tsdbSample {
	ts := extractTimestamp(doc)
	if ts == 0 {
		return nil
	}

	metricName := opts.MetricName
	if metricName == "" {
		if mn, ok := doc["metricName"].(string); ok {
			metricName = mn
		}
	}
	if metricName == "" {
		return nil
	}

	// Prometheus-style metric: has "value" number (with optional "labels" map)
	if valueRaw, hasValue := doc["value"]; hasValue {
		if floatValue, ok := valueRaw.(float64); ok {
			var labelsMap map[string]interface{}
			if labelsRaw, hasLabels := doc["labels"]; hasLabels {
				if lm, ok := labelsRaw.(map[string]interface{}); ok {
					labelsMap = lm
				}
			}
			if labelsMap == nil {
				labelsMap = make(map[string]interface{})
			}
			return promStyleSample(metricName, labelsMap, floatValue, ts, doc)
		}
	}

	// Runtime measurement style: decompose numeric fields into separate samples
	return measurementSamples(metricName, ts, doc)
}

// promStyleSample creates a single TSDB sample from a prometheus-style document.
func promStyleSample(metricName string, labelsMap map[string]interface{}, value float64, ts int64, doc map[string]interface{}) []tsdbSample {
	b := labels.NewBuilder(labels.EmptyLabels())
	b.Set(labels.MetricName, metricName)
	for k, v := range labelsMap {
		if sv, ok := v.(string); ok {
			b.Set(k, sv)
		}
	}
	if uuid, ok := doc["uuid"].(string); ok && uuid != "" {
		b.Set("uuid", uuid)
	}
	if jobName, ok := doc["jobName"].(string); ok && jobName != "" {
		b.Set("job_name", jobName)
	}
	return []tsdbSample{{
		labels:    b.Labels(),
		timestamp: ts,
		value:     value,
	}}
}

// measurementSamples decomposes a measurement document into multiple TSDB samples.
// Each numeric field becomes a separate sample with a "field" label identifying it.
// All string fields become labels on every sample.
func measurementSamples(metricName string, ts int64, doc map[string]interface{}) []tsdbSample {
	stringFields := map[string]string{}
	numericFields := map[string]float64{}

	for k, v := range doc {
		if measurementSkipKeys[k] {
			continue
		}
		switch val := v.(type) {
		case string:
			if val != "" {
				stringFields[k] = val
			}
		case float64:
			numericFields[k] = val
		}
	}

	if len(numericFields) == 0 {
		return nil
	}

	var samples []tsdbSample
	for field, value := range numericFields {
		b := labels.NewBuilder(labels.EmptyLabels())
		b.Set(labels.MetricName, metricName)
		b.Set("field", field)
		for k, v := range stringFields {
			b.Set(k, v)
		}
		samples = append(samples, tsdbSample{
			labels:    b.Labels(),
			timestamp: ts,
			value:     value,
		})
	}
	return samples
}

// writeBlock writes all samples as a single TSDB block to the metrics directory.
func (t *TSDB) writeBlock(samples []tsdbSample) error {
	minTime := samples[0].timestamp
	maxTime := samples[0].timestamp
	for _, s := range samples[1:] {
		if s.timestamp < minTime {
			minTime = s.timestamp
		}
		if s.timestamp > maxTime {
			maxTime = s.timestamp
		}
	}

	blockDuration := maxTime - minTime + 1
	if blockDuration < tsdb.DefaultBlockDuration {
		blockDuration = tsdb.DefaultBlockDuration
	}

	w, err := tsdb.NewBlockWriter(gokitlog.NewNopLogger(), t.metricsDirectory, blockDuration)
	if err != nil {
		return fmt.Errorf("error creating TSDB block writer: %v", err)
	}
	defer func() {
		if closeErr := w.Close(); closeErr != nil {
			log.Infof("TSDB indexer: error closing block writer: %v", closeErr)
		}
	}()

	app := w.Appender(context.Background())
	for _, s := range samples {
		if _, err := app.Append(0, s.labels, s.timestamp, s.value); err != nil {
			log.Infof("TSDB indexer: error appending sample: %v", err)
		}
	}

	if err := app.Commit(); err != nil {
		return fmt.Errorf("error committing TSDB samples: %v", err)
	}

	blockID, err := w.Flush(context.Background())
	if err != nil {
		return fmt.Errorf("error flushing TSDB block: %v", err)
	}

	log.Infof("TSDB indexer: created block %s with %d samples", blockID, len(samples))
	return nil
}

// extractTimestamp parses the "timestamp" field from a document map.
func extractTimestamp(doc map[string]interface{}) int64 {
	tsRaw, ok := doc["timestamp"]
	if ok {
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
			if t, err := time.Parse(layout, tsRaw.(string)); err == nil {
				return t.UnixMilli()
			}
		}
	}
	return 0
}
