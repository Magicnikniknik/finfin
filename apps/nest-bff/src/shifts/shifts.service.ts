import { Injectable } from '@nestjs/common';

type ShiftState = {
  tenant_id: string;
  office_id: string;
  user_id: string;
  user_login: string;
  opened_at: string;
  status: 'open' | 'closed';
  note?: string;
};

@Injectable()
export class ShiftsService {
  private readonly shifts = new Map<string, ShiftState>();

  openShift(params: {
    tenant_id: string;
    office_id: string;
    user_id: string;
    user_login: string;
    note?: string;
  }): ShiftState {
    const key = this.key(params.tenant_id, params.office_id, params.user_id);
    const current = this.shifts.get(key);
    if (current?.status === 'open') {
      return current;
    }

    const next: ShiftState = {
      tenant_id: params.tenant_id,
      office_id: params.office_id,
      user_id: params.user_id,
      user_login: params.user_login,
      opened_at: new Date().toISOString(),
      status: 'open',
      note: params.note,
    };

    this.shifts.set(key, next);
    return next;
  }

  closeShift(params: {
    tenant_id: string;
    office_id: string;
    user_id: string;
    user_login: string;
    note?: string;
  }): ShiftState {
    const key = this.key(params.tenant_id, params.office_id, params.user_id);
    const current = this.shifts.get(key);

    const next: ShiftState = {
      tenant_id: params.tenant_id,
      office_id: params.office_id,
      user_id: params.user_id,
      user_login: params.user_login,
      opened_at: current?.opened_at ?? new Date().toISOString(),
      status: 'closed',
      note: params.note,
    };

    this.shifts.set(key, next);
    return next;
  }

  currentShift(tenantId: string, officeId: string, userId: string): ShiftState | null {
    const key = this.key(tenantId, officeId, userId);
    return this.shifts.get(key) ?? null;
  }

  hasActiveShift(tenantId: string, officeId: string, userId: string): boolean {
    const current = this.currentShift(tenantId, officeId, userId);
    return current?.status === 'open';
  }

  private key(tenantId: string, officeId: string, userId: string): string {
    return `${tenantId}:${officeId}:${userId}`;
  }
}
