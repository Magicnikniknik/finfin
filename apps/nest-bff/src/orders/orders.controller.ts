import { Body, Controller, ForbiddenException, Post, UseGuards } from '@nestjs/common';

import { AuditService } from '../audit/audit.service';
import { Roles } from '../auth/decorators/roles.decorator';
import { RolesGuard } from '../auth/guards/roles.guard';
import { JwtPayload } from '../auth/interfaces/jwt-payload.interface';
import { JwtAuthGuard } from '../auth/jwt-auth.guard';
import { CurrentUser } from '../common/decorators/current-user.decorator';
import { OperatorPolicyService } from '../common/policies/operator-policy.service';
import { CancelOrderDto } from './dto/cancel-order.dto';
import { CompleteOrderDto } from './dto/complete-order.dto';
import { ReserveOrderDto } from './dto/reserve-order.dto';
import { OrdersService } from './orders.service';
import { ShiftsService } from '../shifts/shifts.service';

@Controller('orders')
@UseGuards(JwtAuthGuard, RolesGuard)
export class OrdersController {
  constructor(
    private readonly ordersService: OrdersService,
    private readonly policy: OperatorPolicyService,
    private readonly shifts: ShiftsService,
    private readonly audit: AuditService,
  ) {}

  @Roles('owner', 'manager', 'cashier')
  @Post('reserve')
  async reserve(@CurrentUser() user: JwtPayload, @Body() body: ReserveOrderDto) {
    this.policy.ensureCanReserveOrder(user, body.office_id);

    const result = await this.ordersService.reserve(user.tenant_id, user.login, body);

    await this.audit.record({
      tenant_id: user.tenant_id,
      office_id: body.office_id,
      actor_user_id: user.sub,
      actor_role: user.role,
      action: 'reserve',
      entity_type: 'order',
      entity_id: result.order_id,
      payload_snapshot: body as unknown as Record<string, unknown>,
    });

    return result;
  }

  @Roles('owner', 'manager', 'cashier')
  @Post('complete')
  async complete(@CurrentUser() user: JwtPayload, @Body() body: CompleteOrderDto) {
    this.policy.ensureCanCompleteOrder(user);

    if (user.role === 'cashier') {
      const officeId = user.office_id ?? '';
      if (!officeId || !this.shifts.hasActiveShift(user.tenant_id, officeId, user.sub)) {
        throw new ForbiddenException({
          error: {
            code: 'SHIFT_NOT_OPEN',
            message: 'Cashier requires an active shift to complete orders',
          },
        });
      }
    }

    const payload: CompleteOrderDto = {
      ...body,
      cashier_id: user.role === 'cashier' ? user.login : body.cashier_id,
    };

    const result = await this.ordersService.complete(user.tenant_id, user.login, payload);

    await this.audit.record({
      tenant_id: user.tenant_id,
      office_id: user.office_id ?? null,
      actor_user_id: user.sub,
      actor_role: user.role,
      action: 'complete',
      entity_type: 'order',
      entity_id: body.order_id,
      payload_snapshot: payload as unknown as Record<string, unknown>,
    });

    return result;
  }

  @Roles('owner', 'manager', 'cashier')
  @Post('cancel')
  async cancel(@CurrentUser() user: JwtPayload, @Body() body: CancelOrderDto) {
    this.policy.ensureCanCancelOrder(user);

    if (user.role === 'cashier') {
      const officeId = user.office_id ?? '';
      if (!officeId || !this.shifts.hasActiveShift(user.tenant_id, officeId, user.sub)) {
        throw new ForbiddenException({
          error: {
            code: 'SHIFT_NOT_OPEN',
            message: 'Cashier requires an active shift to cancel orders',
          },
        });
      }
    }

    const result = await this.ordersService.cancel(user.tenant_id, user.login, body);

    await this.audit.record({
      tenant_id: user.tenant_id,
      office_id: user.office_id ?? null,
      actor_user_id: user.sub,
      actor_role: user.role,
      action: 'cancel',
      entity_type: 'order',
      entity_id: body.order_id,
      payload_snapshot: body as unknown as Record<string, unknown>,
    });

    return result;
  }
}
