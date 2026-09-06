package profiler

import (
	"testing"

	googlepprof "github.com/google/pprof/profile"
)

func TestBuildCPUProfile(t *testing.T) {
	rootFunction := &googlepprof.Function{ID: 1, Name: "main.run"}
	leafFunction := &googlepprof.Function{ID: 2, Name: "service.query"}
	rootLocation := &googlepprof.Location{ID: 1, Line: []googlepprof.Line{{Function: rootFunction}}}
	leafLocation := &googlepprof.Location{ID: 2, Line: []googlepprof.Line{{Function: leafFunction}}}
	profile := &googlepprof.Profile{
		SampleType: []*googlepprof.ValueType{{Type: "samples", Unit: "count"}, {Type: "cpu", Unit: "nanoseconds"}},
		Sample:     []*googlepprof.Sample{{Location: []*googlepprof.Location{leafLocation, rootLocation}, Value: []int64{1, 20_000_000}}},
	}

	result, err := buildCPUProfile(profile, 30)
	if err != nil {
		t.Fatalf("buildCPUProfile() error = %v", err)
	}
	if result.Total != 20_000_000 || result.SampleUnit != "nanoseconds" {
		t.Fatalf("unexpected profile summary: %+v", result)
	}
	if len(result.Root.Children) != 1 || result.Root.Children[0].Name != "main.run" {
		t.Fatalf("unexpected root children: %+v", result.Root.Children)
	}
	if len(result.Root.Children[0].Children) != 1 || result.Root.Children[0].Children[0].Name != "service.query" {
		t.Fatalf("unexpected leaf: %+v", result.Root.Children[0].Children)
	}
}
