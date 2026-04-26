import { ForbiddenException, Injectable } from '@nestjs/common';

import { JwtPayload } from '../../auth/interfaces/jwt-payload.interface';

type OperatorAction =
  | 'calculate_quote'
  | 'reserve_order'
  | 'complete_order'
  | 'cancel_order'
  | 'open_shift'
  | 'close_shift'
  | 'view_shift';

@Injectable()
export class OperatorPolicyService {
  ensureCanCalculateQuote(user: JwtPayload, officeId: string) {
    this.ensureAllowed(user, 'calculate_quote', officeId);
  }

  ensureCanReserveOrder(user: JwtPayload, officeId: string) {
    this.ensureAllowed(user, 'reserve_order', officeId);
  }

  ensureCanCompleteOrder(user: JwtPayload) {
    this.ensureAllowed(user, 'complete_order');
  }

  ensureCanCancelOrder(user: JwtPayload) {
    this.ensureAllowed(user, 'cancel_order');
  }

  ensureCanOpenShift(user: JwtPayload, officeId: string) {
    this.ensureAllowed(user, 'open_shift', officeId);
  }

  ensureCanCloseShift(user: JwtPayload, officeId: string) {
    this.ensureAllowed(user, 'close_shift', officeId);
  }

  ensureCanViewShift(user: JwtPayload, officeId: string) {
    this.ensureAllowed(user, 'view_shift', officeId);
  }

  private ensureAllowed(user: JwtPayload, action: OperatorAction, officeId?: string) {
    if (!user?.role) {
      throw forbidden('ROLE_REQUIRED', 'Operator role is required');
    }

    if (user.role === 'owner') {
      return;
    }

    if (user.role === 'cashier') {
      if (officeId && user.office_id !== officeId) {
        throw forbidden('OFFICE_SCOPE_VIOLATION', 'Cashier cannot access foreign office');
      }

      return;
    }

    if (user.role === 'manager') {
      if (!officeId) {
        return;
      }

      const scopedOffices = normalizedScope(user);
      if (scopedOffices.length === 0) {
        throw forbidden('OFFICE_SCOPE_MISSING', 'Manager office scope is not configured');
      }
      if (!scopedOffices.includes(officeId)) {
        throw forbidden('OFFICE_SCOPE_VIOLATION', 'Manager cannot access foreign office');
      }
    }
  }
}

function normalizedScope(user: JwtPayload): string[] {
  const values = user.scope_office_ids ?? [];
  return values.map((v) => v.trim()).filter(Boolean);
}

function forbidden(code: string, message: string): ForbiddenException {
  return new ForbiddenException({
    error: { code, message },
  });
}
