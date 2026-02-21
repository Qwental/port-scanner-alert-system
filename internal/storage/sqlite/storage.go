package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Qwental/port-scanner-alert-system/internal/model"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type Storage struct {
	db  *sql.DB
	log *zap.SugaredLogger
}

func NewStorage(dbPath string, log *zap.SugaredLogger) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s := &Storage{db: db, log: log}
	if err := s.migrate(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS scan_results (
		ip         TEXT    NOT NULL,
		port       INTEGER NOT NULL,
		proto      TEXT    NOT NULL,
		banner     TEXT    DEFAULT '',
		first_seen DATETIME NOT NULL,
		last_seen  DATETIME NOT NULL,
		PRIMARY KEY (ip, port, proto)
	);`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func (s *Storage) Upsert(results []model.ScanResult) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			if rbErr := tx.Rollback(); rbErr != nil {
				s.log.Errorf("Rollback failed: %v", rbErr)
			}
		}
	}()

	query := `
	INSERT INTO scan_results (ip, port, proto, banner, first_seen, last_seen)
	VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(ip, port, proto) DO UPDATE SET
		last_seen = excluded.last_seen,
		banner = CASE 
			WHEN excluded.banner != '' THEN excluded.banner 
			ELSE scan_results.banner 
		END;`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	saved := 0

	for _, res := range results {
		_, execErr := stmt.Exec(
			res.IP,
			res.Port,
			res.Proto,
			res.Banner,
			now,
			now,
		)
		if execErr != nil {
			s.log.Errorf("Failed to upsert %s:%d/%s: %v", res.IP, res.Port, res.Proto, execErr)
			continue
		}
		saved++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	s.log.Infof("Upserted %d/%d records", saved, len(results))
	return nil
}

func (s *Storage) GetAll() (map[string]model.ScanResult, error) {
	query := `SELECT ip, port, proto, banner, first_seen, last_seen FROM scan_results`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query scan_results: %w", err)
	}
	defer rows.Close()

	results := make(map[string]model.ScanResult)

	for rows.Next() {
		var r model.ScanResult
		err := rows.Scan(
			&r.IP,
			&r.Port,
			&r.Proto,
			&r.Banner,
			&r.FirstSeen,
			&r.LastSeen,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results[r.Key()] = r
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil
}

func (s *Storage) Close() {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			s.log.Errorf("Failed to close database: %v", err)
		}
	}
}
