import { Module } from '@nestjs/common';
import { ClientsModule, Transport } from '@nestjs/microservices';
import { join } from 'path';

import { AuditModule } from '../audit/audit.module';
import { OperatorPolicyService } from '../common/policies/operator-policy.service';
import { PricingController } from './pricing.controller';
import { PricingService } from './pricing.service';

@Module({
  imports: [
    AuditModule,
    ClientsModule.register([
      {
        name: 'PRICING_PACKAGE',
        transport: Transport.GRPC,
        options: {
          url: process.env.PRICING_GRPC_URL || '127.0.0.1:9090',
          package: process.env.PRICING_GRPC_PACKAGE || 'exchange.pricing.v1',
          protoPath:
            process.env.PRICING_GRPC_PROTO_PATH ||
            join(
              process.cwd(),
              'proto',
              'exchange',
              'pricing',
              'v1',
              'pricing_service.proto',
            ),
        },
      },
    ]),
  ],
  controllers: [PricingController],
  providers: [PricingService, OperatorPolicyService],
  exports: [PricingService],
})
export class PricingModule {}
