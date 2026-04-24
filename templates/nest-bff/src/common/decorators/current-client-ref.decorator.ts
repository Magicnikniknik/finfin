import {
  BadRequestException,
  createParamDecorator,
  ExecutionContext,
  UnauthorizedException,
} from '@nestjs/common';
import { Request } from 'express';

export const CurrentClientRef = createParamDecorator(
  (_data: unknown, ctx: ExecutionContext): string => {
    const request = ctx.switchToHttp().getRequest<Request>();
    const value = getSingleHeader(request, 'x-client-ref');

    if (!value) {
      throw new UnauthorizedException({
        error: {
          code: 'MISSING_CLIENT_REF',
          message: 'x-client-ref header is required',
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
