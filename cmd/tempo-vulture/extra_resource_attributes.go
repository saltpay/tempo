package main

// -extra-resource-attribute adds resource attributes to the expected trace before it's compared with
// the one read back. It's for a vulture whose traces go through a collector that adds resource
// attributes on the way to Tempo, such as Alloy adding deployment.scope.name: without it, every read
// would be counted as an incorrect result.

import (
	"flag"
	"fmt"
	"strings"

	"github.com/grafana/tempo/pkg/tempopb"
	v1_common "github.com/grafana/tempo/pkg/tempopb/common/v1"
	v1_resource "github.com/grafana/tempo/pkg/tempopb/resource/v1"
)

var tempoExtraResourceAttributes resourceAttrList

func init() {
	flag.Var(&tempoExtraResourceAttributes, "extra-resource-attribute", "Resource attribute in key=value format to inject into expected traces")
}

type resourceAttrList []*v1_common.KeyValue

func (r *resourceAttrList) String() string {
	var parts []string
	for _, kv := range *r {
		parts = append(parts, fmt.Sprintf("%s=%s", kv.Key, kv.Value.GetStringValue()))
	}
	return strings.Join(parts, " ")
}

func (r *resourceAttrList) Set(val string) error {
	parts := strings.SplitN(val, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("expected key=value format, got %q", val)
	}
	*r = append(*r, &v1_common.KeyValue{
		Key: strings.TrimSpace(parts[0]),
		Value: &v1_common.AnyValue{
			Value: &v1_common.AnyValue_StringValue{
				StringValue: strings.TrimSpace(parts[1]),
			},
		},
	})
	return nil
}

// addExtraResourceAttributes adds the -extra-resource-attribute attributes to every batch of t.
func addExtraResourceAttributes(t *tempopb.Trace) {
	if len(tempoExtraResourceAttributes) == 0 {
		return
	}
	for _, rs := range t.ResourceSpans {
		if rs.Resource == nil {
			rs.Resource = &v1_resource.Resource{}
		}
		rs.Resource.Attributes = append(rs.Resource.Attributes, tempoExtraResourceAttributes...)
	}
}
