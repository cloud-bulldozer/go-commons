package indexers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
	opensearchapi "github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	opensearchutil "github.com/opensearch-project/opensearch-go/v2/opensearchutil"
	requestsigner "github.com/opensearch-project/opensearch-go/v2/signer/awsv2"
)

const indexer = "opensearch"

var OpenSearchServiceName = map[bool]string{
  false: "es",
  true:  "aoss",
}

// OpenSearch OpenSearch instance
type OpenSearch struct {
	client *opensearch.Client
	index  string
}

// Init function
func init() {
	indexerMap[indexer] = &OpenSearch{}
}

// Returns new indexer for OpenSearch
func (OpenSearchIndexer *OpenSearch) new(indexerConfig IndexerConfig) error {
	OpenSearchConfig := indexerConfig
	if OpenSearchConfig.Index == "" {
		return fmt.Errorf("index name not specified")
	}
	OpenSearchIndex := strings.ToLower(OpenSearchConfig.Index)
  ctx := context.Background()
	awsCfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load aws configuraiton: %v", err) // Do not log.fatal in a production ready app.
	}
  signer, err := requestsigner.NewSignerWithService(
    awsCfg,
		OpenSearchServiceName[OpenSearchConfig.Serverless],
	)
	if err != nil {
		return fmt.Errorf("failed to create AWS api signer for OpenSearch: %v", err) // Do not log.fatal in a production ready app.
	}
  cfg := opensearch.Config{
  	Addresses: OpenSearchConfig.Servers,
  	Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: OpenSearchConfig.InsecureSkipVerify}},
    Signer:    signer,
  }
	OpenSearchClient, err := opensearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("error creating the OpenSearch client: %s", err)
	}
  fmt.Println(OpenSearchServiceName[OpenSearchConfig.Serverless])
	if OpenSearchConfig.Serverless {
    ping := opensearchapi.PingRequest{}

	  resp, err := ping.Do(ctx, OpenSearchClient)
	  if err != nil {
	  	return fmt.Errorf("failed to ping opensearch serverless: %v", err)
	  }
	  defer resp.Body.Close()
  } else {
    r1, err := OpenSearchClient.Cluster.Health()
  	if err != nil {
  		return fmt.Errorf("OpenSearch health check failed: %s", err)
  	}
  	if r1.StatusCode != 200 {
  		return fmt.Errorf("unexpected OpenSearch status code: %d", r1.StatusCode)
  	}
  }
	OpenSearchIndexer.client = OpenSearchClient
	OpenSearchIndexer.index = OpenSearchIndex
	fmt.Println(OpenSearchIndex)
  r, _ := OpenSearchIndexer.client.Indices.Exists([]string{OpenSearchIndex})
	if r.IsError() {
		r, _ = OpenSearchIndexer.client.Indices.Create(OpenSearchIndex)
		if r.IsError() {
			return fmt.Errorf("error creating index %s on OpenSearch: %s", OpenSearchIndex, r.String())
		}
	  fmt.Println("Created index")
	}
	return nil
}

// Index uses bulkIndexer to index the documents in the given index
func (OpenSearchIndexer *OpenSearch) Index(documents []interface{}, opts IndexingOpts) (string, error) {
	var statString string
	var indexerStatsLock sync.Mutex
	indexerStats := make(map[string]int)
	hasher := sha256.New()
	bi, err := opensearchutil.NewBulkIndexer(opensearchutil.BulkIndexerConfig{
		Client:     OpenSearchIndexer.client,
		Index:      OpenSearchIndexer.index,
		FlushBytes: 5e+6,
		NumWorkers: runtime.NumCPU(),
		Timeout:    10 * time.Minute, // TODO: hardcoded
	})
	if err != nil {
		return "", fmt.Errorf("Error creating the indexer: %s", err)
	}
	start := time.Now().UTC()
	for _, document := range documents {
		j, err := json.Marshal(document)
		if err != nil {
			return "", fmt.Errorf("Cannot encode document %s: %s", document, err)
		}
		hasher.Write(j)
		err = bi.Add(
			context.Background(),
			opensearchutil.BulkIndexerItem{
				Action:     "index",
				Body:       bytes.NewReader(j),
				DocumentID: hex.EncodeToString(hasher.Sum(nil)),
				OnSuccess: func(c context.Context, bii opensearchutil.BulkIndexerItem, biri opensearchutil.BulkIndexerResponseItem) {
					indexerStatsLock.Lock()
					defer indexerStatsLock.Unlock()
					indexerStats[biri.Result]++
				},
			},
		)
		if err != nil {
			return "", fmt.Errorf("Unexpected OpenSearch indexing error: %s", err)
		}
		hasher.Reset()
	}
	if err := bi.Close(context.Background()); err != nil {
		return "", fmt.Errorf("Unexpected OpenSearch error: %s", err)
	}
	dur := time.Since(start)
	for stat, val := range indexerStats {
		statString += fmt.Sprintf(" %s=%d", stat, val)
	}
	return fmt.Sprintf("Indexing finished in %v:%v", dur.Truncate(time.Millisecond), statString), nil
}
