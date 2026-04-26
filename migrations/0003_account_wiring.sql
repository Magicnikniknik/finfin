BEGIN;

CREATE TABLE IF NOT EXISTS core.account_wiring (
  tenant_id uuid NOT NULL,
  office_id uuid NOT NULL,
  currency_id uuid NOT NULL,
  balance_account_id uuid NOT NULL REFERENCES core.accounts(id),
  available_ledger_account_id uuid NOT NULL REFERENCES core.accounts(id),
  reserved_ledger_account_id uuid NOT NULL REFERENCES core.accounts(id),
  settlement_ledger_account_id uuid REFERENCES core.accounts(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT pk_account_wiring PRIMARY KEY (tenant_id, office_id, currency_id)
);

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'core'
      AND table_name = 'account_wiring'
      AND column_name = 'hold_currency_id'
  ) AND NOT EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = 'core'
      AND table_name = 'account_wiring'
      AND column_name = 'currency_id'
  ) THEN
    EXECUTE 'ALTER TABLE core.account_wiring RENAME COLUMN hold_currency_id TO currency_id';
  END IF;
END $$;

ALTER TABLE core.account_wiring
  ALTER COLUMN tenant_id SET NOT NULL,
  ALTER COLUMN office_id SET NOT NULL,
  ALTER COLUMN currency_id SET NOT NULL,
  ALTER COLUMN balance_account_id SET NOT NULL,
  ALTER COLUMN available_ledger_account_id SET NOT NULL,
  ALTER COLUMN reserved_ledger_account_id SET NOT NULL,
  ALTER COLUMN created_at SET DEFAULT now(),
  ALTER COLUMN updated_at SET DEFAULT now();

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE contype = 'p'
      AND conrelid = 'core.account_wiring'::regclass
  ) THEN
    EXECUTE 'ALTER TABLE core.account_wiring ADD CONSTRAINT pk_account_wiring PRIMARY KEY (tenant_id, office_id, currency_id)';
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS ix_account_wiring_lookup
  ON core.account_wiring (tenant_id, office_id, currency_id);

COMMIT;
