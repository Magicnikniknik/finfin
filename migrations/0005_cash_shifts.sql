BEGIN;

CREATE TABLE IF NOT EXISTS core.cash_shifts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id uuid NOT NULL,
  office_id uuid NOT NULL,
  cashier_id text NOT NULL,
  status text NOT NULL DEFAULT 'open',
  opened_at timestamptz NOT NULL DEFAULT now(),
  closed_at timestamptz,
  opened_by text,
  closed_by text,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT chk_cash_shifts_status CHECK (status IN ('open', 'closed')),
  CONSTRAINT chk_cash_shifts_closed_at CHECK (
    (status = 'open' AND closed_at IS NULL) OR
    (status = 'closed' AND closed_at IS NOT NULL)
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_cash_shifts_open_by_cashier
  ON core.cash_shifts (tenant_id, office_id, cashier_id)
  WHERE status = 'open';

CREATE INDEX IF NOT EXISTS ix_cash_shifts_lookup
  ON core.cash_shifts (tenant_id, office_id, cashier_id, status, opened_at DESC);

COMMIT;
