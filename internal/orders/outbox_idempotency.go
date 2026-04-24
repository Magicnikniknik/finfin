package orders

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type OutboxEvent struct {
	TenantID      string
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       map[string]any
}

func insertOutboxEvent(ctx context.Context, tx pgx.Tx, evt OutboxEvent) error {
	payloadJSON, err := json.Marshal(evt.Payload)
	if err != nil {
		return fmt.Errorf("marshal outbox payload: %w", err)
	}

	_, err = tx.Exec(ctx, `
INSERT INTO core.outbox_events (
	tenant_id,
	aggregate_type,
	aggregate_id,
	event_type,
	payload,
	headers,
	status,
	available_at,
	created_at
)
VALUES (
	$1::uuid,
	$2,
	$3::uuid,
	$4,
	$5::jsonb,
	'{}'::jsonb,
	'pending',
	now(),
	now()
)
`,
		evt.TenantID,
		evt.AggregateType,
		evt.AggregateID,
		evt.EventType,
		payloadJSON,
	)
	if err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}

	return nil
}

type idempotencyRecord struct {
	Status       string
	RequestHash  string
	ResponseBody []byte
	ResourceID   *string
}

func buildRequestHash(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		_, _ = h.Write([]byte(p))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func acquireIdempotency(
	ctx context.Context,
	tx pgx.Tx,
	tenantID string,
	scope string,
	key string,
	requestHash string,
) (*idempotencyRecord, error) {
	tag, err := tx.Exec(ctx, `
INSERT INTO core.idempotency_keys (
	tenant_id,
	scope,
	idem_key,
	request_hash,
	status,
	locked_until,
	created_at,
	updated_at
)
VALUES (
	$1::uuid,
	$2,
	$3,
	$4,
	'in_progress',
	now() + interval '30 seconds',
	now(),
	now()
)
ON CONFLICT (tenant_id, scope, idem_key) DO NOTHING
`,
		tenantID,
		scope,
		key,
		requestHash,
	)
	if err != nil {
		return nil, fmt.Errorf("insert idempotency key: %w", err)
	}

	if tag.RowsAffected() == 1 {
		return nil, nil
	}

	var rec idempotencyRecord
	err = tx.QueryRow(ctx, `
SELECT
	status,
	request_hash,
	response_body,
	resource_id::text
FROM core.idempotency_keys
WHERE tenant_id = $1::uuid
  AND scope = $2
  AND idem_key = $3
`,
		tenantID,
		scope,
		key,
	).Scan(
		&rec.Status,
		&rec.RequestHash,
		&rec.ResponseBody,
		&rec.ResourceID,
	)
	if err != nil {
		return nil, fmt.Errorf("load conflicting idempotency key: %w", err)
	}

	if rec.RequestHash != requestHash {
		return nil, ErrIdempotencyConflict
	}
	if rec.Status == "completed" {
		return &rec, nil
	}
	return nil, ErrIdempotencyConflict
}

func finalizeIdempotency(
	ctx context.Context,
	tx pgx.Tx,
	tenantID string,
	scope string,
	key string,
	resourceType string,
	resourceID string,
	response any,
) error {
	body, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal idempotency response: %w", err)
	}

	_, err = tx.Exec(ctx, `
UPDATE core.idempotency_keys
SET
	status = 'completed',
	response_code = 200,
	response_body = $5::jsonb,
	resource_type = $4,
	resource_id = $6::uuid,
	updated_at = now()
WHERE tenant_id = $1::uuid
  AND scope = $2
  AND idem_key = $3
`,
		tenantID,
		scope,
		key,
		resourceType,
		body,
		resourceID,
	)
	if err != nil {
		return fmt.Errorf("finalize idempotency key: %w", err)
	}

	return nil
}

func decodeCachedResponse[T any](rec *idempotencyRecord) (T, error) {
	var out T
	if rec == nil {
		return out, errors.New("nil idempotency record")
	}
	if len(rec.ResponseBody) == 0 {
		return out, errors.New("empty cached response")
	}
	if err := json.Unmarshal(rec.ResponseBody, &out); err != nil {
		return out, fmt.Errorf("unmarshal cached response: %w", err)
	}
	return out, nil
}
