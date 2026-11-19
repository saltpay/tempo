package main

import (
	"testing"

	"github.com/grafana/dskit/flagext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/tempo/pkg/tempopb"
	v1_common "github.com/grafana/tempo/pkg/tempopb/common/v1"
)

func TestQueryTraceIgnoreResourceAttributes(t *testing.T) {
	addEnvironment := func(trace *tempopb.Trace) {
		for _, rs := range trace.ResourceSpans {
			rs.Resource.Attributes = append(rs.Resource.Attributes, &v1_common.KeyValue{
				Key:   "environment",
				Value: &v1_common.AnyValue{Value: &v1_common.AnyValue_StringValue{StringValue: "staging"}},
			})
		}
	}

	tests := []struct {
		name            string
		ignore          flagext.StringSlice
		traceOperation  func(*tempopb.Trace)
		expectedMetrics traceMetrics
	}{
		{
			name:            "queried trace has an attribute that isn't ignored",
			traceOperation:  addEnvironment,
			expectedMetrics: traceMetrics{requested: 1, incorrectResult: 1},
		},
		{
			name:            "queried trace has an ignored attribute",
			ignore:          flagext.StringSlice{"environment"},
			traceOperation:  addEnvironment,
			expectedMetrics: traceMetrics{requested: 1},
		},
		{
			// Set on a vulture whose traces don't go through the collector that adds the attribute.
			name:            "ignored attribute isn't in the queried trace",
			ignore:          flagext.StringSlice{"environment"},
			traceOperation:  func(*tempopb.Trace) {},
			expectedMetrics: traceMetrics{requested: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempoIgnoreResourceAttributes = tt.ignore
			defer func() { tempoIgnoreResourceAttributes = nil }()

			metrics, err := doQueryTrace(tt.traceOperation, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMetrics, metrics)
		})
	}
}
