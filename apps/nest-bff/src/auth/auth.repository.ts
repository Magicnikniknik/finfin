import { Injectable, OnModuleDestroy } from '@nestjs/common';
import { Pool } from 'pg';

import { UserRole } from './interfaces/jwt-payload.interface';

export type AuthUser = {
  id: string;
  tenant_id: string;
  office_id: string | null;
  login: string;
  password_hash: string;
  role: UserRole;
  status: 'active' | 'disabled';
  display_name: string | null;
};

export type RefreshToken = {
  id: string;
  user_id: string;
  token_hash: string;
  expires_at: Date;
  revoked_at: Date | null;
};

@Injectable()
export class AuthRepository implements OnModuleDestroy {
  private readonly pool: Pool;

  constructor() {
    this.pool = new Pool({
      connectionString: process.env.DATABASE_URL,
    });
  }

  async onModuleDestroy(): Promise<void> {
    await this.pool.end();
  }

  async findUserByTenantAndLogin(
    tenantId: string,
    login: string,
  ): Promise<AuthUser | null> {
    const result = await this.pool.query<AuthUser>(
      `
SELECT
  id::text,
  tenant_id::text,
  office_id::text,
  login,
  password_hash,
  role,
  status,
  display_name
FROM auth.users
WHERE tenant_id = $1::uuid
  AND login = $2
LIMIT 1
`,
      [tenantId, login],
    );

    return result.rows[0] ?? null;
  }

  async findUserByID(userId: string): Promise<AuthUser | null> {
    const result = await this.pool.query<AuthUser>(
      `
SELECT
  id::text,
  tenant_id::text,
  office_id::text,
  login,
  password_hash,
  role,
  status,
  display_name
FROM auth.users
WHERE id = $1::uuid
LIMIT 1
`,
      [userId],
    );

    return result.rows[0] ?? null;
  }

  async insertRefreshToken(params: {
    userId: string;
    tokenHash: string;
    expiresAt: Date;
    userAgent?: string | null;
    ip?: string | null;
  }): Promise<void> {
    await this.pool.query(
      `
INSERT INTO auth.refresh_tokens (
  user_id,
  token_hash,
  expires_at,
  user_agent,
  ip
)
VALUES ($1::uuid, $2, $3, $4, $5)
`,
      [
        params.userId,
        params.tokenHash,
        params.expiresAt,
        params.userAgent ?? null,
        params.ip ?? null,
      ],
    );
  }

  async findRefreshTokenByHash(tokenHash: string): Promise<RefreshToken | null> {
    const result = await this.pool.query<RefreshToken>(
      `
SELECT
  id::text,
  user_id::text,
  token_hash,
  expires_at,
  revoked_at
FROM auth.refresh_tokens
WHERE token_hash = $1
LIMIT 1
`,
      [tokenHash],
    );

    return result.rows[0] ?? null;
  }

  async revokeRefreshTokenByHash(tokenHash: string): Promise<void> {
    await this.pool.query(
      `
UPDATE auth.refresh_tokens
SET revoked_at = now()
WHERE token_hash = $1
  AND revoked_at IS NULL
`,
      [tokenHash],
    );
  }
}
