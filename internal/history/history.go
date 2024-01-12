package history

import (
	"cmp"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"golang.org/x/exp/maps"
)

type historicalRadarEntry struct {
	Now      float64 `json:"now"` // unix seconds
	Aircraft []struct {
		Flight string `json:"flight"` // flight number or N number
	} `json:"aircraft"`
}

type aircraft struct {
	Code string `json:"code"` // flight number or N-number
	When int64  `json:"when"` // unix seconds
}

// Run is the entrypoint to historical file processing for files produced by fr24feeder.
func Run(dataDir, outDir string) error {
	// Read the current file into memory as a map for quick duplicate checks.
	b, err := os.ReadFile(path.Join(outDir, "all_aircraft.json"))
	if err != nil {
		return fmt.Errorf("reading all aircraft file: %v", err)
	}
	var current []aircraft
	if err := json.Unmarshal(b, &current); err != nil {
		return fmt.Errorf("unmarshaling all aircraft json: %v", err)
	}
	allAircraft := map[aircraft]struct{}{}
	for _, a := range current {
		allAircraft[a] = struct{}{}
	}

	// Scan the historical files and add them to the allAircraft hash table.
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return fmt.Errorf("read data dir: %v", err)
	}
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "history") {
			continue
		}
		b, err := os.ReadFile(path.Join(dataDir, e.Name()))
		if err != nil {
			return fmt.Errorf("reading historical files: %v", err)
		}
		entry := &historicalRadarEntry{}
		if err := json.Unmarshal(b, entry); err != nil {
			return fmt.Errorf("unmarshaling historical files: %v", err)
		}
		// Skip entries that don't have a timestamp.
		if entry.Now == 0. {
			continue
		}
		when := time.Unix(int64(math.Floor(entry.Now)), 0)
		for _, a := range entry.Aircraft {
			// Skip entries that have no flight number.
			if len(a.Flight) == 0 {
				continue
			}
			allAircraft[aircraft{a.Flight, when.Unix()}] = struct{}{}
		}
	}

	// Find the most recent entry, and use that to only keep 24h of entries.
	keys := maps.Keys(allAircraft)
	latestUnixSecond := slices.MaxFunc(keys, func(a, b aircraft) int {
		return cmp.Compare(a.When, b.When)
	}).When

	cutoff := time.Unix(latestUnixSecond, 0).Add(-24 * time.Hour)
	for _, k := range keys {
		if time.Unix(k.When, 0).Before(cutoff) {
			delete(allAircraft, k)
		}
	}

	// Write back the updated list of aircraft, sorted to make it easier to read output.
	current = maps.Keys(allAircraft)
	slices.SortStableFunc(current, func(a, b aircraft) int {
		return or(cmp.Compare(a.When, b.When), cmp.Compare(a.Code, b.Code))
	})
	b, err = json.MarshalIndent(current, "  ", "  ")
	if err != nil {
		return fmt.Errorf("marshaling all aircraft to json: %v", err)
	}
	if err := os.WriteFile(path.Join(outDir, "all_aircraft.json"), b, 0777); err != nil {
		return fmt.Errorf("writing all aircraft file: %v", err)
	}
	return nil
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
