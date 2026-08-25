package store

import (
	"bytes"
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
				resourceAttributes = []byte("{}")
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
				string(body),
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

func (s *SQLiteStore) ListTraces(ctx context.Context, options ListTracesOptions) (Page[Trace, ListTracesCursor], error) {

	var startTimeNS *int64
	var traceID *string

	page := Page[Trace, ListTracesCursor]{}

	limit := options.Limit
	if limit <= 0 {
		limit = 20
	}
	if options.Cursor != nil {
		ns := options.Cursor.StartTime.UnixNano()
		startTimeNS = &ns
		traceID = &options.Cursor.TraceID
	}

	query := `SELECT trace_id, root_service, environment, root_name, start_time_ns, end_time_ns, span_count, error, updated_at_ns FROM
	 traces WHERE (?1 IS NULL) OR (start_time_ns,trace_id) < (?1, ?2) ORDER BY start_time_ns DESC, trace_id DESC  LIMIT (?3 + 1) `

	rows, err := s.readDB.QueryContext(ctx, query,
		startTimeNS,
		traceID,
		limit)

	if err != nil {
		return page, err
	}
	defer rows.Close()

	traces := make([]Trace, 0, limit+1)

	for rows.Next() {
		var startTimeNS, endTimeNS, updatedNS int64
		var t Trace

		err := rows.Scan(
			&t.TraceID,
			&t.RootService,
			&t.Environment,
			&t.RootName,
			&startTimeNS,
			&endTimeNS,
			&t.SpanCount,
			&t.Error,
			&updatedNS,
		)

		if err != nil {
			return page, err
		}

		t.StartTime = time.Unix(0, startTimeNS)
		t.EndTime = time.Unix(0, endTimeNS)
		t.UpdatedAt = time.Unix(0, updatedNS)

		traces = append(traces, t)

	}

	if err := rows.Err(); err != nil {
		return page, err
	}

	items, hasMore, extra := paginate(traces, limit)

	page.HasMore = hasMore
	page.Items = items

	if extra != nil {
		page.NextCursor = &ListTracesCursor{
			StartTime: extra.StartTime,
			TraceID:   extra.TraceID,
		}
	}

	return page, nil

}

func (s *SQLiteStore) ListLogs(ctx context.Context, options ListLogsOptions) (Page[Log, ListLogsCursor], error) {

	var timeStampNs *int64
	var id *int64

	page := Page[Log, ListLogsCursor]{}

	limit := options.Limit

	if limit <= 0 {
		limit = 50
	}

	if options.Cursor != nil {
		id = &options.Cursor.ID
		ns := options.Cursor.TimeStamp.UnixNano()
		timeStampNs = &ns
	}

	query := `SELECT id, trace_id, span_id, timestamp_ns, severity_text, severity_number, body, service,environment, attributes, resource_attributes FROM logs WHERE (?1 IS NULL) OR (timestamp_ns,id) < (?1,?2) ORDER BY timestamp_ns DESC, id DESC LIMIT (?3 + 1)`

	rows, err := s.readDB.QueryContext(ctx, query,
		timeStampNs,
		id,
		limit)

	if err != nil {
		return page, err
	}
	defer rows.Close()

	logs := make([]Log, 0, limit+1)

	for rows.Next() {
		var timeStampNs int64
		var attributes, resourceAttributes, body []byte

		var l Log

		err := rows.Scan(
			&l.ID,
			&l.TraceID,
			&l.SpanID,
			&timeStampNs,
			&l.SeverityText,
			&l.SeverityNumber,
			&body,
			&l.Service,
			&l.Environment,
			&attributes,
			&resourceAttributes,
		)

		if err != nil {
			return page, err
		}

		if err := decodeJson(body, &l.Body); err != nil {
			return page, err
		}

		if err := decodeJson(attributes, &l.Attributes); err != nil {
			return page, err
		}

		if err := decodeJson(resourceAttributes, &l.ResourceAttributes); err != nil {
			return page, err
		}

		l.Timestamp = time.Unix(0, timeStampNs)
		logs = append(logs, l)

	}

	if err := rows.Err(); err != nil {
		return page, err
	}

	items, hasMore, extra := paginate(logs, limit)

	page.HasMore = hasMore
	page.Items = items

	if extra != nil {
		page.NextCursor = &ListLogsCursor{
			ID:        extra.ID,
			TimeStamp: extra.Timestamp,
		}

	}
	return page, nil
}

func (s *SQLiteStore) ListSpans(ctx context.Context, options ListSpansOptions) (Page[Span, ListSpansCursor], error) {

	var startTimeNS *int64
	var spanID *string
	var traceID string

	page := Page[Span, ListSpansCursor]{}

	limit := options.Limit

	if limit <= 0 && limit != -1 {
		limit = 50
	}

	if options.Cursor != nil {
		spanID = &options.Cursor.SpanID
		ns := options.Cursor.StartTime.UnixNano()
		startTimeNS = &ns
	}

	if options.TraceID != "" {
		traceID = options.TraceID
	}

	query := `SELECT span_id, trace_id, parent_span_id, service,name, kind, start_time_ns, end_time_ns, status_code, status_message, attributes, events
	FROM spans
	WHERE ((?1 IS NULL) OR (start_time_ns,span_id) < (?1,?2)) AND (?4 = 
	'' OR trace_id = ?4) ORDER BY start_time_ns DESC , span_id DESC LIMIT (?3)`

	rows, err := s.readDB.QueryContext(ctx, query,
		startTimeNS,
		spanID,
		limit,
		traceID)

	if err != nil {
		return page, err
	}

	defer rows.Close()

	spans := make([]Span, 0, limit+1)

	for rows.Next() {

		var startTimeNS, endTimeNS int64
		var attributes, events []byte
		var statusCode, kind int32

		var s Span

		err := rows.Scan(
			&s.SpanID,
			&s.TraceID,
			&s.ParentSpanID,
			&s.Service,
			&s.Name,
			&kind,
			&startTimeNS,
			&endTimeNS,
			&statusCode,
			&s.StatusMessage,
			&attributes,
			&events,
		)

		if err != nil {
			return page, err
		}
		s.StartTime = time.Unix(0, startTimeNS)
		s.EndTime = time.Unix(0, endTimeNS)
		s.Kind = toSpanKind(kind)
		s.StatusCode = toStatusCode(statusCode)

		if err := decodeJson(attributes, &s.Attributes); err != nil {
			return page, err
		}

		if err := decodeJson(events, &s.Events); err != nil {
			return page, err
		}

		spans = append(spans, s)

	}

	if err := rows.Err(); err != nil {
		return page, err
	}

	items, hasMore, extra := paginate(spans, limit)
	page.HasMore = hasMore
	page.Items = items

	if extra != nil {
		page.NextCursor = &ListSpansCursor{
			StartTime: extra.StartTime,
			SpanID:    extra.SpanID,
		}
	}
	return page, nil
}

func (s *SQLiteStore) GetTrace(ctx context.Context, traceID string) (*Trace, []Span, error) {

	query := `SELECT trace_id, root_service, environment, root_name, start_time_ns, end_time_ns, span_count, error, updated_at_ns FROM
	 traces WHERE trace_id = ?1`

	var t Trace
	var startTimeNS, endTimeNS, updatedNS int64

	err := s.readDB.QueryRowContext(ctx, query, traceID).Scan(&t.TraceID,
		&t.RootService,
		&t.Environment,
		&t.RootName,
		&startTimeNS,
		&endTimeNS,
		&t.SpanCount,
		&t.Error,
		&updatedNS)

	if err != nil {
		return nil, nil, err
	}
	t.StartTime = time.Unix(0, startTimeNS)
	t.EndTime = time.Unix(0, endTimeNS)
	t.UpdatedAt = time.Unix(0, updatedNS)

	listSpanOptions := ListSpansOptions{
		ListOptions: ListOptions[ListSpansCursor]{
			Limit: -1, //sqlite treats negative as no limit
		},
		TraceID: traceID,
	}

	page, err := s.ListSpans(ctx, listSpanOptions)

	if err != nil {
		return nil, nil, err
	}

	return &t, page.Items, nil

}

func decodeJson[T any](b []byte, target *T) error {

	if len(b) == 0 {
		return nil
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	return dec.Decode(&target)

}
