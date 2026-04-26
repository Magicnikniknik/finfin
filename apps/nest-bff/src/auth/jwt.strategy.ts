import { Injectable, UnauthorizedException } from '@nestjs/common';

import { JwtPayload } from './interfaces/jwt-payload.interface';

@Injectable()
export class JwtStrategy {
  validate(payload: JwtPayload): JwtPayload {
    if (!payload?.sub || !payload?.tenant_id || !payload?.role) {
      throw new UnauthorizedException({
        error: {
          code: 'INVALID_ACCESS_TOKEN',
          message: 'Access token payload is invalid',
        },
      });
    }

    return payload;
  }
}
