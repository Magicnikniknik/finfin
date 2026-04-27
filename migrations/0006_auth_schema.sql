BEGIN;

CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL,
  office_id uuid,
  login text NOT NULL,
  password_hash text NOT NULL,
  role text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  display_name text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_auth_users_role CHECK (role IN ('owner', 'manager', 'cashier')),
  CONSTRAINT chk_auth_users_status CHECK (status IN ('active', 'disabled')),
  CONSTRAINT uq_auth_users_tenant_login UNIQUE (tenant_id, login)
);

CREATE TABLE IF NOT EXISTS auth.refresh_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  token_hash text NOT NULL,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,
  user_agent text,
  ip inet,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_auth_refresh_tokens_user_id
  ON auth.refresh_tokens (user_id);

CREATE INDEX IF NOT EXISTS ix_auth_refresh_tokens_token_hash
  ON auth.refresh_tokens (token_hash);

COMMIT;
