import { Module } from '@nestjs/common';

import { AuditModule } from '../audit/audit.module';
import { OperatorPolicyService } from '../common/policies/operator-policy.service';
import { ShiftsController } from './shifts.controller';
import { ShiftsService } from './shifts.service';

@Module({
  imports: [AuditModule],
  controllers: [ShiftsController],
  providers: [ShiftsService, OperatorPolicyService],
  exports: [ShiftsService],
})
export class ShiftsModule {}
