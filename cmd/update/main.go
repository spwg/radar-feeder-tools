// Binary update writes scans the historical json files in the given --data_dir
// for updates to a file accumulating all past runs.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/exp/maps"
)

var (
	dataDir = flag.String("data_dir", "/run/dump1090-mutability", "")
)

type historicalRadarEntry struct {
	Now      float32 `json:"now"` // unix seconds
	Aircraft []struct {
		Flight string `json:"flight"` // flight number or N number
	} `json:"aircraft"`
}

type aircraft struct {
	Code string `json:"code"` // flight number or N-number
	When int64  `json:"when"` // unix seconds
}

func run() error {
	entries, err := os.ReadDir(*dataDir)
	if err != nil {
		return err
	}

	// Read the current file into memory as a map for quick duplicate checks.
	b, err := os.ReadFile("/run/historical-radar-uploader/all_aircraft.json")
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

	// Scan the historical files, suspressing those with missing information.
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "history") {
			continue
		}
		b, err := os.ReadFile(path.Join(*dataDir, e.Name()))
		if err != nil {
			return fmt.Errorf("reading historical files: %v", err)
		}
		entry := &historicalRadarEntry{}
		if err := json.Unmarshal(b, entry); err != nil {
			return fmt.Errorf("unmarshaling historical files: %v", err)
		}
		// Skip entries that don't have a timestamp.
		if entry.Now == 0 {
			continue
		}
		when := time.Unix(int64(entry.Now), 0)
		for _, a := range entry.Aircraft {
			// Skip entries that have no flight number.
			if len(a.Flight) == 0 {
				continue
			}
			allAircraft[aircraft{a.Flight, when.Unix()}] = struct{}{}
		}
	}

	b, err = json.Marshal(maps.Keys(allAircraft))
	if err != nil {
		return fmt.Errorf("marshaling all aircraft to json: %v", err)
	}
	if err := os.WriteFile("/run/historical-radar-uploader/all_aircraft.json", b, 0777); err != nil {
		return fmt.Errorf("writing all aircraft file: %v", err)
	}
	return nil
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
