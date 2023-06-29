package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/common/model"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/stretchr/testify/mock"
)

type flag int

type MockAPI struct {
	flag
	v1.API
	mock.Mock
}

// Override the QueryRange method with a mock implementation
func (m *MockAPI) QueryRange(context.Context, string, v1.Range, ...v1.Option) (model.Value, v1.Warnings, error) {
	if m.flag == 1 {
		samples := []model.SamplePair{
			{Value: 1, Timestamp: model.TimeFromUnix(1)},
			{Value: 2, Timestamp: model.TimeFromUnix(2)},
			{Value: 3, Timestamp: model.TimeFromUnix(3)},
		}

		// Create a SampleStream
		stream := model.SampleStream{
			Metric: model.Metric{},
			Values: samples,
		}

		// Wrap the SampleStream in a Matrix
		matrix := model.Matrix{&stream}
		return matrix, nil, nil
	}
	if m.flag == 2 {
		return model.Matrix{}, nil, fmt.Errorf("sample error")
	}
	return nil, nil, nil
}

func (m *MockAPI) Query(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error) {
	return nil, nil, nil
}

func (m *MockAPI) Runtimeinfo(ctx context.Context) (v1.RuntimeinfoResult, error) {
	return v1.RuntimeinfoResult{}, nil
}
