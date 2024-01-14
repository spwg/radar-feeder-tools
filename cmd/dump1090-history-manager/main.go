// Binary dump1090-history-manager interacts with the historical files written
// from the dump1090-mutability program that the FR24 radar uploader runs.
package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spwg/radar-feeder-tools/internal/history"
	"github.com/spwg/radar-feeder-tools/internal/radarstorage"
)

var (
	dataDir          = flag.String("data_dir", "/run/dump1090-mutability", "")
	outDir           = flag.String("out_dir", "/tmp/dump1090-history-manager", "")
	uploadToPostgres = flag.Bool("postgres_upload", false, "Whether to upload to postgres.")
)

func run() error {
	ctx := context.Background()
	switch {
	case *uploadToPostgres:
		flights, err := history.ReadHistoricalFiles(*dataDir)
		if err != nil {
			return err
		}
		db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
		if err != nil {
			return err
		}
		defer db.Close()
		if err := radarstorage.UploadToFlyPostgresInstance(ctx, db, flights); err != nil {
			return err
		}
		return nil
	default:
		return history.MergeHistoryFiles(*dataDir, *outDir)
	}
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
