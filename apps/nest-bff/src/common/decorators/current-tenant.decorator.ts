import {
  BadRequestException,
  createParamDecorator,
  ExecutionContext,
  UnauthorizedException,
} from '@nestjs/common';
import { Request } from 'express';

import { JwtPayload } from '../../auth/interfaces/jwt-payload.interface';

export const CurrentTenant = createParamDecorator(
  (_data: unknown, ctx: ExecutionContext): string => {
    const request = ctx
      .switchToHttp()
      .getRequest<Request & { user?: JwtPayload }>();

    if (request.user?.tenant_id) {
      return request.user.tenant_id;
    }

    const value = getSingleHeader(request, 'x-tenant-id');
    if (!value) {
      throw new UnauthorizedException({
        error: {
          code: 'MISSING_TENANT',
          message: 'tenant_id is required',
        },
      });
    }

    return value;
  },
);

function getSingleHeader(request: Request, name: string): string | undefined {
  const raw = request.headers[name];

  if (Array.isArray(raw)) {
    if (raw.length !== 1) {
      throw new BadRequestException({
        error: {
          code: 'INVALID_HEADER',
          message: `${name} header must be singular`,
        },
      });
    }
    return raw[0]?.trim() || undefined;
  }

  if (typeof raw === 'string') {
    return raw.trim() || undefined;
  }

  return undefined;
}
