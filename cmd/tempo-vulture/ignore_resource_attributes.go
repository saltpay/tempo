package main

// -ignore-resource-attribute drops resource attributes from the trace read back before it's compared
// with the expected one. It's for a vulture whose traces go through a collector that adds resource
// attributes with values the vulture can't predict, such as Alloy's k8s metadata processor adding
// k8s.pod.name: without it, every read would be counted as an incorrect result.

import (
	"flag"

	"github.com/grafana/dskit/flagext"

	"github.com/grafana/tempo/pkg/tempopb"
)

var tempoIgnoreResourceAttributes flagext.StringSlice

func init() {
	flag.Var(&tempoIgnoreResourceAttributes, "ignore-resource-attribute", "Resource attribute key to ignore when comparing expected and queried traces")
}

// removeIgnoredResourceAttributes drops the -ignore-resource-attribute keys from every batch of t.
func removeIgnoredResourceAttributes(t *tempopb.Trace) {
	if len(tempoIgnoreResourceAttributes) == 0 {
		return
	}
	ignore := make(map[string]struct{}, len(tempoIgnoreResourceAttributes))
	for _, k := range tempoIgnoreResourceAttributes {
		ignore[k] = struct{}{}
	}

	for _, rs := range t.ResourceSpans {
		if rs.Resource == nil {
			continue
		}
		filtered := rs.Resource.Attributes[:0]
		for _, kv := range rs.Resource.Attributes {
			if _, ok := ignore[kv.Key]; !ok {
				filtered = append(filtered, kv)
			}
		}
		rs.Resource.Attributes = filtered
	}
}
