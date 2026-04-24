CREATE TABLE IF NOT EXISTS core.quote_snapshots (
    id TEXT PRIMARY KEY,
    tenant_id UUID NOT NULL,
    office_id UUID NOT NULL,
    side TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    give_currency_id UUID NOT NULL,
    give_currency_code TEXT NOT NULL,
    give_currency_network TEXT NOT NULL DEFAULT '',
    amount_give NUMERIC(38, 18) NOT NULL,
    get_currency_id UUID NOT NULL,
    get_currency_code TEXT NOT NULL,
    get_currency_network TEXT NOT NULL DEFAULT '',
    amount_get NUMERIC(38, 18) NOT NULL,
    fixed_rate NUMERIC(38, 18) NOT NULL,
    hold_currency_id UUID NOT NULL,
    hold_amount NUMERIC(38, 18) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_quote_snapshots_tenant_office_expires
    ON core.quote_snapshots (tenant_id, office_id, expires_at);
