import { AuditService } from '../src/audit/audit.service';

describe('Audit service', () => {
  it('writes and returns audit rows', async () => {
    const audit = new AuditService();

    await audit.record({
      tenant_id: '11111111-1111-1111-1111-111111111111',
      office_id: '22222222-2222-2222-2222-222222222222',
      actor_user_id: '00000000-0000-0000-0000-000000000001',
      actor_role: 'cashier',
      action: 'login',
      entity_type: 'user',
      entity_id: '00000000-0000-0000-0000-000000000001',
    });

    const records = audit.all();
    expect(records).toHaveLength(1);
    expect(records[0].action).toBe('login');
    expect(records[0].actor_role).toBe('cashier');
    expect(records[0].created_at).toBeDefined();
  });
});
