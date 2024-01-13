// Binary history-accumulator writes scans the historical json files in the given --data_dir
// for updates to a file accumulating all past runs.
package main

import (
	"flag"
	"log"

	"github.com/spwg/radar-feeder-tools/internal/history"
)

var (
	dataDir = flag.String("data_dir", "/run/dump1090-mutability", "")
	outDir  = flag.String("out_dir", "/tmp/history-accumulator", "")
)

func main() {
	flag.Parse()
	if err := history.MergeHistoryFiles(*dataDir, *outDir); err != nil {
		log.Fatal(err)
	}
}
