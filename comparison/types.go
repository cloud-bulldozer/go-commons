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

package comparison

type Stat string

// Constant for all the stats
const (
	Min Stat = "min"
	Max Stat = "max"
	Avg Stat = "avg"
	Sum Stat = "sum"
)

// Type to store query response
type QueryStringResponse struct {
	Aggregations struct {
		stats `json:"stats"`
	} `json:"aggregations"`
}

// Type to store the stats
type stats struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	Avg float64 `json:"avg"`
	Sum float64 `json:"sum"`
}
