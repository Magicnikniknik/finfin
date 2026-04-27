import { Module } from '@nestjs/common';

import { AuditModule } from './audit/audit.module';
import { AuthModule } from './auth/auth.module';
import { HealthModule } from './health/health.module';
import { OrdersModule } from './orders/orders.module';
import { PricingModule } from './pricing/pricing.module';
import { ShiftsModule } from './shifts/shifts.module';

@Module({
  imports: [
    HealthModule,
    AuthModule,
    OrdersModule,
    PricingModule,
    ShiftsModule,
    AuditModule,
  ],
})
export class AppModule {}
