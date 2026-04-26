import {
  CanActivate,
  ExecutionContext,
  ForbiddenException,
  Injectable,
} from '@nestjs/common';
import { Reflector } from '@nestjs/core';
import { Request } from 'express';

import { JwtPayload, UserRole } from '../interfaces/jwt-payload.interface';
import { ROLES_KEY } from '../decorators/roles.decorator';

@Injectable()
export class RolesGuard implements CanActivate {
  constructor(private readonly reflector: Reflector) {}

  canActivate(context: ExecutionContext): boolean {
    const roles = this.reflector.getAllAndOverride<UserRole[]>(ROLES_KEY, [
      context.getHandler(),
      context.getClass(),
    ]);

    if (!roles || roles.length === 0) {
      return true;
    }

    const req = context.switchToHttp().getRequest<Request & { user?: JwtPayload }>();
    const user = req.user;

    if (!user?.role) {
      throw new ForbiddenException({
        error: {
          code: 'ROLE_REQUIRED',
          message: 'Role is required',
        },
      });
    }

    if (!roles.includes(user.role)) {
      throw new ForbiddenException({
        error: {
          code: 'ROLE_FORBIDDEN',
          message: 'Role is not allowed for this action',
        },
      });
    }

    return true;
  }
}
