package store

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

const batchSize = 10000

type SQLiteStore struct {
	readDB  *sql.DB
	writeDB *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	dsn := dbPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"

	writeDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open write db: %w", err)
	}
	writeDB.SetMaxOpenConns(1) //single writer connection

	readDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		writeDB.Close()
		return nil, fmt.Errorf("failed to open read Db: %w", err)
	}

	store := &SQLiteStore{writeDB: writeDB, readDB: readDB}

	if err := store.initDB(); err != nil {
		writeDB.Close()
		readDB.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) initDB() error {

	driver, err := sqlite.WithInstance(s.writeDB, &sqlite.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	migrationFS, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create migration sub filesystem: %w", err)
	}

	sourceDriver, err := iofs.New(migrationFS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source driver: %w", err)
	}

	d, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := d.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (s *SQLiteStore) Close() error {
	return errors.Join(s.readDB.Close(), s.writeDB.Close())
}

func (s *SQLiteStore) StoreTrace(ctx context.Context, trace Trace, spans []Span) error {

	tx, err := s.writeDB.BeginTx(ctx, nil)

	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := s.upsertTrace(ctx, tx, trace); err != nil {
		return err
	}

	if err := s.bulkInsertSpans(ctx, tx, spans); err != nil {
		return err
	}

	return tx.Commit()

}

func (s *SQLiteStore) upsertTrace(ctx context.Context, tx *sql.Tx, trace Trace) error {
	now := time.Now().UnixNano()

	query := `INSERT INTO traces (trace_id, root_service, environment, root_name, start_time_ns, end_time_ns, span_count, error, updated_at_ns) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT(trace_id) DO UPDATE SET start_time_ns = MIN(start_time_ns, excluded.start_time_ns), end_time_ns = MAX(end_time_ns, excluded.end_time_ns), span_count = span_count + excluded.span_count, error = error OR excluded.error, updated_at_ns = excluded.updated_at_ns`

	_, err := tx.ExecContext(ctx, query,
		trace.TraceID,
		trace.RootService,
		trace.Environment,
		trace.RootName,
		trace.StartTime.UnixNano(),
		trace.EndTime.UnixNano(),
		trace.SpanCount,
		trace.Error,
		now,
	)

	return err
}

func (s *SQLiteStore) bulkInsertSpans(ctx context.Context, tx *sql.Tx, spans []Span) error {

	if len(spans) == 0 {
		return nil
	}

	for left := 0; left < len(spans); left += batchSize {

		right := left + batchSize

		if right > len(spans) {
			right = len(spans)
		}

		spanBatch := spans[left:right]

		placeholders := make([]string, 0, len(spanBatch))

		query := `INSERT OR IGNORE INTO spans (span_id, trace_id, parent_span_id, service, name, kind, start_time_ns, end_time_ns, status_code, status_message, attributes, events) VALUES`

		//underlying array size is set to the total number of spans * the number of fields in each span (currently 12)
		values := make([]any, 0, len(spanBatch)*12)

		for _, span := range spanBatch {
			attributes, err := json.Marshal(span.Attributes)
			if err != nil {
				attributes = []byte("{}")
			}

			events, err := json.Marshal(span.Events)
			if err != nil {
				events = []byte("[]")
			}

			placeholders = append(placeholders, "(?,?,?,?,?,?,?,?,?,?,?,?)")

			values = append(values,
				span.SpanID,
				span.TraceID,
				span.ParentSpanID,
				span.Service,
				span.Name,
				int32(span.Kind),
				span.StartTime.UnixNano(),
				span.EndTime.UnixNano(),
				int32(span.StatusCode),
				span.StatusMessage,
				attributes,
				events,
			)

		}

		query += strings.Join(placeholders, ",")

		_, err := tx.ExecContext(ctx, query, values...)
		if err != nil {
			return err
		}

	}

	return nil

}

func (s *SQLiteStore) StoreLogs(ctx context.Context, logs []Log) error {

	if len(logs) == 0 {
		return nil
	}

	for left := 0; left < len(logs); left += batchSize {

		right := left + batchSize

		if right > len(logs) {
			right = len(logs)
		}

		batch := logs[left:right]

		query := `INSERT INTO logs (trace_id, span_id, timestamp_ns, severity_text, severity_number, body, service, environment, attributes, resource_attributes) VALUES`

		values := make([]any, 0, len(batch)*10) //size set to number of fields inserting into * lenght of batch

		placeholders := make([]string, 0, len(batch))

		for _, log := range batch {
			attributes, err := json.Marshal(log.Attributes)
			if err != nil {
				attributes = []byte("{}")
			}
			resourceAttributes, err := json.Marshal(log.ResourceAttributes)
			if err != nil {
				attributes = []byte("{}")
			}

			body, err := json.Marshal(log.Body)

			if err != nil {
				body = []byte("{}")
			}

			values = append(values,
				log.TraceID,
				log.SpanID,
				log.Timestamp.UnixNano(),
				log.SeverityText,
				int32(log.SeverityNumber),
				body,
				log.Service,
				log.Environment,
				attributes,
				resourceAttributes)

			placeholders = append(placeholders, "(?,?,?,?,?,?,?,?,?,?)")
		}

		query += strings.Join(placeholders, ",")

		_, err := s.writeDB.ExecContext(ctx, query, values...)
		if err != nil {
			slog.Warn("failed to store log batch", "error", err)
		}
	}

	return nil

}
