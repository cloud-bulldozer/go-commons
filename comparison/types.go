package comparison

type Stat string

const (
	Min Stat = "min"
	Max Stat = "max"
	Avg Stat = "avg"
	Sum Stat = "sum"
)

type QueryStringResponse struct {
	Aggregations struct {
		stats `json:"stats"`
	} `json:"aggregations"`
}

type stats struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	Avg float64 `json:"avg"`
	Sum float64 `json:"sum"`
}
