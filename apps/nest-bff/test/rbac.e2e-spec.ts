import { ForbiddenException } from '@nestjs/common';

import { JwtPayload } from '../src/auth/interfaces/jwt-payload.interface';
import { OperatorPolicyService } from '../src/common/policies/operator-policy.service';

describe('RBAC + Office Scoping', () => {
  const policy = new OperatorPolicyService();

  const cashier: JwtPayload = {
    sub: 'cashier-1',
    tenant_id: 'tenant-1',
    office_id: 'office-a',
    scope_office_ids: ['office-a'],
    role: 'cashier',
    login: 'cashier-a',
  };

  const manager: JwtPayload = {
    sub: 'manager-1',
    tenant_id: 'tenant-1',
    office_id: 'office-a',
    scope_office_ids: ['office-a', 'office-b'],
    role: 'manager',
    login: 'manager-1',
  };

  const owner: JwtPayload = {
    sub: 'owner-1',
    tenant_id: 'tenant-1',
    office_id: null,
    scope_office_ids: [],
    role: 'owner',
    login: 'owner-1',
  };

  it('cashier can reserve in own office', () => {
    expect(() => policy.ensureCanReserveOrder(cashier, 'office-a')).not.toThrow();
  });

  it('cashier blocked in foreign office', () => {
    expect(() => policy.ensureCanReserveOrder(cashier, 'office-b')).toThrow(
      ForbiddenException,
    );
  });

  it('cashier can cancel (shift gate handled in orders flow)', () => {
    expect(() => policy.ensureCanCancelOrder(cashier)).not.toThrow();
  });

  it('manager can cancel', () => {
    expect(() => policy.ensureCanCancelOrder(manager)).not.toThrow();
  });

  it('owner can do all actions', () => {
    expect(() => policy.ensureCanCalculateQuote(owner, 'office-z')).not.toThrow();
    expect(() => policy.ensureCanReserveOrder(owner, 'office-z')).not.toThrow();
    expect(() => policy.ensureCanCompleteOrder(owner)).not.toThrow();
    expect(() => policy.ensureCanCancelOrder(owner)).not.toThrow();
  });
});
