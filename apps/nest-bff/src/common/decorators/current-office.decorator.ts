import { createParamDecorator, ExecutionContext } from '@nestjs/common';
import { Request } from 'express';

import { JwtPayload } from '../../auth/interfaces/jwt-payload.interface';

export const CurrentOffice = createParamDecorator(
  (_data: unknown, ctx: ExecutionContext): string | null | undefined => {
    const req = ctx.switchToHttp().getRequest<Request & { user?: JwtPayload }>();
    return req.user?.office_id ?? null;
  },
);
