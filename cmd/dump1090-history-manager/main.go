// Binary dump1090-history-manager interacts with the historical files written
// from the dump1090-mutability program that the FR24 radar uploader runs.
package main

import (
	"flag"
	"log"

	"github.com/spwg/radar-feeder-tools/internal/history"
)

var (
	dataDir          = flag.String("data_dir", "/run/dump1090-mutability", "")
	outDir           = flag.String("out_dir", "/tmp/dump1090-history-manager", "")
	uploadToPostgres = flag.Bool("postgres_upload", false, "Whether to upload to postgres.")
)

func main() {
	flag.Parse()
	if *uploadToPostgres {
		if err := history.UploadToFlyPostgresInstance(*dataDir); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := history.MergeHistoryFiles(*dataDir, *outDir); err != nil {
		log.Fatal(err)
	}
}
