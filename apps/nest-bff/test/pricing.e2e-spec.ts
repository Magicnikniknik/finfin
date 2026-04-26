import {
  ForbiddenException,
  INestApplication,
  NotFoundException,
  UnprocessableEntityException,
  ValidationPipe,
} from '@nestjs/common';
import { Test, TestingModule } from '@nestjs/testing';
import request = require('supertest');

import { JwtAuthGuard } from '../src/auth/jwt-auth.guard';
import { RolesGuard } from '../src/auth/guards/roles.guard';

import { GlobalExceptionFilter } from '../src/common/filters/global-exception.filter';
import { AuditService } from '../src/audit/audit.service';
import { OperatorPolicyService } from '../src/common/policies/operator-policy.service';
import { PricingController } from '../src/pricing/pricing.controller';
import { PricingService } from '../src/pricing/pricing.service';

describe('PricingController (e2e)', () => {
  let app: INestApplication;

  const pricingServiceMock = {
    calculate: jest.fn(),
  };

  const policyMock = { ensureCanCalculateQuote: jest.fn() };
  const auditMock = { record: jest.fn() };

  const jwtGuardMock = { canActivate: jest.fn(() => true) };
  const rolesGuardMock = { canActivate: jest.fn(() => true) };

  beforeAll(async () => {
    const moduleRef: TestingModule = await Test.createTestingModule({
      controllers: [PricingController],
      providers: [
        { provide: JwtAuthGuard, useValue: jwtGuardMock },
        { provide: RolesGuard, useValue: rolesGuardMock },
        {
          provide: PricingService,
          useValue: pricingServiceMock,
        },
        {
          provide: OperatorPolicyService,
          useValue: policyMock,
        },
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

    app.use((req: any, _res, next) => {
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

  it('POST /quotes/calculate -> success', async () => {
    pricingServiceMock.calculate.mockResolvedValue({
      quote_id: 'q-123',
      side: 'SELL',
      give: { amount: '100.00', currency_id: 'cur-usdt' },
      get: { amount: '3550.00', currency_id: 'cur-thb' },
      fixed_rate: '35.500000000000000000',
      base_rate: '35.450000000000000000',
      fee_amount: '0.100000000000000000',
      expires_at_ts: '1735000000',
      source_name: 'provider-1',
    });

    const res = await request(app.getHttpServer())
      .post('/quotes/calculate')
      .send({
        office_id: 'office-1',
        give_currency_id: 'cur-usdt',
        get_currency_id: 'cur-thb',
        input_mode: 'GIVE',
        amount: '100.00',
      })
      .expect(201);

    expect(res.body.quote_id).toBe('q-123');
  });

  it('POST /quotes/calculate -> cashier blocked in foreign office', async () => {
    policyMock.ensureCanCalculateQuote.mockImplementation(() => {
      throw new ForbiddenException({ error: { code: 'OFFICE_SCOPE_VIOLATION', message: 'blocked' } });
    });

    await request(app.getHttpServer())
      .post('/quotes/calculate')
      .set('x-role', 'cashier')
      .send({
        office_id: 'office-foreign',
        give_currency_id: 'cur-usdt',
        get_currency_id: 'cur-thb',
        input_mode: 'GIVE',
        amount: '100.00',
      })
      .expect(403);
  });

  it('POST /quotes/calculate -> invalid dto', async () => {
    const res = await request(app.getHttpServer())
      .post('/quotes/calculate')
      .send({ office_id: 'office-1' })
      .expect(400);

    expect(res.body.error.code).toBe('VALIDATION_FAILED');
  });

  it('POST /quotes/calculate -> base rate not found', async () => {
    pricingServiceMock.calculate.mockRejectedValue(
      new NotFoundException({
        error: {
          code: 'BASE_RATE_NOT_FOUND',
          message: 'base rate not found',
        },
      }),
    );

    const res = await request(app.getHttpServer())
      .post('/quotes/calculate')
      .send({
        office_id: 'office-1',
        give_currency_id: 'cur-usdt',
        get_currency_id: 'cur-thb',
        input_mode: 'GIVE',
        amount: '100.00',
      })
      .expect(404);

    expect(res.body.error.code).toBe('BASE_RATE_NOT_FOUND');
  });

  it('POST /quotes/calculate -> no margin rule', async () => {
    pricingServiceMock.calculate.mockRejectedValue(
      new NotFoundException({
        error: {
          code: 'NO_MARGIN_RULE_FOUND',
          message: 'no margin rule found',
        },
      }),
    );

    const res = await request(app.getHttpServer())
      .post('/quotes/calculate')
      .send({
        office_id: 'office-1',
        give_currency_id: 'cur-usdt',
        get_currency_id: 'cur-thb',
        input_mode: 'GIVE',
        amount: '100.00',
      })
      .expect(404);

    expect(res.body.error.code).toBe('NO_MARGIN_RULE_FOUND');
  });

  it('POST /quotes/calculate -> stale rate', async () => {
    pricingServiceMock.calculate.mockRejectedValue(
      new UnprocessableEntityException({
        error: {
          code: 'RATE_STALE',
          message: 'rate stale',
        },
      }),
    );

    const res = await request(app.getHttpServer())
      .post('/quotes/calculate')
      .send({
        office_id: 'office-1',
        give_currency_id: 'cur-usdt',
        get_currency_id: 'cur-thb',
        input_mode: 'GIVE',
        amount: '100.00',
      })
      .expect(422);

    expect(res.body.error.code).toBe('RATE_STALE');
  });
});
