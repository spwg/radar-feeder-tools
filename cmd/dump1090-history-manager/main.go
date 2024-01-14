// Binary dump1090-history-manager interacts with the historical files written
// from the dump1090-mutability program that the FR24 radar uploader runs.
package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"time"

	"github.com/golang/glog"
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
	defer glog.Flush()
	ctx := context.Background()
	switch {
	case *uploadToPostgres:
		glog.Infof("Reading historical files.")
		flights, err := history.ReadHistoricalFiles(*dataDir)
		if err != nil {
			return err
		}
		glog.Infof("Flights %+v", flights)
		glog.Infof("Connecting to Postgres.")
		db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
		if err != nil {
			return err
		}
		defer db.Close()

		glog.Infof("Uploading to Postgres.")
		now := time.Now()
		if err := radarstorage.UploadToFlyPostgresInstance(ctx, db, flights); err != nil {
			return err
		}
		glog.Infof("Uploaded in %v.", time.Since(now))
		return nil
	default:
		return history.MergeHistoryFiles(*dataDir, *outDir)
	}
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		glog.Exit(err)
	}
}
