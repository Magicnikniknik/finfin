import {
  CanActivate,
  ExecutionContext,
  Injectable,
  UnauthorizedException,
} from '@nestjs/common';
import { Request } from 'express';

import { JwtPayload } from './interfaces/jwt-payload.interface';

@Injectable()
export class JwtAuthGuard implements CanActivate {
  canActivate(context: ExecutionContext): boolean {
    const req = context
      .switchToHttp()
      .getRequest<Request & { user?: JwtPayload }>();

    const authHeader = req.headers.authorization;
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      throw unauthorized('MISSING_ACCESS_TOKEN', 'Bearer access token is required');
    }

    const token = authHeader.slice('Bearer '.length).trim();
    if (!token) {
      throw unauthorized('MISSING_ACCESS_TOKEN', 'Bearer access token is required');
    }

    const payload = decodeJwtPayload(token);
    if (!payload) {
      throw unauthorized('INVALID_ACCESS_TOKEN', 'Access token is invalid');
    }

    const exp = (payload as { exp?: number }).exp;
    if (typeof exp === 'number' && exp * 1000 <= Date.now()) {
      throw unauthorized('INVALID_ACCESS_TOKEN', 'Access token is invalid');
    }

    req.user = payload;
    return true;
  }
}

function decodeJwtPayload(token: string): JwtPayload | null {
  const parts = token.split('.');
  if (parts.length !== 3) {
    return null;
  }

  try {
    const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
    const raw = Buffer.from(base64, 'base64').toString('utf8');
    const parsed = JSON.parse(raw) as JwtPayload;
    if (!parsed?.sub || !parsed?.tenant_id || !parsed?.role) {
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

function unauthorized(code: string, message: string): UnauthorizedException {
  return new UnauthorizedException({
    error: { code, message },
  });
}
