import { ForbiddenException, INestApplication, ValidationPipe } from '@nestjs/common';
import { Test, TestingModule } from '@nestjs/testing';
import type { NextFunction, Response } from 'express';
import request = require('supertest');

import { JwtAuthGuard } from '../src/auth/jwt-auth.guard';
import { RolesGuard } from '../src/auth/guards/roles.guard';

import { GlobalExceptionFilter } from '../src/common/filters/global-exception.filter';
import { AuditService } from '../src/audit/audit.service';
import { ShiftsService } from '../src/shifts/shifts.service';
import { OrdersController } from '../src/orders/orders.controller';
import { OperatorPolicyService } from '../src/common/policies/operator-policy.service';
import { OrdersService } from '../src/orders/orders.service';

describe('OrdersController (e2e)', () => {
  let app: INestApplication;

  const ordersServiceMock = {
    reserve: jest.fn(),
    complete: jest.fn(),
    cancel: jest.fn(),
  };

  const shiftsMock = { hasActiveShift: jest.fn(() => true) };
  const auditMock = { record: jest.fn() };

  const policyMock = {
    ensureCanReserveOrder: jest.fn(),
    ensureCanCompleteOrder: jest.fn(),
    ensureCanCancelOrder: jest.fn(),
  };

  const jwtGuardMock = { canActivate: jest.fn(() => true) };
  const rolesGuardMock = { canActivate: jest.fn(() => true) };

  beforeAll(async () => {
    jest.spyOn(JwtAuthGuard.prototype, 'canActivate').mockReturnValue(true as any);
    jest.spyOn(RolesGuard.prototype, 'canActivate').mockReturnValue(true as any);

    const moduleRef: TestingModule = await Test.createTestingModule({
      controllers: [OrdersController],
      providers: [
        { provide: JwtAuthGuard, useValue: jwtGuardMock },
        { provide: RolesGuard, useValue: rolesGuardMock },
        {
          provide: OrdersService,
          useValue: ordersServiceMock,
        },
        {
          provide: OperatorPolicyService,
          useValue: policyMock,
        },
        { provide: ShiftsService, useValue: shiftsMock },
        { provide: AuditService, useValue: auditMock },
      ],
    }).compile();

    app = moduleRef.createNestApplication();

    app.useGlobalPipes(
      new ValidationPipe({
        whitelist: true,
        transform: true,
      }),
    );

    app.use((req: any, _res: Response, next: NextFunction) => {
      req.user = {
        sub: 'u-1',
        tenant_id: 'tenant-1',
        office_id: 'office-1',
        role: req.headers['x-role'] || 'owner',
        login: 'operator-1',
        scope_office_ids: ['office-1'],
      };
      next();
    });

    app.useGlobalFilters(new GlobalExceptionFilter());
    await app.init();
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  afterAll(async () => {
    await app.close();
  });

  it('POST /orders/reserve -> happy path', async () => {
    ordersServiceMock.reserve.mockResolvedValue({
      order_id: 'order-123',
      status: 'RESERVED',
      expires_at_ts: '1735000000',
      version: '1',
    });

    const payload = {
      idempotency_key: 'idem-reserve-1',
      office_id: 'office-1',
      quote_id: 'quote-1',
      side: 'BUY',
      give: { amount: '100.00', currency: { code: 'USDT', network: 'TRC20' } },
      get: { amount: '3550.00', currency: { code: 'THB', network: 'cash' } },
    };

    const res = await request(app.getHttpServer())
      .post('/orders/reserve')
      .send(payload)
      .expect(201);

    expect(res.body).toEqual({
      order_id: 'order-123',
      status: 'RESERVED',
      expires_at_ts: '1735000000',
      version: '1',
    });
  });

  it('POST /orders/complete -> happy path', async () => {
    ordersServiceMock.complete.mockResolvedValue({
      order_id: 'order-123',
      status: 'COMPLETED',
      completed_at_ts: '1735000099',
      version: '2',
    });

    const res = await request(app.getHttpServer())
      .post('/orders/complete')
      .send({
        idempotency_key: 'idem-complete-1',
        order_id: 'order-123',
        expected_version: 1,
        cashier_id: 'cashier-1',
      })
      .expect(201);

    expect(res.body).toEqual({
      order_id: 'order-123',
      status: 'COMPLETED',
      completed_at_ts: '1735000099',
      version: '2',
    });
  });

  it('POST /orders/cancel -> happy path', async () => {
    ordersServiceMock.cancel.mockResolvedValue({
      order_id: 'order-123',
      status: 'CANCELLED',
      version: '2',
    });

    const res = await request(app.getHttpServer())
      .post('/orders/cancel')
      .send({
        idempotency_key: 'idem-cancel-1',
        order_id: 'order-123',
        expected_version: 1,
        reason: 'user_request',
      })
      .expect(201);

    expect(res.body).toEqual({
      order_id: 'order-123',
      status: 'CANCELLED',
      version: '2',
    });
  });

  it('POST /orders/reserve -> cashier blocked in foreign office', async () => {
    policyMock.ensureCanReserveOrder.mockImplementation(() => {
      throw new ForbiddenException({ error: { code: 'OFFICE_SCOPE_VIOLATION', message: 'blocked' } });
    });

    await request(app.getHttpServer())
      .post('/orders/reserve')
      .set('x-role', 'cashier')
      .send({
        idempotency_key: 'idem-reserve-2',
        office_id: 'office-foreign',
        quote_id: 'quote-1',
        side: 'BUY',
        give: { amount: '100.00', currency: { code: 'USDT', network: 'TRC20' } },
        get: { amount: '3550.00', currency: { code: 'THB', network: 'cash' } },
      })
      .expect(403);
  });

  it('POST /orders/cancel -> cashier blocked without active shift', async () => {
    shiftsMock.hasActiveShift.mockReturnValue(false);

    const res = await request(app.getHttpServer())
      .post('/orders/cancel')
      .set('x-role', 'cashier')
      .send({
        idempotency_key: 'idem-cancel-2',
        order_id: 'order-123',
        expected_version: 1,
        reason: 'user_request',
      })
      .expect(403);

    expect(res.body.error.code).toBe('SHIFT_NOT_OPEN');
  });

  it('POST /orders/reserve -> validation failed', async () => {
    const res = await request(app.getHttpServer())
      .post('/orders/reserve')
      .send({
        idempotency_key: 'idem-reserve-3',
      })
      .expect(400);

    expect(res.body.error.code).toBe('VALIDATION_FAILED');
  });
});
