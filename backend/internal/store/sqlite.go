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

const (
	spanBatchSize = 2000
	logBatchSize  = 3000
)

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

	query := `INSERT INTO traces (trace_id, root_service, environment, root_name, start_time_ns, end_time_ns, span_count, error, updated_at_ns) VALUES (?,?,?,?,?,?,?,?,?) ON CONFLICT(trace_id) DO UPDATE SET root_service = CASE WHEN root_service = '' THEN excluded.root_service ELSE root_service END, environment = CASE WHEN environment IS NULL OR environment = '' THEN excluded.environment ELSE environment END, root_name = CASE WHEN root_name = '' THEN excluded.root_name ELSE root_name END, start_time_ns = MIN(start_time_ns, excluded.start_time_ns), end_time_ns = MAX(end_time_ns, excluded.end_time_ns), span_count = span_count + excluded.span_count, error = error OR excluded.error, updated_at_ns = excluded.updated_at_ns`

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

	for left := 0; left < len(spans); left += spanBatchSize {

		right := left + spanBatchSize

		if right > len(spans) {
			right = len(spans)
		}

		spanBatch := spans[left:right]

		placeholders := make([]string, 0, len(spanBatch))

		query := `INSERT OR IGNORE INTO spans (span_id, trace_id, parent_span_id, service, environment, name, kind, start_time_ns, end_time_ns, status_code, status_message, attributes, events) VALUES`

		//underlying array size is set to the total number of spans * the number of fields in each span (currently 13)
		values := make([]any, 0, len(spanBatch)*13)

		for _, span := range spanBatch {
			attributes, err := json.Marshal(span.Attributes)
			if err != nil {
				attributes = []byte("{}")
			}

			events, err := json.Marshal(span.Events)
			if err != nil {
				events = []byte("[]")
			}

			placeholders = append(placeholders, "(?,?,?,?,?,?,?,?,?,?,?,?,?)")

			values = append(values,
				span.SpanID,
				span.TraceID,
				span.ParentSpanID,
				span.Service,
				span.Environment,
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

	for left := 0; left < len(logs); left += logBatchSize {

		right := left + logBatchSize

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
			return err
		}
	}

	return nil

}

func inPlaceholders(n int) string {
	var b strings.Builder
	b.WriteByte('(')
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('?')
	}
	b.WriteByte(')')
	return b.String()
}

func appendStrings(args []any, vals []string) []any {
	for _, v := range vals {
		args = append(args, v)
	}
	return args
}

func (s *SQLiteStore) ListTraces(ctx context.Context, options ListTracesOptions) (Page[Trace, ListTracesCursor], error) {

	page := Page[Trace, ListTracesCursor]{}

	limit, fetchLimit := resolvePagingLimits(options.Limit, 20)

	comparator := "<"
	orderDir := "DESC"
	if options.Ascending {
		comparator = ">"
		orderDir = "ASC"
	}

	conds := make([]string, 0, 8)
	args := make([]any, 0, 16)

	if options.Cursor != nil {
		conds = append(conds, "(start_time_ns,trace_id) "+comparator+" (?, ?)")
		args = append(args, options.Cursor.StartTime.UnixNano(), options.Cursor.TraceID)
	}
	if options.Search != "" {
		conds = append(conds, "(root_name LIKE ? ESCAPE '\\' COLLATE NOCASE OR trace_id LIKE ? ESCAPE '\\' COLLATE NOCASE)")
		likePattern := "%" + escapeLike(options.Search) + "%"
		args = append(args, likePattern, likePattern)
	}
	if options.ErrorFilter != nil {
		conds = append(conds, "error = ?")
		args = append(args, *options.ErrorFilter)
	}
	if len(options.Services) > 0 {
		conds = append(conds, "EXISTS (SELECT 1 FROM spans WHERE spans.trace_id = traces.trace_id AND service IN "+inPlaceholders(len(options.Services))+")")
		args = appendStrings(args, options.Services)
	}
	if options.Start != nil {
		conds = append(conds, "start_time_ns >= ?")
		args = append(args, options.Start.UnixNano())
	}
	if options.End != nil {
		conds = append(conds, "start_time_ns <= ?")
		args = append(args, options.End.UnixNano())
	}
	if len(options.Environments) > 0 {
		conds = append(conds, "environment IN "+inPlaceholders(len(options.Environments)))
		args = appendStrings(args, options.Environments)
	}

	query := `SELECT trace_id, root_service, environment, root_name, start_time_ns, end_time_ns, span_count, error, updated_at_ns FROM traces`
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += ` ORDER BY start_time_ns ` + orderDir + `, trace_id ` + orderDir + ` LIMIT (?)`
	args = append(args, fetchLimit)

	rows, err := s.readDB.QueryContext(ctx, query, args...)

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

	items, hasMore, last := paginate(traces, limit)

	page.HasMore = hasMore
	page.Items = items

	if last != nil {
		page.NextCursor = &ListTracesCursor{
			StartTime: last.StartTime,
			TraceID:   last.TraceID,
		}
	}

	return page, nil

}

func (s *SQLiteStore) ListLogs(ctx context.Context, options ListLogsOptions) (Page[Log, ListLogsCursor], error) {

	page := Page[Log, ListLogsCursor]{}

	limit, fetchLimit := resolvePagingLimits(options.Limit, 30)

	var levelMin, levelMax *int32
	if options.Level != nil {
		switch *options.Level {
		case LogLevelInfo:
			max := int32(12)
			levelMax = &max
		case LogLevelWarn:
			min, max := int32(13), int32(16)
			levelMin, levelMax = &min, &max
		case LogLevelError:
			min := int32(17)
			levelMin = &min
		}
	}

	conds := make([]string, 0, 10)
	args := make([]any, 0, 16)

	if options.Cursor != nil {
		conds = append(conds, "(timestamp_ns,id) < (?,?)")
		args = append(args, options.Cursor.TimeStamp.UnixNano(), options.Cursor.ID)
	}
	if options.Search != "" {
		conds = append(conds, "body LIKE ? ESCAPE '\\' COLLATE NOCASE")
		args = append(args, "%"+escapeLike(options.Search)+"%")
	}
	if levelMin != nil {
		conds = append(conds, "severity_number >= ?")
		args = append(args, *levelMin)
	}
	if levelMax != nil {
		conds = append(conds, "severity_number <= ?")
		args = append(args, *levelMax)
	}
	if len(options.Services) > 0 {
		conds = append(conds, "service IN "+inPlaceholders(len(options.Services)))
		args = appendStrings(args, options.Services)
	}
	if options.TraceID != "" {
		conds = append(conds, "trace_id = ?")
		args = append(args, options.TraceID)
	}
	if options.SpanID != "" {
		conds = append(conds, "span_id = ?")
		args = append(args, options.SpanID)
	}
	if options.Start != nil {
		conds = append(conds, "timestamp_ns >= ?")
		args = append(args, options.Start.UnixNano())
	}
	if options.End != nil {
		conds = append(conds, "timestamp_ns <= ?")
		args = append(args, options.End.UnixNano())
	}
	if len(options.Environments) > 0 {
		conds = append(conds, "environment IN "+inPlaceholders(len(options.Environments)))
		args = appendStrings(args, options.Environments)
	}

	query := `SELECT id, trace_id, span_id, timestamp_ns, severity_text, severity_number, body, service,environment, attributes, resource_attributes FROM logs`
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += ` ORDER BY timestamp_ns DESC, id DESC LIMIT (?)`
	args = append(args, fetchLimit)

	rows, err := s.readDB.QueryContext(ctx, query, args...)

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

	items, hasMore, last := paginate(logs, limit)

	page.HasMore = hasMore
	page.Items = items

	if last != nil {
		page.NextCursor = &ListLogsCursor{
			ID:        last.ID,
			TimeStamp: last.Timestamp,
		}

	}
	return page, nil
}

func (s *SQLiteStore) ListSpans(ctx context.Context, options ListSpansOptions) (Page[Span, ListSpansCursor], error) {

	page := Page[Span, ListSpansCursor]{}

	limit, fetchLimit := resolvePagingLimits(options.Limit, 50)

	var statusEquals *int32
	var statusNotEquals *int32
	if options.Status != nil {
		switch *options.Status {
		case StatusCodeOk:
			ne := int32(StatusCodeError)
			statusNotEquals = &ne
		case StatusCodeError:
			eq := int32(StatusCodeError)
			statusEquals = &eq
		default:
			eq := int32(*options.Status)
			statusEquals = &eq
		}
	}

	var kind *int32
	if options.Kind != nil {
		k := int32(*options.Kind)
		kind = &k
	}

	conds := make([]string, 0, 10)
	args := make([]any, 0, 18)

	if options.Cursor != nil {
		conds = append(conds, "(start_time_ns,span_id) < (?,?)")
		args = append(args, options.Cursor.StartTime.UnixNano(), options.Cursor.SpanID)
	}
	if options.TraceID != "" {
		conds = append(conds, "trace_id = ?")
		args = append(args, options.TraceID)
	}
	if options.Search != "" {
		conds = append(conds, "name LIKE ? ESCAPE '\\' COLLATE NOCASE")
		args = append(args, "%"+escapeLike(options.Search)+"%")
	}
	if statusEquals != nil {
		conds = append(conds, "status_code = ?")
		args = append(args, *statusEquals)
	}
	if statusNotEquals != nil {
		conds = append(conds, "status_code != ?")
		args = append(args, *statusNotEquals)
	}
	if kind != nil {
		conds = append(conds, "kind = ?")
		args = append(args, *kind)
	}
	if len(options.Services) > 0 {
		conds = append(conds, "service IN "+inPlaceholders(len(options.Services)))
		args = appendStrings(args, options.Services)
	}
	if options.Start != nil {
		conds = append(conds, "start_time_ns >= ?")
		args = append(args, options.Start.UnixNano())
	}
	if options.End != nil {
		conds = append(conds, "start_time_ns <= ?")
		args = append(args, options.End.UnixNano())
	}
	if len(options.Environments) > 0 {
		conds = append(conds, "environment IN "+inPlaceholders(len(options.Environments)))
		args = appendStrings(args, options.Environments)
	}

	query := `SELECT span_id, trace_id, parent_span_id, service, environment, name, kind, start_time_ns, end_time_ns, status_code, status_message, attributes, events FROM spans`
	if len(conds) > 0 {
		query += " WHERE " + strings.Join(conds, " AND ")
	}
	query += ` ORDER BY start_time_ns DESC, span_id DESC LIMIT (?)`
	args = append(args, fetchLimit)

	rows, err := s.readDB.QueryContext(ctx, query, args...)

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
			&s.Environment,
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
		s.Kind = ToSpanKind(kind)
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

	items, hasMore, last := paginate(spans, limit)
	page.HasMore = hasMore
	page.Items = items

	if last != nil {
		page.NextCursor = &ListSpansCursor{
			StartTime: last.StartTime,
			SpanID:    last.SpanID,
		}
	}
	return page, nil
}

func (s *SQLiteStore) ListServices(ctx context.Context) ([]string, error) {
	rows, err := s.readDB.QueryContext(ctx, `SELECT DISTINCT service FROM spans ORDER BY service`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := []string{}

	for rows.Next() {
		var service string
		if err := rows.Scan(&service); err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (s *SQLiteStore) ListEnvironments(ctx context.Context) ([]string, error) {
	rows, err := s.readDB.QueryContext(ctx, `SELECT DISTINCT environment FROM (SELECT environment FROM spans UNION SELECT environment FROM logs) WHERE environment IS NOT NULL AND environment != '' ORDER BY environment`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	environments := []string{}

	for rows.Next() {
		var environment string
		if err := rows.Scan(&environment); err != nil {
			return nil, err
		}
		environments = append(environments, environment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return environments, nil
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

func (s *SQLiteStore) DeleteOldData(ctx context.Context, cutoff time.Time) error {

	tx, err := s.writeDB.BeginTx(ctx, &sql.TxOptions{})

	if err != nil {
		return err
	}
	defer tx.Rollback()

	cutoffNS := cutoff.UnixNano()

	rows, err := tx.QueryContext(ctx, `SELECT trace_id from traces WHERE end_time_ns <= ?1`, cutoffNS)

	if err != nil {
		return err
	}

	defer rows.Close()

	traceIDs := []string{}

	for rows.Next() {
		var traceID string

		err := rows.Scan(&traceID)

		if err != nil {
			return err
		}

		traceIDs = append(traceIDs, traceID)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("found %v traces for deletion", len(traceIDs)))

	placeholders := strings.Repeat("?,", len(traceIDs))
	placeholders = strings.TrimSuffix(placeholders, ",")

	logQuery := fmt.Sprintf("DELETE FROM logs WHERE trace_id IN (%s) OR timestamp_ns <= (%v)", placeholders, cutoffNS)
	spanQuery := fmt.Sprintf("DELETE FROM spans WHERE trace_id IN (%s)", placeholders)
	traceQuery := fmt.Sprintf("DELETE FROM traces WHERE trace_id IN (%s)", placeholders)

	//convert []string to []any to match args type
	args := make([]any, 0, len(traceIDs))
	for _, id := range traceIDs {
		args = append(args, id)
	}

	if _, err := tx.ExecContext(ctx, logQuery, args...); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, spanQuery, args...); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, traceQuery, args...); err != nil {
		return err
	}

	return tx.Commit()
}

func decodeJson[T any](b []byte, target *T) error {

	if len(b) == 0 {
		return nil
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	return dec.Decode(&target)

}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func resolvePagingLimits(requested, fallback int) (limit, fetch int) {
	limit, fetch = requested, requested
	if limit <= 0 && limit != -1 {
		limit, fetch = fallback, fallback+1
	} else if limit >= 0 {
		fetch = limit + 1
	}
	return limit, fetch
}
