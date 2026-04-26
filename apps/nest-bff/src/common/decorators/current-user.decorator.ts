import { createParamDecorator, ExecutionContext } from '@nestjs/common';
import { Request } from 'express';

import { JwtPayload } from '../../auth/interfaces/jwt-payload.interface';

export const CurrentUser = createParamDecorator(
  (_data: unknown, ctx: ExecutionContext): JwtPayload | undefined => {
    const request = ctx.switchToHttp().getRequest<Request & { user?: JwtPayload }>();
    return request.user;
  },
);
