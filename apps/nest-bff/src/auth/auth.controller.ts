import { Body, Controller, Get, Post, Req, UseGuards } from '@nestjs/common';
import { Request } from 'express';

import { AuditService } from '../audit/audit.service';
import { CurrentUser } from '../common/decorators/current-user.decorator';
import { AuthService } from './auth.service';
import { LoginDto } from './dto/login.dto';
import { RefreshDto } from './dto/refresh.dto';
import { JwtPayload } from './interfaces/jwt-payload.interface';
import { JwtAuthGuard } from './jwt-auth.guard';

@Controller('auth')
export class AuthController {
  constructor(
    private readonly authService: AuthService,
    private readonly audit: AuditService,
  ) {}

  @Post('login')
  async login(@Body() body: LoginDto, @Req() req: Request) {
    const result = await this.authService.login(body, {
      userAgent: headerAsString(req.headers['user-agent']),
      ip: req.ip,
    });

    await this.audit.record({
      tenant_id: result.user.tenant_id,
      office_id: result.user.office_id,
      actor_user_id: result.user.id,
      actor_role: result.user.role,
      action: 'login',
      entity_type: 'user',
      entity_id: result.user.id,
      ip: req.ip,
      request_id: headerAsString(req.headers['x-request-id']) ?? null,
    });

    return result;
  }

  @Post('refresh')
  async refresh(@Body() body: RefreshDto, @Req() req: Request) {
    return this.authService.refresh(body, {
      userAgent: headerAsString(req.headers['user-agent']),
      ip: req.ip,
    });
  }

  @UseGuards(JwtAuthGuard)
  @Get('me')
  me(@CurrentUser() user: JwtPayload) {
    return this.authService.me(user);
  }
}

function headerAsString(value: string | string[] | undefined): string | undefined {
  if (Array.isArray(value)) {
    return value[0];
  }
  return value;
}
