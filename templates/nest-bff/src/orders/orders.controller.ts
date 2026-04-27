import { Body, Controller, Post } from '@nestjs/common';

import { CurrentClientRef } from '../common/decorators/current-client-ref.decorator';
import { CurrentTenant } from '../common/decorators/current-tenant.decorator';
import { CancelOrderDto } from './dto/cancel-order.dto';
import { CompleteOrderDto } from './dto/complete-order.dto';
import { ReserveOrderDto } from './dto/reserve-order.dto';
import { OrdersService } from './orders.service';

@Controller('orders')
export class OrdersController {
  constructor(private readonly ordersService: OrdersService) {}

  @Post('reserve')
  async reserve(
    @CurrentTenant() tenantId: string,
    @CurrentClientRef() clientRef: string,
    @Body() body: ReserveOrderDto,
  ) {
    return this.ordersService.reserve(tenantId, clientRef, body);
  }

  @Post('complete')
  async complete(
    @CurrentTenant() tenantId: string,
    @CurrentClientRef() clientRef: string,
    @Body() body: CompleteOrderDto,
  ) {
    return this.ordersService.complete(tenantId, clientRef, body);
  }

  @Post('cancel')
  async cancel(
    @CurrentTenant() tenantId: string,
    @CurrentClientRef() clientRef: string,
    @Body() body: CancelOrderDto,
  ) {
    return this.ordersService.cancel(tenantId, clientRef, body);
  }
}
