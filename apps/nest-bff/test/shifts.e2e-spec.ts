import { INestApplication, ValidationPipe } from '@nestjs/common';
import { Test, TestingModule } from '@nestjs/testing';
import request = require('supertest');

import { AuditService } from '../src/audit/audit.service';
import { RolesGuard } from '../src/auth/guards/roles.guard';
import { JwtAuthGuard } from '../src/auth/jwt-auth.guard';
import { GlobalExceptionFilter } from '../src/common/filters/global-exception.filter';
import { OperatorPolicyService } from '../src/common/policies/operator-policy.service';
import { ShiftsController } from '../src/shifts/shifts.controller';
import { ShiftsService } from '../src/shifts/shifts.service';

describe('Shifts + thin gate (e2e)', () => {
  let app: INestApplication;
  let shifts: ShiftsService;

  const token = `x.${Buffer.from(JSON.stringify({ sub: 'cashier-1', tenant_id: '11111111-1111-1111-1111-111111111111', office_id: '22222222-2222-4222-8222-222222222222', role: 'cashier', login: 'cashier_1', exp: Math.floor(Date.now()/1000)+3600 })).toString('base64url')}.y`;

  beforeAll(async () => {
    const moduleRef: TestingModule = await Test.createTestingModule({
      controllers: [ShiftsController],
      providers: [
        ShiftsService,
        OperatorPolicyService,
        AuditService,
        { provide: JwtAuthGuard, useValue: { canActivate: () => true } },
        { provide: RolesGuard, useValue: { canActivate: () => true } },
      ],
    }).compile();

    shifts = moduleRef.get(ShiftsService);
    app = moduleRef.createNestApplication();
    app.useGlobalPipes(new ValidationPipe({ whitelist: true, transform: true }));

    app.use((req: any, _res: any, next: any) => {
      req.user = {
        sub: req.headers['x-user-id'] || 'cashier-1',
        tenant_id: '11111111-1111-1111-1111-111111111111',
        office_id: req.headers['x-office-id'] || '22222222-2222-4222-8222-222222222222',
        scope_office_ids: ['22222222-2222-4222-8222-222222222222'],
        role: req.headers['x-role'] || 'cashier',
        login: req.headers['x-login'] || 'cashier_1',
      };
      next();
    });

    app.useGlobalFilters(new GlobalExceptionFilter());
    await app.init();
  });

  afterAll(async () => {
    await app.close();
  });

  it('cashier can open shift in own office', async () => {
    const res = await request(app.getHttpServer())
      .post('/shifts/open')
      .set('authorization', `Bearer ${token}`)
      .send({ office_id: '22222222-2222-4222-8222-222222222222' })
      .expect(201);

    expect(res.body.status).toBe('open');
  });

  it('cashier blocked in foreign office', async () => {
    const res = await request(app.getHttpServer())
      .post('/shifts/open')
      .set('authorization', `Bearer ${token}`)
      .send({ office_id: '33333333-3333-4333-8333-333333333333' })
      .expect(403);

    expect(res.body.error.code).toBe('OFFICE_SCOPE_VIOLATION');
  });

  it('open+close shift lifecycle works', async () => {
    await request(app.getHttpServer())
      .post('/shifts/open')
      .set('authorization', `Bearer ${token}`)
      .send({ office_id: '22222222-2222-4222-8222-222222222222' })
      .expect(201);

    const closed = await request(app.getHttpServer())
      .post('/shifts/close')
      .set('authorization', `Bearer ${token}`)
      .send({ office_id: '22222222-2222-4222-8222-222222222222' })
      .expect(201);

    expect(closed.body.status).toBe('closed');
  });

  it('active shift state toggles for gate checks', () => {
    shifts.closeShift({
      tenant_id: '11111111-1111-1111-1111-111111111111',
      office_id: '22222222-2222-4222-8222-222222222222',
      user_id: 'cashier-1',
      user_login: 'cashier_1',
    });
    expect(
      shifts.hasActiveShift(
        '11111111-1111-1111-1111-111111111111',
        '22222222-2222-4222-8222-222222222222',
        'cashier-1',
      ),
    ).toBe(false);

    shifts.openShift({
      tenant_id: '11111111-1111-1111-1111-111111111111',
      office_id: '22222222-2222-4222-8222-222222222222',
      user_id: 'cashier-1',
      user_login: 'cashier_1',
    });

    expect(
      shifts.hasActiveShift(
        '11111111-1111-1111-1111-111111111111',
        '22222222-2222-4222-8222-222222222222',
        'cashier-1',
      ),
    ).toBe(true);
  });
});
