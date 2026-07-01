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

package ocpmetadata

import "testing"

func TestDetectStream(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "OKD version with okd identifier",
			version: "4.15.0-0.okd-2024-03-10-010116",
			want:    StreamOKD,
		},
		{
			name:    "OKD version uppercase",
			version: "4.15.0-0.OKD-2024-03-10-010116",
			want:    StreamOKD,
		},
		{
			name:    "OCP version simple",
			version: "4.16.19",
			want:    StreamOCP,
		},
		{
			name:    "OCP version with patch",
			version: "4.18.21",
			want:    StreamOCP,
		},
		{
			name:    "OCP version with RC",
			version: "4.17.0-rc.1",
			want:    StreamOCP,
		},
		{
			name:    "Empty version defaults to OCP",
			version: "",
			want:    StreamOCP,
		},
		{
			name:    "OKD another example",
			version: "4.14.0-0.okd-2023-11-14-101924",
			want:    StreamOKD,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectStream(tt.version)
			if got != tt.want {
				t.Errorf("detectStream(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}
