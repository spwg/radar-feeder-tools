package history

import (
	"cmp"
	"encoding/json"
	"errors"
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

// FlightObservation is an observation of a flight at a certain point in time.
type FlightObservation struct {
	Code string `json:"code"` // flight number or N-number
	When int64  `json:"when"` // unix seconds
}

// MergeHistoryFiles combines the history_*.json files from dataDir (absolute
// path of a directory) and writes them into all_aircraft.json in outDir
// (absolute path of a directory).
func MergeHistoryFiles(dataDir, outDir string) error {
	// Read the current file into memory as a map for quick duplicate checks.
	// But first, make sure the file exists.
	if err := os.MkdirAll(outDir, 0777); err != nil {
		return err
	}
	p := path.Join(outDir, "all_aircraft.json")
	b, err := os.ReadFile(p)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("reading all aircraft file: %v", err)
		}
		b = []byte("[]")
	}
	var current []FlightObservation
	if err := json.Unmarshal(b, &current); err != nil {
		return fmt.Errorf("unmarshaling all aircraft json: %v", err)
	}
	allAircraft := map[FlightObservation]struct{}{}
	for _, a := range current {
		allAircraft[a] = struct{}{}
	}

	// Merge the new entries from the history files into the current ones.
	fromHistory, err := ReadHistoricalFiles(dataDir)
	if err != nil {
		return err
	}
	for k := range fromHistory {
		allAircraft[k] = struct{}{}
	}

	// Find the most recent entry, and use that to only keep 24h of entries.
	keys := maps.Keys(allAircraft)
	var latestUnixSecond int64
	if len(keys) == 0 {
		latestUnixSecond = math.MinInt64
	} else {
		latestUnixSecond = slices.MaxFunc(keys, func(a, b FlightObservation) int {
			return cmp.Compare(a.When, b.When)
		}).When
	}

	cutoff := time.Unix(latestUnixSecond, 0).Add(-24 * time.Hour)
	for _, k := range keys {
		if time.Unix(k.When, 0).Before(cutoff) {
			delete(allAircraft, k)
		}
	}

	// Write back the updated list of aircraft, sorted to make it easier to read output.
	current = maps.Keys(allAircraft)
	slices.SortStableFunc(current, func(a, b FlightObservation) int {
		return cmp.Or(cmp.Compare(a.When, b.When), cmp.Compare(a.Code, b.Code))
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

// ReadHistoricalFiles scans all the history_*.json files in dataDir (absolute
// path of a directory) and merges the flight entries found in them. The
// returned slice will not have duplicates with respect to flight number &
// observation time, though it may have multiple entries for the same flight.
func ReadHistoricalFiles(dataDir string) (map[FlightObservation]struct{}, error) {
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read data dir: %v", err)
	}
	allAircraft := make(map[FlightObservation]struct{})
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "history") {
			continue
		}
		b, err := os.ReadFile(path.Join(dataDir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading historical files: %v", err)
		}
		entry := &historicalRadarEntry{}
		if err := json.Unmarshal(b, entry); err != nil {
			return nil, fmt.Errorf("unmarshaling historical files: %v", err)
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
			allAircraft[FlightObservation{a.Flight, when.Unix()}] = struct{}{}
		}
	}
	return allAircraft, nil
}
