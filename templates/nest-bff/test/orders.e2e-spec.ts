import { INestApplication, ValidationPipe } from '@nestjs/common';
import { Test, TestingModule } from '@nestjs/testing';
import request from 'supertest';

import { GlobalExceptionFilter } from '../src/common/filters/global-exception.filter';
import { OrdersController } from '../src/orders/orders.controller';
import { OrdersService } from '../src/orders/orders.service';

describe('OrdersController (e2e)', () => {
  let app: INestApplication;

  const ordersServiceMock = {
    reserve: jest.fn(),
    complete: jest.fn(),
    cancel: jest.fn(),
  };

  beforeAll(async () => {
    const moduleRef: TestingModule = await Test.createTestingModule({
      controllers: [OrdersController],
      providers: [
        {
          provide: OrdersService,
          useValue: ordersServiceMock,
        },
      ],
    }).compile();

    app = moduleRef.createNestApplication();

    app.useGlobalPipes(
      new ValidationPipe({
        whitelist: true,
        transform: true,
      }),
    );

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
      .set('x-tenant-id', 'tenant-1')
      .set('x-client-ref', 'client-1')
      .send(payload)
      .expect(201);

    expect(res.body).toEqual({
      order_id: 'order-123',
      status: 'RESERVED',
      expires_at_ts: '1735000000',
      version: '1',
    });
  });

  it('POST /orders/reserve -> missing tenant header', async () => {
    const payload = {
      idempotency_key: 'idem-reserve-2',
      office_id: 'office-1',
      quote_id: 'quote-1',
      side: 'BUY',
      give: { amount: '100.00', currency: { code: 'USDT', network: 'TRC20' } },
      get: { amount: '3550.00', currency: { code: 'THB', network: 'cash' } },
    };

    const res = await request(app.getHttpServer())
      .post('/orders/reserve')
      .set('x-client-ref', 'client-1')
      .send(payload)
      .expect(401);

    expect(res.body).toEqual({
      error: { code: 'MISSING_TENANT', message: 'x-tenant-id header is required' },
    });
  });
});
