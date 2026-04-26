import { Injectable } from '@nestjs/common';

import { UserRole } from '../auth/interfaces/jwt-payload.interface';

export type AuditRecord = {
  tenant_id?: string | null;
  office_id?: string | null;
  actor_user_id?: string | null;
  actor_role?: UserRole | null;
  action: string;
  entity_type?: string | null;
  entity_id?: string | null;
  ip?: string | null;
  request_id?: string | null;
  payload_snapshot?: Record<string, unknown> | null;
  created_at?: string;
};

@Injectable()
export class AuditService {
  private readonly records: AuditRecord[] = [];

  async record(input: AuditRecord): Promise<void> {
    this.records.push({
      ...input,
      created_at: new Date().toISOString(),
    });
  }

  all(): AuditRecord[] {
    return [...this.records];
  }

  clear(): void {
    this.records.length = 0;
  }
}
