package comparison

import (
	"encoding/json"
	"fmt"
	"strings"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
)

type Comparator struct {
	client elasticsearch.Client
	index  string
}

// NewComparator returns a new comparator for the given index and elasticsearch client
func NewComparator(client elasticsearch.Client, index string) Comparator {
	return Comparator{
		client: client,
		index:  index,
	}
}

// Compare returns error if value does not meet the tolerance as
// compared with the field extracted from the given query
//
// Where field is the field we want to compare, query is the query string
// to use for the search, stat is the type of aggregation to compare with value
// and toleration is the percentaje difference tolerated, it can negative
// it returns an error when the field doesn't meet the tolerancy, and an
// informative message when it does
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
	if tolerancy >= 0 {
		baselineTolerancy := baseline * (100 - float64(tolerancy)) / 100
		if value < baselineTolerancy {
			return "", fmt.Errorf("with a tolerancy of %d%%: %.2f rps is %.2f%% lower than baseline: %.2f rps", tolerancy, value, 100-(value*100/baseline), baseline)
		}
	} else if tolerancy < 0 {
		baselineTolerancy := baseline * (100 + float64(tolerancy)) / 100
		if value > baselineTolerancy {
			return "", fmt.Errorf("with a tolerancy of %d%%: %.2f is %.2f%% rps higher than baseline: %.2f rps", tolerancy, value, (value*100/baseline)-100, baseline)
		}
	}
	return fmt.Sprintf("%2.f rps meets %d%% tolerancy against %.2f rps", value, tolerancy, baseline), nil
}

// queryStringStats perform a query of type query_string,to fetch the stats of a specific field
// this type of query accepts a simple query format similar to the kibana queries, i.e:
//
//	{
//	 "aggs": {
//	   "stats": {
//	     "stats": {
//	       "field": "our_field"
//	     }
//	   }
//	 },
//	 "query": {
//	   "query_string": {
//	     "query": "uuid.keyword: our_uuid AND param1.keyword: value1"
//	   }
//	 },
//	 "size": 0
//	}
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
		return stats{}, fmt.Errorf("error parsing the response body: %s", err)
	}

	return response.Aggregations.stats, nil
}
