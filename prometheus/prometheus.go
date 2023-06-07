// Copyright 2023 The go-commons Authors.
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
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/montanaflynn/stats"
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

// NewClient creates a prometheus struct instance with the given parameters
func NewClient(url, token, username, password string, tlsSkipVerify bool) (*Prometheus, error) {
	prometheus := Prometheus{
		Endpoint: url,
	}
	cfg := api.Config{
		Address: url,
		RoundTripper: authTransport{
			Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, TLSClientConfig: &tls.Config{InsecureSkipVerify: tlsSkipVerify}},
			token:     token,
			username:  username,
			password:  password,
		},
	}
	c, err := api.NewClient(cfg)
	if err != nil {
		return &prometheus, err
	}
	prometheus.api = apiv1.NewAPI(c)
	// Verify Prometheus connection prior returning
	if err := prometheus.verifyConnection(); err != nil {
		return &prometheus, err
	}
	return &prometheus, nil
}

// Query prometheus query wrapper
func (p *Prometheus) Query(query string, time time.Time) (model.Value, error) {
	var v model.Value
	v, _, err := p.api.Query(context.TODO(), query, time)
	if err != nil {
		return v, err
	}
	return v, nil
}

// QueryRange prometheus queryRange wrapper
func (p *Prometheus) QueryRange(query string, start, end time.Time, step time.Duration) (model.Value, error) {
	var v model.Value
	r := apiv1.Range{Start: start, End: end, Step: step}
	v, _, err := p.api.QueryRange(context.TODO(), query, r)
	if err != nil {
		return v, err
	}
	return v, nil
}

// QueryRangeAggregation returns the aggregation from the given query
// if the query returns multiple timeseries, their data points are aggregated into a single one
func (p *Prometheus) QueryRangeAggregatedTS(query string, start, end time.Time, step time.Duration, aggregation Aggregation) (float64, error) {
	var err error
	var datapoints []float64
	var result float64
	v, err := p.QueryRange(query, start, end, step)
	if err != nil {
		return result, err
	}
	data, ok := v.(model.Matrix)
	if !ok {
		return result, fmt.Errorf("result format is not a range vector: %s", data.Type().String())
	}
	for _, ts := range data {
		for _, dp := range ts.Values {
			datapoints = append(datapoints, float64(dp.Value))
		}
	}
	switch aggregation {
	case Avg:
		result, err = stats.Mean(datapoints)
	case Max:
		result, err = stats.Max(datapoints)
	case Min:
		result, err = stats.Min(datapoints)
	case P99, P95, P90, P50:
		percentile, _ := strconv.ParseFloat(string(aggregation), 64)
		stats.Percentile(datapoints, percentile)
	case Stdev:
		result, err = stats.StandardDeviation(datapoints)
	default:
		return result, fmt.Errorf("aggregation not supported: %s", aggregation)
	}
	return result, err
}

// Verifies prometheus connection
func (p *Prometheus) verifyConnection() error {
	_, err := p.api.Runtimeinfo(context.TODO())
	if err != nil {
		return err
	}
	return nil
}
