import {
  Body,
  Controller,
  ForbiddenException,
  Get,
  Post,
  Query,
  UseGuards,
} from '@nestjs/common';

import { AuditService } from '../audit/audit.service';
import { Roles } from '../auth/decorators/roles.decorator';
import { RolesGuard } from '../auth/guards/roles.guard';
import { JwtPayload } from '../auth/interfaces/jwt-payload.interface';
import { JwtAuthGuard } from '../auth/jwt-auth.guard';
import { CurrentUser } from '../common/decorators/current-user.decorator';
import { OperatorPolicyService } from '../common/policies/operator-policy.service';
import { CloseShiftDto } from './dto/close-shift.dto';
import { OpenShiftDto } from './dto/open-shift.dto';
import { ShiftsService } from './shifts.service';

@Controller('shifts')
@UseGuards(JwtAuthGuard, RolesGuard)
export class ShiftsController {
  constructor(
    private readonly shiftsService: ShiftsService,
    private readonly policy: OperatorPolicyService,
    private readonly audit: AuditService,
  ) {}

  @Roles('owner', 'manager', 'cashier')
  @Post('open')
  async open(@CurrentUser() user: JwtPayload, @Body() body: OpenShiftDto) {
    this.policy.ensureCanOpenShift(user, body.office_id);
    const shift = this.shiftsService.openShift({
      tenant_id: user.tenant_id,
      office_id: body.office_id,
      user_id: user.sub,
      user_login: user.login,
      note: body.note,
    });

    await this.audit.record({
      tenant_id: user.tenant_id,
      office_id: body.office_id,
      actor_user_id: user.sub,
      actor_role: user.role,
      action: 'open_shift',
      entity_type: 'cash_shift',
      entity_id: `${user.sub}:${body.office_id}`,
      payload_snapshot: body as unknown as Record<string, unknown>,
    });

    return shift;
  }

  @Roles('owner', 'manager', 'cashier')
  @Post('close')
  async close(@CurrentUser() user: JwtPayload, @Body() body: CloseShiftDto) {
    this.policy.ensureCanCloseShift(user, body.office_id);
    const shift = this.shiftsService.closeShift({
      tenant_id: user.tenant_id,
      office_id: body.office_id,
      user_id: user.sub,
      user_login: user.login,
      note: body.note,
    });

    await this.audit.record({
      tenant_id: user.tenant_id,
      office_id: body.office_id,
      actor_user_id: user.sub,
      actor_role: user.role,
      action: 'close_shift',
      entity_type: 'cash_shift',
      entity_id: `${user.sub}:${body.office_id}`,
      payload_snapshot: body as unknown as Record<string, unknown>,
    });

    return shift;
  }

  @Roles('owner', 'manager', 'cashier')
  @Get('current')
  async current(
    @CurrentUser() user: JwtPayload,
    @Query('office_id') officeId?: string,
    @Query('user_id') userId?: string,
  ) {
    const scopedOffice = officeId ?? user.office_id ?? '';
    if (!scopedOffice) {
      throw new ForbiddenException({
        error: {
          code: 'OFFICE_REQUIRED',
          message: 'office_id is required',
        },
      });
    }
    this.policy.ensureCanViewShift(user, scopedOffice);

    const targetUser = user.role === 'owner' || user.role === 'manager' ? userId ?? user.sub : user.sub;
    return this.shiftsService.currentShift(user.tenant_id, scopedOffice, targetUser);
  }
}
