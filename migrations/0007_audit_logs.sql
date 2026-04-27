BEGIN;

CREATE SCHEMA IF NOT EXISTS audit;

CREATE TABLE IF NOT EXISTS audit.audit_logs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid,
  office_id uuid,
  actor_user_id uuid,
  actor_role text,
  action text NOT NULL,
  entity_type text,
  entity_id text,
  ip inet,
  request_id text,
  payload_snapshot jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_audit_logs_tenant_created_at
  ON audit.audit_logs (tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_audit_logs_actor_created_at
  ON audit.audit_logs (actor_user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_audit_logs_action_created_at
  ON audit.audit_logs (action, created_at DESC);

COMMIT;
