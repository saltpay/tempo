package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grafana/tempo/pkg/tempopb"
	v1_common "github.com/grafana/tempo/pkg/tempopb/common/v1"
)

func TestResourceAttrListSet(t *testing.T) {
	var r resourceAttrList
	require.NoError(t, r.Set("deployment.scope.name=shared"))
	require.NoError(t, r.Set(" a = b=c "))
	assert.Equal(t, "deployment.scope.name=shared a=b=c", r.String())
	assert.Error(t, r.Set("no-value"))
}

func TestQueryTraceExtraResourceAttributes(t *testing.T) {
	environment := &v1_common.KeyValue{
		Key:   "environment",
		Value: &v1_common.AnyValue{Value: &v1_common.AnyValue_StringValue{StringValue: "staging"}},
	}
	addEnvironment := func(trace *tempopb.Trace) {
		for _, rs := range trace.ResourceSpans {
			rs.Resource.Attributes = append(rs.Resource.Attributes, environment)
		}
	}

	tests := []struct {
		name            string
		extra           resourceAttrList
		traceOperation  func(*tempopb.Trace)
		expectedMetrics traceMetrics
	}{
		{
			name:            "queried trace has an attribute that isn't configured",
			traceOperation:  addEnvironment,
			expectedMetrics: traceMetrics{requested: 1, incorrectResult: 1},
		},
		{
			name:            "configured attribute is in the queried trace",
			extra:           resourceAttrList{environment},
			traceOperation:  addEnvironment,
			expectedMetrics: traceMetrics{requested: 1},
		},
		{
			// Set on a vulture whose traces don't go through the collector that adds the attribute.
			name:            "configured attribute isn't in the queried trace",
			extra:           resourceAttrList{environment},
			traceOperation:  func(*tempopb.Trace) {},
			expectedMetrics: traceMetrics{requested: 1, incorrectResult: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempoExtraResourceAttributes = tt.extra
			defer func() { tempoExtraResourceAttributes = nil }()

			metrics, err := doQueryTrace(tt.traceOperation, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMetrics, metrics)
		})
	}
}
