package history_accumulator_test

import (
	"cmp"
	_ "embed"
	"encoding/json"
	"os"
	"path"
	"slices"
	"testing"

	cmpdiff "github.com/google/go-cmp/cmp"
	"github.com/spwg/radar-feeder-tools/internal/history"
)

var (
	//go:embed all_aircraft.json
	allAircraft []byte
	//go:embed history_47.json
	history47 []byte
)

type aircraft struct {
	Code string `json:"code"`
	When int64  `json:"when"`
}

func TestHistoryAccumulator(t *testing.T) {
	dataDir := t.TempDir()
	outDir := t.TempDir()
	if err := os.WriteFile(path.Join(outDir, "all_aircraft.json"), allAircraft, 0777); err != nil {
		t.Error(err)
	}
	if err := os.WriteFile(path.Join(dataDir, "history_47.json"), history47, 0777); err != nil {
		t.Error(err)
	}
	if err := history.MergeHistoryFiles(dataDir, outDir); err != nil {
		t.Errorf("Run(%q, %q): %v", dataDir, outDir, err)
	}
	b, err := os.ReadFile(path.Join(outDir, "all_aircraft.json"))
	if err != nil {
		t.Error(err)
	}
	var got []*aircraft
	if err := json.Unmarshal(b, &got); err != nil {
		t.Error(err)
	}
	want := []*aircraft{
		{Code: "405LP   ", When: 1705012736},
		{Code: "405LP   ", When: 1705012864},
		{Code: "405LP   ", When: 1705012992},
		{Code: "DAL798  ", When: 1705026154},
		{Code: "N920PD  ", When: 1705026154},
	}
	less := func(a, b *aircraft) int {
		return or(cmp.Compare(a.Code, b.Code), cmp.Compare(a.When, b.When))
	}
	slices.SortStableFunc(got, less)
	if diff := cmpdiff.Diff(got, want); diff != "" {
		t.Error(diff)
	}
}

func or[T cmp.Ordered](args ...T) T {
	var zero T
	for _, e := range args {
		if e != zero {
			return e
		}
	}
	return zero
}
