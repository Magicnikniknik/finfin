package outbox

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type fakePublisher struct {
	err       error
	callCount int
	lastTopic string
	lastKey   string
	lastBody  []byte
}

func (p *fakePublisher) Publish(_ context.Context, topic string, key string, payload []byte) error {
	p.callCount++
	p.lastTopic = topic
	p.lastKey = key
	p.lastBody = append([]byte(nil), payload...)
	return p.err
}

func TestOutboxPublisher_PublishSuccess(t *testing.T) {
	ctx := context.Background()
	pool := openOutboxTestDB(t)
	resetAndMigrateCoreSchema(t, ctx, pool)

	tenantID := uuid.NewString()
	aggregateID := uuid.NewString()

	eventID := seedPendingOutboxEvent(
		t,
		ctx,
		pool,
		tenantID,
		"order",
		aggregateID,
		"order_reserved",
		`{"order_id":"`+aggregateID+`","status":"reserved"}`,
	)

	pub := &fakePublisher{}
	worker := NewWorker(pool, pub, slog.Default(), WorkerConfig{
		TickInterval: time.Second,
		BatchSize:    10,
	})

	processed, err := worker.processOne(ctx)
	if err != nil {
		t.Fatalf("processOne returned error: %v", err)
	}
	if !processed {
		t.Fatal("expected processOne to process one event")
	}

	if pub.callCount != 1 {
		t.Fatalf("expected publisher to be called once, got %d", pub.callCount)
	}
	if pub.lastTopic != "order_reserved" {
		t.Fatalf("expected topic order_reserved, got %q", pub.lastTopic)
	}
	expectedKey := "order:" + aggregateID
	if pub.lastKey != expectedKey {
		t.Fatalf("expected key %q, got %q", expectedKey, pub.lastKey)
	}
	if string(pub.lastBody) != `{"order_id":"`+aggregateID+`","status":"reserved"}` {
		t.Fatalf("unexpected payload: %s", string(pub.lastBody))
	}

	var (
		status      string
		attempts    int
		lastError   *string
		publishedAt *time.Time
	)
	err = pool.QueryRow(ctx, `
SELECT status, attempts, last_error, published_at
FROM core.outbox_events
WHERE id = $1
`, eventID).Scan(&status, &attempts, &lastError, &publishedAt)
	if err != nil {
		t.Fatalf("load outbox row after publish success: %v", err)
	}

	if status != "published" {
		t.Fatalf("expected status published, got %q", status)
	}
	if attempts != 0 {
		t.Fatalf("expected attempts 0, got %d", attempts)
	}
	if lastError != nil {
		t.Fatalf("expected last_error NULL, got %q", *lastError)
	}
	if publishedAt == nil {
		t.Fatal("expected published_at to be non-NULL")
	}
}

func TestOutboxPublisher_PublishFailureRetry(t *testing.T) {
	ctx := context.Background()
	pool := openOutboxTestDB(t)
	resetAndMigrateCoreSchema(t, ctx, pool)

	tenantID := uuid.NewString()
	aggregateID := uuid.NewString()

	eventID := seedPendingOutboxEvent(
		t,
		ctx,
		pool,
		tenantID,
		"order",
		aggregateID,
		"order_completed",
		`{"order_id":"`+aggregateID+`","status":"completed"}`,
	)

	pub := &fakePublisher{err: errors.New("broker unavailable")}
	worker := NewWorker(pool, pub, slog.Default(), WorkerConfig{
		TickInterval: time.Second,
		BatchSize:    10,
	})

	start := time.Now().UTC()

	processed, err := worker.processOne(ctx)
	if err != nil {
		t.Fatalf("processOne returned error on publish failure path: %v", err)
	}
	if !processed {
		t.Fatal("expected processOne to process one event even on publish failure")
	}

	if pub.callCount != 1 {
		t.Fatalf("expected publisher to be called once, got %d", pub.callCount)
	}

	var (
		status      string
		attempts    int
		lastError   *string
		publishedAt *time.Time
		availableAt time.Time
	)
	err = pool.QueryRow(ctx, `
SELECT status, attempts, last_error, published_at, available_at
FROM core.outbox_events
WHERE id = $1
`, eventID).Scan(&status, &attempts, &lastError, &publishedAt, &availableAt)
	if err != nil {
		t.Fatalf("load outbox row after publish failure: %v", err)
	}

	if status != "pending" {
		t.Fatalf("expected status pending after retry scheduling, got %q", status)
	}
	if attempts != 1 {
		t.Fatalf("expected attempts 1, got %d", attempts)
	}
	if lastError == nil {
		t.Fatal("expected last_error to be set")
	}
	if *lastError != "broker unavailable" {
		t.Fatalf("expected last_error to be broker unavailable, got %q", *lastError)
	}
	if publishedAt != nil {
		t.Fatal("expected published_at to remain NULL on failure")
	}
	if !availableAt.After(start) {
		t.Fatalf("expected available_at to be moved forward, got %v (start %v)", availableAt, start)
	}
}

func openOutboxTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL_TEST")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("DATABASE_URL_TEST or DATABASE_URL is required for integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create pgx pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

func resetAndMigrateCoreSchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	if _, err := pool.Exec(ctx, `DROP SCHEMA IF EXISTS core CASCADE`); err != nil {
		t.Fatalf("drop core schema: %v", err)
	}

	migrationPath := mustFindMigrationPath(t, "0001_core_schema.sql")
	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration file: %v", err)
	}

	if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
}

func mustFindMigrationPath(t *testing.T, filename string) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	base := filepath.Dir(currentFile)
	path := filepath.Clean(filepath.Join(base, "..", "..", "migrations", filename))

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("migration file not found at %s: %v", path, err)
	}

	return path
}

func seedPendingOutboxEvent(
	t *testing.T,
	ctx context.Context,
	pool *pgxpool.Pool,
	tenantID string,
	aggregateType string,
	aggregateID string,
	eventType string,
	payload string,
) int64 {
	t.Helper()

	var id int64
	err := pool.QueryRow(ctx, `
INSERT INTO core.outbox_events (
	tenant_id,
	aggregate_type,
	aggregate_id,
	event_type,
	payload,
	headers,
	status,
	attempts,
	available_at,
	created_at,
	updated_at
)
VALUES (
	$1::uuid,
	$2,
	$3::uuid,
	$4,
	$5::jsonb,
	'{}'::jsonb,
	'pending',
	0,
	now(),
	now(),
	now()
)
RETURNING id
`,
		tenantID,
		aggregateType,
		aggregateID,
		eventType,
		payload,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed pending outbox event: %v", err)
	}

	return id
}
