import { Body, Controller, Post, UseGuards } from '@nestjs/common';

import { AuditService } from '../audit/audit.service';
import { Roles } from '../auth/decorators/roles.decorator';
import { RolesGuard } from '../auth/guards/roles.guard';
import { JwtAuthGuard } from '../auth/jwt-auth.guard';
import { JwtPayload } from '../auth/interfaces/jwt-payload.interface';
import { CurrentUser } from '../common/decorators/current-user.decorator';
import { OperatorPolicyService } from '../common/policies/operator-policy.service';
import { CalculateQuoteDto } from './dto/calculate-quote.dto';
import { PricingService } from './pricing.service';

@Controller('quotes')
@UseGuards(JwtAuthGuard, RolesGuard)
export class PricingController {
  constructor(
    private readonly pricingService: PricingService,
    private readonly policy: OperatorPolicyService,
    private readonly audit: AuditService,
  ) {}

  @Roles('owner', 'manager', 'cashier')
  @Post('calculate')
  async calculate(@CurrentUser() user: JwtPayload, @Body() body: CalculateQuoteDto) {
    this.policy.ensureCanCalculateQuote(user, body.office_id);

    const result = await this.pricingService.calculate(user.tenant_id, user.login, body);

    await this.audit.record({
      tenant_id: user.tenant_id,
      office_id: body.office_id,
      actor_user_id: user.sub,
      actor_role: user.role,
      action: 'quote_calculate',
      entity_type: 'quote',
      entity_id: result.quote_id,
      payload_snapshot: body as unknown as Record<string, unknown>,
    });

    return result;
  }
}
