package store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"noname/internal/jsonvalue"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

const (
	spanBatchSize        = 2000
	logBatchSize         = 3000
	linkBatchSize        = 6000
	deleteBatchSize      = 10000
	dashboardRecentLimit = 5
	dashboardErrorLevel  = 17
)

type SQLiteStore struct {
	readDB  *sql.DB
	writeDB *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve db path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	slog.Info("Opening SQLite database", "path", absPath)

	dsn := absPath + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)"

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

func (s *SQLiteStore) ResolveResource(ctx context.Context, attributes jsonvalue.Attributes) (Resource, error) {
	hash, raw, err := hashAttributes(attributes)
	if err != nil {
		return Resource{}, err
	}

	result, err := s.writeDB.ExecContext(ctx, `INSERT INTO resources (hash, attributes) VALUES (?, ?) ON CONFLICT(hash) DO NOTHING`, hash, string(raw))
	if err != nil {
		return Resource{}, err
	}

	var id int64
	affected, err := result.RowsAffected()
	if err != nil {
		return Resource{}, err
	}

	if affected == 1 {
		id, err = result.LastInsertId()
		if err != nil {
			return Resource{}, err
		}
	} else {
		if err := s.readDB.QueryRowContext(ctx, `SELECT id FROM resources WHERE hash = ?`, hash).Scan(&id); err != nil {
			return Resource{}, err
		}
	}

	return Resource{
		ID:         id,
		Attributes: attributes,
	}, nil
}

func hashAttributes(attributes jsonvalue.Attributes) (hash string, raw []byte, err error) {
	raw, err = json.Marshal(attributes)
	if err != nil {
		return "", nil, err
	}

	sum := sha256.Sum256(raw)

	return hex.EncodeToString(sum[:]), raw, nil
}

const resourceLookupBatchSize = 500

func (s *SQLiteStore) loadResources(ctx context.Context, tx *sql.Tx, ids []int64) (map[int64]jsonvalue.Attributes, error) {
	resources := make(map[int64]jsonvalue.Attributes, len(ids))
	if len(ids) == 0 {
		return resources, nil
	}

	uniqueIDs := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}

	for left := 0; left < len(uniqueIDs); left += resourceLookupBatchSize {
		right := min(left+resourceLookupBatchSize, len(uniqueIDs))
		if err := s.loadResourceBatch(ctx, tx, uniqueIDs[left:right], resources); err != nil {
			return nil, err
		}
	}

	return resources, nil
}

func (s *SQLiteStore) loadResourceBatch(ctx context.Context, tx *sql.Tx, ids []int64, resources map[int64]jsonvalue.Attributes) (err error) {
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}

	query := `SELECT id, attributes FROM resources WHERE id IN ` + inPlaceholders(len(ids))

	var rows *sql.Rows
	if tx != nil {
		rows, err = tx.QueryContext(ctx, query, args...)
	} else {
		rows, err = s.readDB.QueryContext(ctx, query, args...)
	}
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := rows.Close(); err == nil {
			err = closeErr
		}
	}()

	for rows.Next() {
		var id int64
		var raw []byte
		if err := rows.Scan(&id, &raw); err != nil {
			return err
		}

		var attributes jsonvalue.Attributes
		if err := decodeJson(raw, &attributes); err != nil {
			return err
		}
		resources[id] = attributes
	}

	return rows.Err()
}

func attachResourceAttributes(logs []Log, resources map[int64]jsonvalue.Attributes) {
	for i := range logs {
		logs[i].ResourceAttributes = resources[logs[i].ResourceID]
	}
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

	if err := s.bulkInsertLinks(ctx, tx, spans); err != nil {
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

		query := `INSERT OR IGNORE INTO spans (span_id, trace_id, parent_span_id, service, environment, name, kind, start_time_ns, end_time_ns, status_code, status_message, attributes, events, resource_id) VALUES`

		// underlying array size is set to the number of spans in the batch * the number of fields in each span (currently 14)
		values := make([]any, 0, len(spanBatch)*14)

		for _, span := range spanBatch {
			attributes, err := json.Marshal(span.Attributes)
			if err != nil {
				return fmt.Errorf("marshal span attributes: %w", err)
			}

			events, err := json.Marshal(span.Events)
			if err != nil {
				return fmt.Errorf("marshal span events: %w", err)
			}

			placeholders = append(placeholders, "(?,?,?,?,?,?,?,?,?,?,?,?,?,?)")

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
				span.ResourceID,
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

func (s *SQLiteStore) bulkInsertLinks(ctx context.Context, tx *sql.Tx, spans []Span) error {

	type linkRow struct {
		traceID string
		spanID  string
		link    Link
	}

	rows := make([]linkRow, 0)

	for _, span := range spans {
		for _, link := range span.Links {
			rows = append(rows, linkRow{traceID: span.TraceID, spanID: span.SpanID, link: link})
		}
	}

	if len(rows) == 0 {
		return nil
	}

	for left := 0; left < len(rows); left += linkBatchSize {

		right := left + linkBatchSize

		if right > len(rows) {
			right = len(rows)
		}

		batch := rows[left:right]

		query := `INSERT OR IGNORE INTO span_links (trace_id, span_id, linked_trace_id, linked_span_id, attributes) VALUES`

		placeholders := make([]string, 0, len(batch))
		values := make([]any, 0, len(batch)*5)

		for _, row := range batch {
			attributes, err := json.Marshal(row.link.Attributes)
			if err != nil {
				return fmt.Errorf("marshal span link attributes: %w", err)
			}

			placeholders = append(placeholders, "(?,?,?,?,?)")

			values = append(values,
				row.traceID,
				row.spanID,
				row.link.TraceID,
				row.link.SpanID,
				attributes,
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

		query := `INSERT INTO logs (trace_id, span_id, timestamp_ns, severity_text, severity_number, body, service, environment, attributes, resource_id) VALUES`

		values := make([]any, 0, len(batch)*10) //size set to number of fields inserting into * lenght of batch

		placeholders := make([]string, 0, len(batch))

		for _, log := range batch {
			attributes, err := json.Marshal(log.Attributes)
			if err != nil {
				return fmt.Errorf("marshal log attributes: %w", err)
			}

			body, err := json.Marshal(log.Body)
			if err != nil {
				return fmt.Errorf("marshal log body: %w", err)
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
				log.ResourceID)

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
		conds = append(conds, "start_time_ns < ?")
		args = append(args, options.End.UnixNano())
	}
	if options.MinDurationMs != nil {
		conds = append(conds, "end_time_ns - start_time_ns >= ?")
		args = append(args, *options.MinDurationMs*1_000_000)
	}
	if options.MaxDurationMs != nil {
		conds = append(conds, "end_time_ns - start_time_ns < ?")
		args = append(args, *options.MaxDurationMs*1_000_000)
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

	query := `SELECT id, trace_id, span_id, timestamp_ns, severity_text, severity_number, body, service,environment, attributes, resource_id FROM logs`
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
		var attributes, body []byte

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
			&l.ResourceID,
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

		l.Timestamp = time.Unix(0, timeStampNs)
		logs = append(logs, l)

	}

	if err := rows.Err(); err != nil {
		return page, err
	}

	if err := rows.Close(); err != nil {
		return page, err
	}

	resourceIDs := make([]int64, 0, len(logs))
	for _, l := range logs {
		resourceIDs = append(resourceIDs, l.ResourceID)
	}

	resourceAttributes, err := s.loadResources(ctx, nil, resourceIDs)
	if err != nil {
		return page, err
	}

	attachResourceAttributes(logs, resourceAttributes)

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

	query := `SELECT span_id, trace_id, parent_span_id, service, environment, name, kind, start_time_ns, end_time_ns, status_code, status_message, attributes, events, resource_id FROM spans`
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
			&s.ResourceID,
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

func (s *SQLiteStore) ListSpanLinks(ctx context.Context, traceID, spanID string) ([]SpanLink, error) {

	query := `SELECT sl.linked_trace_id, sl.linked_span_id, sl.attributes, s.name, s.service, s.status_code FROM span_links sl LEFT JOIN spans s ON s.trace_id = sl.linked_trace_id AND s.span_id = sl.linked_span_id WHERE sl.trace_id = ? AND sl.span_id = ? ORDER BY sl.linked_trace_id, sl.linked_span_id`

	rows, err := s.readDB.QueryContext(ctx, query, traceID, spanID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]SpanLink, 0)

	for rows.Next() {
		var link SpanLink
		var attributes []byte
		var name, service sql.NullString
		var statusCode sql.NullInt64

		if err := rows.Scan(&link.TraceID, &link.SpanID, &attributes, &name, &service, &statusCode); err != nil {
			return nil, err
		}

		if err := decodeJson(attributes, &link.Attributes); err != nil {
			return nil, err
		}

		if name.Valid {
			link.Target = &SpanLinkTarget{
				Name:       name.String,
				Service:    service.String,
				StatusCode: toStatusCode(int32(statusCode.Int64)),
			}
		}

		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return links, nil
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

func (s *SQLiteStore) GetServiceGraph(ctx context.Context, start, end *time.Time) (ServiceGraph, error) {
	graph := ServiceGraph{Nodes: []ServiceNode{}, Edges: []ServiceEdge{}}
	nodeIDs := make(map[string]struct{})

	conds := make([]string, 0, 3)
	args := make([]any, 0, 3)
	if start != nil {
		conds = append(conds, "start_time_ns >= ?")
		args = append(args, start.UnixNano())
	}
	if end != nil {
		conds = append(conds, "start_time_ns <= ?")
		args = append(args, end.UnixNano())
	}
	conds = append(conds, "service != ''")
	where := " WHERE " + strings.Join(conds, " AND ")

	rows, err := s.readDB.QueryContext(ctx, `SELECT DISTINCT service FROM spans`+where+` ORDER BY service`, args...)
	if err != nil {
		return graph, err
	}
	defer rows.Close()

	for rows.Next() {
		var service string
		if err := rows.Scan(&service); err != nil {
			return graph, err
		}
		nodeIDs[service] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return graph, err
	}

	edgeConds := make([]string, 0, 5)
	edgeArgs := make([]any, 0, 5)
	edgeConds = append(edgeConds, "child.service != parent.service")
	edgeConds = append(edgeConds, "child.service != ''")
	edgeConds = append(edgeConds, "parent.service != ''")
	if start != nil {
		edgeConds = append(edgeConds, "child.start_time_ns >= ?")
		edgeArgs = append(edgeArgs, start.UnixNano())
	}
	if end != nil {
		edgeConds = append(edgeConds, "child.start_time_ns <= ?")
		edgeArgs = append(edgeArgs, end.UnixNano())
	}

	edgeQuery := `SELECT DISTINCT parent.service, child.service FROM spans child JOIN spans parent ON child.parent_span_id = parent.span_id AND child.trace_id = parent.trace_id WHERE ` + strings.Join(edgeConds, " AND ") + ` ORDER BY parent.service, child.service`
	edgeRows, err := s.readDB.QueryContext(ctx, edgeQuery, edgeArgs...)
	if err != nil {
		return graph, err
	}
	defer edgeRows.Close()

	for edgeRows.Next() {
		var edge ServiceEdge
		if err := edgeRows.Scan(&edge.From, &edge.To); err != nil {
			return graph, err
		}
		nodeIDs[edge.From] = struct{}{}
		nodeIDs[edge.To] = struct{}{}
		graph.Edges = append(graph.Edges, edge)
	}
	if err := edgeRows.Err(); err != nil {
		return graph, err
	}

	serviceIDs := make([]string, 0, len(nodeIDs))
	for service := range nodeIDs {
		serviceIDs = append(serviceIDs, service)
	}
	sort.Strings(serviceIDs)
	for _, service := range serviceIDs {
		graph.Nodes = append(graph.Nodes, ServiceNode{ID: service})
	}

	return graph, nil
}

func (s *SQLiteStore) GetDashboardStats(ctx context.Context, start, end time.Time) (DashboardStats, error) {
	stats := DashboardStats{
		TraceBuckets:      newTraceBuckets(start, end),
		RecentErrorTraces: []Trace{},
		RecentErrorLogs:   []Log{},
	}

	tx, err := s.readDB.BeginTx(ctx, nil)
	if err != nil {
		return stats, err
	}
	defer tx.Rollback()

	startNS, endNS := start.UnixNano(), end.UnixNano()
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN error THEN 1 ELSE 0 END), 0)
		FROM traces
		WHERE start_time_ns >= ? AND start_time_ns < ?`, startNS, endNS,
	).Scan(&stats.Overview.TraceCount, &stats.Overview.ErrorTraceCount)
	if err != nil {
		return stats, err
	}

	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT service)
		FROM spans
		WHERE start_time_ns >= ? AND start_time_ns < ? AND service != ''`, startNS, endNS,
	).Scan(&stats.Overview.ActiveServiceCount)
	if err != nil {
		return stats, err
	}

	bucketSizeNS := dashboardTraceBucketSize(end.Sub(start)).Nanoseconds()
	firstBucketIndex := startNS / bucketSizeNS
	bucketRows, err := tx.QueryContext(ctx, `
		SELECT start_time_ns / ? AS bucket_index, error, COUNT(*)
		FROM traces
		WHERE start_time_ns >= ? AND start_time_ns < ?
		GROUP BY bucket_index, error`, bucketSizeNS, startNS, endNS,
	)
	if err != nil {
		return stats, err
	}
	for bucketRows.Next() {
		var bucketIndex int64
		var bucketError bool
		var count int64
		if err := bucketRows.Scan(&bucketIndex, &bucketError, &count); err != nil {
			bucketRows.Close()
			return stats, err
		}
		bucket := &stats.TraceBuckets[int(bucketIndex-firstBucketIndex)]
		if bucketError {
			bucket.ErrorCount = count
		} else {
			bucket.OkCount = count
		}
	}
	if err := bucketRows.Err(); err != nil {
		bucketRows.Close()
		return stats, err
	}
	if err := bucketRows.Close(); err != nil {
		return stats, err
	}

	traceRows, err := tx.QueryContext(ctx, `
		SELECT trace_id, root_service, environment, root_name, start_time_ns, end_time_ns,
			span_count, error, updated_at_ns
		FROM traces
		WHERE error = 1 AND start_time_ns >= ? AND start_time_ns < ?
		ORDER BY start_time_ns DESC, trace_id DESC
		LIMIT ?`, startNS, endNS, dashboardRecentLimit,
	)
	if err != nil {
		return stats, err
	}
	for traceRows.Next() {
		var trace Trace
		var startTimeNS, endTimeNS, updatedAtNS int64
		if err := traceRows.Scan(
			&trace.TraceID,
			&trace.RootService,
			&trace.Environment,
			&trace.RootName,
			&startTimeNS,
			&endTimeNS,
			&trace.SpanCount,
			&trace.Error,
			&updatedAtNS,
		); err != nil {
			traceRows.Close()
			return stats, err
		}
		trace.StartTime = time.Unix(0, startTimeNS)
		trace.EndTime = time.Unix(0, endTimeNS)
		trace.UpdatedAt = time.Unix(0, updatedAtNS)
		stats.RecentErrorTraces = append(stats.RecentErrorTraces, trace)
	}
	if err := traceRows.Err(); err != nil {
		traceRows.Close()
		return stats, err
	}
	if err := traceRows.Close(); err != nil {
		return stats, err
	}

	logRows, err := tx.QueryContext(ctx, `
		SELECT id, trace_id, span_id, timestamp_ns, severity_text, severity_number,
			body, service, environment, attributes, resource_id
		FROM logs
		WHERE severity_number >= ? AND timestamp_ns >= ? AND timestamp_ns < ?
		ORDER BY timestamp_ns DESC, id DESC
		LIMIT ?`, dashboardErrorLevel, startNS, endNS, dashboardRecentLimit,
	)
	if err != nil {
		return stats, err
	}
	logIDs := make([]int64, 0, dashboardRecentLimit)
	for logRows.Next() {
		var log Log
		var timestampNS int64
		var body, attributes []byte
		if err := logRows.Scan(
			&log.ID,
			&log.TraceID,
			&log.SpanID,
			&timestampNS,
			&log.SeverityText,
			&log.SeverityNumber,
			&body,
			&log.Service,
			&log.Environment,
			&attributes,
			&log.ResourceID,
		); err != nil {
			logRows.Close()
			return stats, err
		}
		if err := decodeJson(body, &log.Body); err != nil {
			logRows.Close()
			return stats, err
		}
		if err := decodeJson(attributes, &log.Attributes); err != nil {
			logRows.Close()
			return stats, err
		}
		log.Timestamp = time.Unix(0, timestampNS)
		logIDs = append(logIDs, log.ID)
		stats.RecentErrorLogs = append(stats.RecentErrorLogs, log)
	}
	if err := logRows.Err(); err != nil {
		logRows.Close()
		return stats, err
	}
	if err := logRows.Close(); err != nil {
		return stats, err
	}

	resourceAttributes, err := s.loadResources(ctx, tx, logIDs)
	if err != nil {
		return stats, err
	}
	attachResourceAttributes(stats.RecentErrorLogs, resourceAttributes)

	if err := tx.Commit(); err != nil {
		return stats, err
	}
	return stats, nil
}

const dashboardTraceBucketLimit = 100

var dashboardTraceBucketSizes = [...]time.Duration{
	15 * time.Second,
	30 * time.Second,
	time.Minute,
	5 * time.Minute,
	15 * time.Minute,
	30 * time.Minute,
	time.Hour,
	2 * time.Hour,
	6 * time.Hour,
	12 * time.Hour,
	24 * time.Hour,
}

func dashboardTraceBucketSize(window time.Duration) time.Duration {
	for _, size := range dashboardTraceBucketSizes {
		if window <= size*dashboardTraceBucketLimit {
			return size
		}
	}
	return dashboardTraceBucketSizes[len(dashboardTraceBucketSizes)-1]
}

func newTraceBuckets(start, end time.Time) []TraceTimeBucket {
	if !end.After(start) {
		return []TraceTimeBucket{}
	}

	startNS, endNS := start.UnixNano(), end.UnixNano()
	bucketSizeNS := dashboardTraceBucketSize(end.Sub(start)).Nanoseconds()
	firstBucketIndex := startNS / bucketSizeNS
	lastBucketIndex := (endNS - 1) / bucketSizeNS
	buckets := make([]TraceTimeBucket, lastBucketIndex-firstBucketIndex+1)

	for bucketIndex := firstBucketIndex; bucketIndex <= lastBucketIndex; bucketIndex++ {
		bucketStartNS := bucketIndex * bucketSizeNS
		bucketEndNS := bucketStartNS + bucketSizeNS
		if bucketStartNS < startNS {
			bucketStartNS = startNS
		}
		if bucketEndNS > endNS {
			bucketEndNS = endNS
		}
		buckets[bucketIndex-firstBucketIndex] = TraceTimeBucket{
			StartTime: time.Unix(0, bucketStartNS).UTC(),
			EndTime:   time.Unix(0, bucketEndNS).UTC(),
		}
	}

	return buckets
}

func (s *SQLiteStore) GetTrace(ctx context.Context, traceID string) (*Trace, []Span, []Resource, error) {

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
		return nil, nil, nil, err
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
		return nil, nil, nil, err
	}

	resourceIDs := make([]int64, 0, len(page.Items))
	for _, span := range page.Items {
		resourceIDs = append(resourceIDs, span.ResourceID)
	}

	attributesByID, err := s.loadResources(ctx, nil, resourceIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	resources := make([]Resource, 0, len(attributesByID))
	for id, attributes := range attributesByID {
		resources = append(resources, Resource{
			ID:         id,
			Attributes: attributes,
		})
	}
	sort.Slice(resources, func(i, j int) bool { return resources[i].ID < resources[j].ID })

	return &t, page.Items, resources, nil

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

	if len(traceIDs) == 0 {
		return tx.Commit()
	}

	for left := 0; left < len(traceIDs); left += deleteBatchSize {
		right := left + deleteBatchSize
		if right > len(traceIDs) {
			right = len(traceIDs)
		}
		chunk := traceIDs[left:right]

		placeholders := strings.Repeat("?,", len(chunk))
		placeholders = strings.TrimSuffix(placeholders, ",")

		logQuery := fmt.Sprintf("DELETE FROM logs WHERE trace_id IN (%s) OR timestamp_ns <= (%v)", placeholders, cutoffNS)
		linkQuery := fmt.Sprintf("DELETE FROM span_links WHERE trace_id IN (%s)", placeholders)
		spanQuery := fmt.Sprintf("DELETE FROM spans WHERE trace_id IN (%s)", placeholders)
		traceQuery := fmt.Sprintf("DELETE FROM traces WHERE trace_id IN (%s)", placeholders)

		args := make([]any, 0, len(chunk))
		for _, id := range chunk {
			args = append(args, id)
		}

		if _, err := tx.ExecContext(ctx, logQuery, args...); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, linkQuery, args...); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, spanQuery, args...); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, traceQuery, args...); err != nil {
			return err
		}
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
