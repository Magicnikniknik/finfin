import { Module } from '@nestjs/common';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { join } from 'path';

import { AuditModule } from '../audit/audit.module';
import { OperatorPolicyService } from '../common/policies/operator-policy.service';
import { ShiftsModule } from '../shifts/shifts.module';
import { OrdersController } from './orders.controller';
import { OrdersService } from './orders.service';

@Module({
  imports: [
    AuditModule,
    ShiftsModule,
    ClientsModule.register([
      {
        name: 'ORDER_PACKAGE',
        transport: Transport.GRPC,
        options: {
          url: process.env.ORDER_GRPC_URL || '127.0.0.1:9090',
          package: process.env.ORDER_GRPC_PACKAGE || 'exchange.order.v1',
          protoPath:
            process.env.ORDER_GRPC_PROTO_PATH ||
            join(
              process.cwd(),
              'proto',
              'exchange',
              'order',
              'v1',
              'order_service.proto',
            ),
        },
      },
    ]),
  ],
  controllers: [OrdersController],
  providers: [OrdersService, OperatorPolicyService],
  exports: [OrdersService],
})
export class OrdersModule {}
