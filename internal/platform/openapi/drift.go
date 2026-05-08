package openapi

import "sort"

// DriftResult stores route-contract comparison result.
type DriftResult struct {
	MissingInSpec    []RouteKey
	MissingInRuntime []RouteKey
}

// HasDrift returns true when either side has missing routes.
func (d DriftResult) HasDrift() bool {
	return len(d.MissingInSpec) > 0 || len(d.MissingInRuntime) > 0
}

// CompareRouteSets compares runtime routes against spec routes.
func CompareRouteSets(runtime map[RouteKey]struct{}, spec map[RouteKey]struct{}) DriftResult {
	res := DriftResult{}
	for route := range runtime {
		if _, ok := spec[route]; !ok {
			res.MissingInSpec = append(res.MissingInSpec, route)
		}
	}
	for route := range spec {
		if _, ok := runtime[route]; !ok {
			res.MissingInRuntime = append(res.MissingInRuntime, route)
		}
	}

	sort.Slice(res.MissingInSpec, func(i, j int) bool {
		return res.MissingInSpec[i].String() < res.MissingInSpec[j].String()
	})
	sort.Slice(res.MissingInRuntime, func(i, j int) bool {
		return res.MissingInRuntime[i].String() < res.MissingInRuntime[j].String()
	})

	return res
}
