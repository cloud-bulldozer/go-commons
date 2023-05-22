package comparison

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
)

type Comparator struct {
	client elasticsearch.Client
	index  string
}

// NewComparator returns a new comparator for the given index and elasticsearch instance
func NewComparator(cli elasticsearch.Client, index string) (Comparator, error) {
	return Comparator{
		client: cli,
		index:  index}, nil
}

// Compare returns error if value does not meet the tolerance as
// compared with the field extracted from the given query
//
// Where field is the field we want to compare, query is the query string
// to use for the search, stat is the type of aggregation to compare with value
// and toleration is the percentaje difference tolerated, it can negative
func (c *Comparator) Compare(field, query string, stat Stat, value float64, tolerancy int) (string, error) {
	stats, err := c.queryStringStats(field, query)
	var baseline float64
	if err != nil {
		return "", err
	}
	switch stat {
	case Avg:
		baseline = stats.Avg
	case Max:
		baseline = stats.Max
	case Min:
		baseline = stats.Min
	case Sum:
		baseline = stats.Sum
	}
	relativeChange := (value - baseline) * 100 / baseline
	if tolerancy >= 0 {
		if relativeChange < float64(tolerancy) {
			return "", fmt.Errorf("%f doesn't meet tolerance (%d) against %f", value, tolerancy, baseline)
		}
	} else if tolerancy < 0 {
		if relativeChange > float64(tolerancy) {
			return "", fmt.Errorf("%f doesn't meet tolerance (%d) against %f", value, tolerancy, baseline)
		}
	}
	return fmt.Sprintf("%f meets %d tolerancy against %f", value, tolerancy, baseline), nil
}

func (c *Comparator) queryStringStats(field, query string) (stats, error) {
	var response QueryStringResponse
	var queryStringRequest map[string]interface{} = map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": query,
			},
		},
		"aggs": map[string]interface{}{
			"stats": map[string]interface{}{
				"stats": map[string]string{
					"field": field,
				},
			},
		},
	}
	queryStringRequestJSON, _ := json.Marshal(queryStringRequest)
	res, err := c.client.Search(
		c.client.Search.WithBody(strings.NewReader(string(queryStringRequestJSON))),
		c.client.Search.WithIndex(c.index),
	)
	if err != nil {
		return stats{}, err
	}
	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			return stats{}, fmt.Errorf("error parsing the response body: %s", err)
		} else {
			// Return the response status and error information.
			return stats{}, fmt.Errorf("%s %s %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}
	return response.Aggregations.stats, nil
}
