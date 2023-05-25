// Copyright 2020 The Kube-burner Authors.
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

package prometheus

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"net/http"
	"text/template"
	"time"

	api "github.com/prometheus/client_golang/api"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Used to intercept and passed custom auth headers to prometheus client request
func (bat authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if bat.username != "" {
		req.SetBasicAuth(bat.username, bat.password)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bat.token))
	return bat.Transport.RoundTrip(req)
}

// NewPrometheusClient creates a prometheus struct instance with the given parameters
func NewPrometheusClient(url, token, username, password, uuid string, tlsVerify bool, step time.Duration, metadata map[string]interface{}) (*Prometheus, error) {
	p := Prometheus{
		Step:       step,
		UUID:       uuid,
		Endpoint:   url,
		metadata:   metadata,
	}

	cfg := api.Config{
		Address: url,
		RoundTripper: authTransport{
			Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsVerify}},
			token:     token,
			username:  username,
			password:  password,
		},
	}
	c, err := api.NewClient(cfg)
	if err != nil {
		return &p, err
	}
	p.api = apiv1.NewAPI(c)
	// Verify Prometheus connection prior returning
	if err := p.verifyConnection(); err != nil {
		return &p, err
	}
	return &p, nil
}

// ScrapeJobsMetrics gets all prometheus metrics required and handles them
func (p *Prometheus) ScrapeJobsMetrics() []QueryResult {
	start := p.JobList[0].Start
	end := p.JobList[len(p.JobList)-1].End
	elapsed := int(end.Sub(start).Minutes())
	var err error
	var v model.Value
	var renderedQuery bytes.Buffer
	vars := utils.EnvToMap()
	vars["elapsed"] = fmt.Sprintf("%dm", elapsed)
	var datapointsList []QueryResult
	for _, md := range p.MetricProfile {
		var datapoints []interface{}
		t, _ := template.New("").Parse(md.Query)
		t.Execute(&renderedQuery, vars)
		query := renderedQuery.String()
		renderedQuery.Reset()
		if md.Instant {
			if v, err = p.query(query, end); err != nil {
				datapointsList = append(datapointsList, QueryResult{
					DataPoints: nil,
					Err: fmt.Errorf("Error found with query %s: %s", query, err),
				})
				continue
			}
			if err := p.parseVector(md.MetricName, query, v, &datapoints); err != nil {
				datapointsList = append(datapointsList, QueryResult{
					DataPoints: nil,
					Err: fmt.Errorf("Error found parsing result from query %s: %s", query, err),
				})
				continue
			}
		} else {
			v, err = p.QueryRange(query, start, end)
			if err != nil {
				datapointsList = append(datapointsList, QueryResult{
					DataPoints: nil,
					Err: fmt.Errorf("Error found with query %s: %s", query, err),
				})
				continue
			}
			if err := p.parseMatrix(md.MetricName, query, v, &datapoints); err != nil {
				datapointsList = append(datapointsList, QueryResult{
					DataPoints: nil,
					Err: fmt.Errorf("Error found parsing result from query %s: %s", query, err),
				})
				continue
			}
		}
		datapointsList = append(datapointsList, QueryResult{
			DataPoints: datapoints,
			Err: nil,
		})
	}
	return datapointsList
}

// QueryRange prometheus queryRange wrapper
func (p *Prometheus) QueryRange(query string, start, end time.Time) (model.Value, error) {
	var v model.Value
	r := apiv1.Range{Start: start, End: end, Step: p.Step}
	v, _, err := p.api.QueryRange(context.TODO(), query, r)
	if err != nil {
		return v, err
	}
	return v, nil
}

// Find Job fills up job attributes if any
func (p *Prometheus) findJob(timestamp time.Time) (string, interface{}) {
	var jobName string
	var jobConfig interface{}
	for _, prometheusJob := range p.JobList {
		if timestamp.Before(prometheusJob.End) {
			jobName = prometheusJob.Name
			jobConfig = prometheusJob.JobConfig
		}
	}
	return jobName, jobConfig
}

// Parse vector parses results for an instant query
func (p *Prometheus) parseVector(metricName, query string, value model.Value, metrics *[]interface{}) error {
	data, ok := value.(model.Vector)
	if !ok {
		return fmt.Errorf("unsupported result format: %s", value.Type().String())
	}
	for _, vector := range data {
		jobName, jobConfig := p.findJob(vector.Timestamp.Time())

		m := createMetric(p.UUID, query, metricName, jobName, jobConfig, p.metadata, vector.Metric, vector.Value, vector.Timestamp.Time())
		*metrics = append(*metrics, m)
	}
	return nil
}

// Parse matrix parses results for an non-instant query
func (p *Prometheus) parseMatrix(metricName, query string, value model.Value, metrics *[]interface{}) error {
	data, ok := value.(model.Matrix)
	if !ok {
		return fmt.Errorf("unsupported result format: %s", value.Type().String())
	}
	for _, matrix := range data {
		for _, val := range matrix.Values {
			jobName, jobConfig := p.findJob(val.Timestamp.Time())

			m := createMetric(p.UUID, query, metricName, jobName, jobConfig, p.metadata, matrix.Metric, val.Value, val.Timestamp.Time())
			*metrics = append(*metrics, m)
		}
	}
	return nil
}

// Query prometheus query wrapper
func (p *Prometheus) query(query string, time time.Time) (model.Value, error) {
	var v model.Value
	v, _, err := p.api.Query(context.TODO(), query, time)
	if err != nil {
		return v, err
	}
	return v, nil
}

// Verifies prometheus connection
func (p *Prometheus) verifyConnection() error {
	_, err := p.api.Runtimeinfo(context.TODO())
	if err != nil {
		return err
	}
	return nil
}

// Create metric creates metric to be indexed
func createMetric(uuid, query, metricName, jobName string, jobConfig interface{}, metadata map[string]interface{}, labels model.Metric, value model.SampleValue, timestamp time.Time) metric {
	m := metric{
		Labels:     make(map[string]string),
		UUID:       uuid,
		Query:      query,
		MetricName: metricName,
		JobName:    jobName,
		JobConfig:  jobConfig,
		Timestamp:  timestamp,
		Metadata:   metadata,
	}
	for k, v := range labels {
		if k != "__name__" {
			m.Labels[string(k)] = string(v)
		}
	}
	if math.IsNaN(float64(value)) {
		m.Value = 0
	} else {
		m.Value = float64(value)
	}
	return m
}
