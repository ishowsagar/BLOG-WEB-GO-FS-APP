package database

import (
	"database/sql"
	"log/slog"
)

// test supplied db connection alive or not
func PingTestingDb(db *sql.DB) error {
	err := db.Ping()
	if err != nil {
		return err
	}
	slog.Info("successfully pinged database🌟✅")
	return nil
}