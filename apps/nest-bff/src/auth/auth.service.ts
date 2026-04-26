import { Injectable, UnauthorizedException } from '@nestjs/common';
import { createHash, createHmac } from 'crypto';

import { AuthRepository } from './auth.repository';
import { LoginDto } from './dto/login.dto';
import { RefreshDto } from './dto/refresh.dto';
import { JwtPayload } from './interfaces/jwt-payload.interface';
import { PasswordService } from './password.service';

@Injectable()
export class AuthService {
  private readonly accessSecret = process.env.JWT_ACCESS_SECRET || 'dev_access_secret';
  private readonly refreshSecret = process.env.JWT_REFRESH_SECRET || 'dev_refresh_secret';
  private readonly accessTtl = process.env.JWT_ACCESS_TTL || '15m';
  private readonly refreshTtl = process.env.JWT_REFRESH_TTL || '30d';

  constructor(
    private readonly repo: AuthRepository,
    private readonly passwordService: PasswordService,
  ) {}

  async login(dto: LoginDto, meta?: { ip?: string; userAgent?: string }) {
    const user = await this.repo.findUserByTenantAndLogin(dto.tenant_id, dto.login);

    if (!user) {
      throw unauthorized('INVALID_CREDENTIALS', 'Invalid login or password');
    }

    if (user.status !== 'active') {
      throw unauthorized('USER_DISABLED', 'User is disabled');
    }

    const validPassword = await this.passwordService.verify(dto.password, user.password_hash);
    if (!validPassword) {
      throw unauthorized('INVALID_CREDENTIALS', 'Invalid login or password');
    }

    return this.issueTokens(
      {
        id: user.id,
        tenant_id: user.tenant_id,
        office_id: user.office_id,
        role: user.role,
        login: user.login,
        display_name: user.display_name,
      },
      meta,
    );
  }

  async refresh(dto: RefreshDto, meta?: { ip?: string; userAgent?: string }) {
    const verified = verifyJwt(dto.refresh_token, this.refreshSecret);
    if (!verified) {
      throw unauthorized('INVALID_REFRESH_TOKEN', 'Refresh token is invalid');
    }

    const tokenHash = hashToken(dto.refresh_token);
    const stored = await this.repo.findRefreshTokenByHash(tokenHash);

    if (!stored) {
      throw unauthorized('INVALID_REFRESH_TOKEN', 'Refresh token is invalid');
    }

    if (stored.revoked_at) {
      throw unauthorized('REFRESH_TOKEN_REVOKED', 'Refresh token is revoked');
    }

    if (new Date(stored.expires_at).getTime() <= Date.now()) {
      throw unauthorized('REFRESH_TOKEN_EXPIRED', 'Refresh token is expired');
    }

    const user = await this.repo.findUserByID(stored.user_id);
    if (!user || user.status !== 'active') {
      throw unauthorized('USER_DISABLED', 'User is disabled');
    }

    await this.repo.revokeRefreshTokenByHash(tokenHash);

    return this.issueTokens(
      {
        id: user.id,
        tenant_id: user.tenant_id,
        office_id: user.office_id,
        role: user.role,
        login: user.login,
        display_name: user.display_name,
      },
      meta,
    );
  }

  async me(user: JwtPayload) {
    return {
      id: user.sub,
      tenant_id: user.tenant_id,
      office_id: user.office_id ?? null,
      role: user.role,
      login: user.login,
    };
  }

  private async issueTokens(
    user: {
      id: string;
      tenant_id: string;
      office_id?: string | null;
      role: 'owner' | 'manager' | 'cashier';
      login: string;
      display_name?: string | null;
    },
    meta?: { ip?: string; userAgent?: string },
  ) {
    const payload: JwtPayload = {
      sub: user.id,
      tenant_id: user.tenant_id,
      office_id: user.office_id ?? null,
      scope_office_ids: defaultScopeForRole(user.role, user.office_id ?? null),
      role: user.role,
      login: user.login,
    };

    const accessToken = signJwt(payload, this.accessSecret, this.accessTtl);
    const refreshToken = signJwt(payload, this.refreshSecret, this.refreshTtl);

    const refreshExp = decodeExpiry(refreshToken);
    await this.repo.insertRefreshToken({
      userId: user.id,
      tokenHash: hashToken(refreshToken),
      expiresAt: refreshExp,
      ip: meta?.ip ?? null,
      userAgent: meta?.userAgent ?? null,
    });

    return {
      access_token: accessToken,
      refresh_token: refreshToken,
      user: {
        id: user.id,
        tenant_id: user.tenant_id,
        office_id: user.office_id ?? null,
        role: user.role,
        login: user.login,
        display_name: user.display_name ?? null,
      },
    };
  }
}

function unauthorized(code: string, message: string) {
  return new UnauthorizedException({
    error: { code, message },
  });
}

function hashToken(token: string): string {
  return createHash('sha256').update(token).digest('hex');
}

function decodeExpiry(token: string): Date {
  const payload = parseJwtPayload(token);
  if (!payload?.exp) {
    throw new Error('refresh token missing exp');
  }
  return new Date(payload.exp * 1000);
}

function defaultScopeForRole(role: 'owner' | 'manager' | 'cashier', officeId: string | null): string[] {
  if (role === 'owner' || !officeId) {
    return [];
  }
  return [officeId];
}

function signJwt(payload: JwtPayload, secret: string, ttl: string): string {
  const header = { alg: 'HS256', typ: 'JWT' };
  const exp = Math.floor((Date.now() + parseDuration(ttl)) / 1000);
  const body = { ...payload, exp };

  const encodedHeader = base64UrlEncode(JSON.stringify(header));
  const encodedBody = base64UrlEncode(JSON.stringify(body));
  const unsigned = `${encodedHeader}.${encodedBody}`;

  const signature = createHmac('sha256', secret)
    .update(unsigned)
    .digest('base64url');

  return `${unsigned}.${signature}`;
}

function verifyJwt(token: string, secret: string): boolean {
  const parts = token.split('.');
  if (parts.length !== 3) {
    return false;
  }

  const [encodedHeader, encodedBody, signature] = parts;
  const unsigned = `${encodedHeader}.${encodedBody}`;
  const expected = createHmac('sha256', secret).update(unsigned).digest('base64url');

  if (expected !== signature) {
    return false;
  }

  const payload = parseJwtPayload(token);
  if (!payload?.exp) {
    return false;
  }

  return payload.exp * 1000 > Date.now();
}

function parseJwtPayload(token: string): Record<string, any> | null {
  const parts = token.split('.');
  if (parts.length !== 3) {
    return null;
  }

  try {
    const raw = Buffer.from(parts[1], 'base64url').toString('utf8');
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function parseDuration(ttl: string): number {
  const match = /^(\d+)([smhd])$/.exec(ttl.trim());
  if (!match) {
    throw new Error(`invalid duration format: ${ttl}`);
  }

  const amount = Number(match[1]);
  const unit = match[2];

  if (unit === 's') return amount * 1000;
  if (unit === 'm') return amount * 60 * 1000;
  if (unit === 'h') return amount * 60 * 60 * 1000;
  return amount * 24 * 60 * 60 * 1000;
}

function base64UrlEncode(value: string): string {
  return Buffer.from(value, 'utf8').toString('base64url');
}
