import { createParamDecorator, ExecutionContext } from '@nestjs/common';
import { Request } from 'express';

import { JwtPayload, UserRole } from '../../auth/interfaces/jwt-payload.interface';

export const CurrentRole = createParamDecorator(
  (_data: unknown, ctx: ExecutionContext): UserRole | undefined => {
    const req = ctx.switchToHttp().getRequest<Request & { user?: JwtPayload }>();
    return req.user?.role;
  },
);
