package storage

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/eng-graph/eng-graph/internal/source"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS data_points (
			profile_name TEXT NOT NULL,
			id           TEXT NOT NULL,
			source       TEXT NOT NULL,
			kind         TEXT NOT NULL,
			author       TEXT NOT NULL,
			body         TEXT NOT NULL,
			timestamp    TEXT NOT NULL,
			context      TEXT NOT NULL DEFAULT '{}',
			PRIMARY KEY (profile_name, id)
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Insert(profileName string, dp source.DataPoint) error {
	ctx, err := json.Marshal(dp.Context)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT OR IGNORE INTO data_points (profile_name, id, source, kind, author, body, timestamp, context)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		profileName, dp.ID, dp.Source, string(dp.Kind), dp.Author, dp.Body,
		dp.Timestamp.Format(time.RFC3339), string(ctx),
	)
	return err
}

func (s *SQLiteStore) InsertBatch(profileName string, dps []source.DataPoint) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT OR IGNORE INTO data_points (profile_name, id, source, kind, author, body, timestamp, context)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, dp := range dps {
		ctx, err := json.Marshal(dp.Context)
		if err != nil {
			return err
		}
		_, err = stmt.Exec(
			profileName, dp.ID, dp.Source, string(dp.Kind), dp.Author, dp.Body,
			dp.Timestamp.Format(time.RFC3339), string(ctx),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) Exists(profileName, id string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(1) FROM data_points WHERE profile_name = ? AND id = ?`,
		profileName, id,
	).Scan(&count)
	return count > 0, err
}

func (s *SQLiteStore) LoadAll(profileName string) ([]source.DataPoint, error) {
	rows, err := s.db.Query(
		`SELECT id, source, kind, author, body, timestamp, context
		 FROM data_points WHERE profile_name = ? ORDER BY timestamp`,
		profileName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dps []source.DataPoint
	for rows.Next() {
		var dp source.DataPoint
		var ts, ctx string
		if err := rows.Scan(&dp.ID, &dp.Source, &dp.Kind, &dp.Author, &dp.Body, &ts, &ctx); err != nil {
			return nil, err
		}
		dp.Timestamp, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(ctx), &dp.Context); err != nil {
			return nil, err
		}
		dps = append(dps, dp)
	}
	return dps, rows.Err()
}

func (s *SQLiteStore) Count(profileName string) (int, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(1) FROM data_points WHERE profile_name = ?`,
		profileName,
	).Scan(&count)
	return count, err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
