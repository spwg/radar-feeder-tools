// Package radarstorage provides functionality for handling storage of radar
// data.
package radarstorage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/spwg/radar-feeder-tools/internal/history"
)

// UploadToFlyPostgresInstance performs a bulk insert of the given
// [history.FlightObservation] records into the database.
func UploadToFlyPostgresInstance(ctx context.Context, db *sql.DB, flights map[history.FlightObservation]struct{}) error {
	if len(flights) == 0 {
		return nil
	}
	var qb strings.Builder
	qb.WriteString("insert into flights (flight_designator, seen_time) values ")
	var rows []any
	i := 1
	for k := range flights {
		qb.WriteString(fmt.Sprintf("($%d, $%d), ", i, i+1))
		i = i + 2
		rows = append(rows, k.Code, time.Unix(k.When, 0).UTC().Format(time.RFC3339))
	}
	q := qb.String()
	// Remove the trailing comma and space.
	q = q[:len(q)-2]
	// "on conflict do nothing" documented at
	// https://www.postgresql.org/docs/current/sql-insert.html. It does what is
	// sounds like. When the data would otherwise have a primary key conflict
	// (composite primary key on flight & seen time), it instead does nothing.
	// This is desirable because it means rerunning the inserts won't cause an
	// error, making this code simpler.
	q += " on conflict do nothing;"
	result, err := db.ExecContext(ctx, q, rows...)
	if err != nil {
		return fmt.Errorf("UploadToFlyPostgresInstance: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	glog.Infof("Inserted %v rows.", n)
	return nil
}
