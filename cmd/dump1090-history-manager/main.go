// Binary dump1090-history-manager interacts with the historical files written
// from the dump1090-mutability program that the FR24 radar uploader runs.
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang/glog"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spwg/radar-feeder-tools/internal/history"
	"github.com/spwg/radar-feeder-tools/internal/radarstorage"
	"golang.org/x/sys/unix"
)

var (
	dataDir          = flag.String("data_dir", "", "")
	outDir           = flag.String("out_dir", "/tmp/dump1090-history-manager", "")
	uploadToPostgres = flag.Bool("postgres_upload", false, "Whether to upload to postgres.")
)

func run() error {
	defer glog.Flush()
	if *dataDir == "" {
		return fmt.Errorf("--data_dir is required")
	}
	ctx := context.Background()
	if err := history.MergeHistoryFiles(*dataDir, *outDir); err != nil {
		return fmt.Errorf("merge history: %v", err)
	}
	if !*uploadToPostgres {
		return nil
	}
	glog.Infof("Reading historical files.")
	flights, err := history.ReadHistoricalFiles(*dataDir)
	if err != nil {
		return err
	}
	glog.Infof("Flights %+v", flights)
	var errUpload error
	for retry := 0; retry < 3; retry++ {
		glog.Infof("Connecting to Postgres.")
		db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
		if err != nil {
			return err
		}
		defer db.Close()
		glog.Infof("Uploading to Postgres.")
		now := time.Now()
		errUpload = radarstorage.UploadToFlyPostgresInstance(ctx, db, flights)
		if errUpload != nil {
			if errors.Is(err, unix.ECONNRESET) {
				glog.Warningf("Error uploading to postgres: %v", errUpload)
				continue
			}
			return errUpload
		}
		glog.Infof("Uploaded in %v.", time.Since(now))
		return nil
	}
	return fmt.Errorf("failed to upload flights to the database in 3 tries: latest error: %v", errUpload)
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		glog.Exit(err)
	}
}
