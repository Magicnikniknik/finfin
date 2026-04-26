import {
  INestApplication,
  UnauthorizedException,
  ValidationPipe,
} from '@nestjs/common';
import { Test, TestingModule } from '@nestjs/testing';
import request from 'supertest';

import { AuditService } from '../src/audit/audit.service';
import { AuthController } from '../src/auth/auth.controller';
import { AuthService } from '../src/auth/auth.service';

import { GlobalExceptionFilter } from '../src/common/filters/global-exception.filter';

describe('AuthController (e2e)', () => {
  let app: INestApplication;

  const authServiceMock = {
    login: jest.fn(),
    refresh: jest.fn(),
    me: jest.fn(),
  };

  const auditMock = { record: jest.fn() };

  beforeAll(async () => {
    const moduleRef: TestingModule = await Test.createTestingModule({
      controllers: [AuthController],
      providers: [
        {
          provide: AuthService,
          useValue: authServiceMock,
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

    app.useGlobalFilters(new GlobalExceptionFilter());
    await app.init();
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  afterAll(async () => {
    await app.close();
  });

  it('POST /auth/login -> success', async () => {
    authServiceMock.login.mockResolvedValue({
      access_token: 'access-1',
      refresh_token: 'refresh-1',
      user: {
        id: 'u-1',
        tenant_id: '11111111-1111-1111-1111-111111111111',
        office_id: null,
        role: 'owner',
        login: 'owner_demo',
        display_name: 'Owner Demo',
      },
    });

    const res = await request(app.getHttpServer())
      .post('/auth/login')
      .send({
        tenant_id: '11111111-1111-1111-1111-111111111111',
        login: 'owner_demo',
        password: 'owner_demo_password',
      })
      .expect(201);

    expect(res.body.access_token).toBe('access-1');
    expect(res.body.refresh_token).toBe('refresh-1');
  });

  it('POST /auth/login -> invalid password', async () => {
    authServiceMock.login.mockRejectedValue(
      new UnauthorizedException({
        error: {
          code: 'INVALID_CREDENTIALS',
          message: 'Invalid login or password',
        },
      }),
    );

    const res = await request(app.getHttpServer())
      .post('/auth/login')
      .send({
        tenant_id: '11111111-1111-1111-1111-111111111111',
        login: 'owner_demo',
        password: 'wrong_password',
      })
      .expect(401);

    expect(res.body.error.code).toBe('INVALID_CREDENTIALS');
  });

  it('POST /auth/login -> disabled user', async () => {
    authServiceMock.login.mockRejectedValue(
      new UnauthorizedException({
        error: {
          code: 'USER_DISABLED',
          message: 'User is disabled',
        },
      }),
    );

    const res = await request(app.getHttpServer())
      .post('/auth/login')
      .send({
        tenant_id: '11111111-1111-1111-1111-111111111111',
        login: 'owner_demo',
        password: 'owner_demo_password',
      })
      .expect(401);

    expect(res.body.error.code).toBe('USER_DISABLED');
  });

  it('POST /auth/refresh -> success', async () => {
    authServiceMock.refresh.mockResolvedValue({
      access_token: 'access-next',
      refresh_token: 'refresh-next',
      user: {
        id: 'u-1',
        tenant_id: '11111111-1111-1111-1111-111111111111',
        office_id: '22222222-2222-2222-2222-222222222222',
        role: 'manager',
        login: 'manager_demo',
        display_name: 'Manager Demo',
      },
    });

    const res = await request(app.getHttpServer())
      .post('/auth/refresh')
      .send({
        refresh_token: 'refresh-1-token-long-enough',
      })
      .expect(201);

    expect(res.body.access_token).toBe('access-next');
  });


  it('POST /auth/refresh -> invalid token', async () => {
    authServiceMock.refresh.mockRejectedValue(
      new UnauthorizedException({
        error: {
          code: 'INVALID_REFRESH_TOKEN',
          message: 'Refresh token is invalid',
        },
      }),
    );

    const res = await request(app.getHttpServer())
      .post('/auth/refresh')
      .send({
        refresh_token: 'invalid-token',
      })
      .expect(401);

    expect(res.body.error.code).toBe('INVALID_REFRESH_TOKEN');
  });

  it('POST /auth/refresh -> revoked token', async () => {
    authServiceMock.refresh.mockRejectedValue(
      new UnauthorizedException({
        error: {
          code: 'REFRESH_TOKEN_REVOKED',
          message: 'Refresh token is revoked',
        },
      }),
    );

    const res = await request(app.getHttpServer())
      .post('/auth/refresh')
      .send({
        refresh_token: 'refresh-1-token-long-enough',
      })
      .expect(401);

    expect(res.body.error.code).toBe('REFRESH_TOKEN_REVOKED');
  });

  it('POST /auth/refresh -> expired token', async () => {
    authServiceMock.refresh.mockRejectedValue(
      new UnauthorizedException({
        error: {
          code: 'REFRESH_TOKEN_EXPIRED',
          message: 'Refresh token is expired',
        },
      }),
    );

    const res = await request(app.getHttpServer())
      .post('/auth/refresh')
      .send({
        refresh_token: 'refresh-1-token-long-enough',
      })
      .expect(401);

    expect(res.body.error.code).toBe('REFRESH_TOKEN_EXPIRED');
  });
});
