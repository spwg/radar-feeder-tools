// Package radarstorage provides functionality for handling storage of radar
// data.
package radarstorage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/spwg/radar-feeder-tools/internal/history"
)

type flightRow struct {
	FlightDesignator string
	SeenTime         time.Time
}

// UploadToFlyPostgresInstance performs a bulk insert of the given
// [history.FlightObservation] records into the database.
func UploadToFlyPostgresInstance(ctx context.Context, db *sql.DB, flights map[history.FlightObservation]struct{}) error {
	var entries []*flightRow
	for k := range flights {
		entries = append(entries, &flightRow{k.Code, time.Unix(k.When, 0).UTC()})
	}
	// "on conflict do nothing" documented at
	// https://www.postgresql.org/docs/current/sql-insert.html. It does what is
	// sounds like. When the data would otherwise have a primary key conflict
	// (composite primary key on flight & seen time), it instead does nothing.
	// This is desirable because it means rerunning the inserts won't cause an
	// error, making this code simpler.
	stmt, err := db.PrepareContext(ctx, "insert into flights values ($1, $2) on conflict do nothing;")
	if err != nil {
		return fmt.Errorf("UploadToFlyPostgresInstance: %v", err)
	}
	defer stmt.Close()
	for _, e := range entries {
		if _, err := stmt.ExecContext(ctx, e.FlightDesignator, e.SeenTime.Format(time.RFC3339)); err != nil {
			return fmt.Errorf("UploadToFlyPostgresInstance: insert %+v: %v", e, err)
		}
	}
	return nil
}
