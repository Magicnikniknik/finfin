# NestJS BFF hardening pack

This add-on complements `docs/nest_bff_plan.md` with three production-facing pieces:

1. global mobile-safe HTTP error envelope
2. metadata extraction helper (avoid repeating `@Headers(...)`)
3. minimal e2e smoke matrix for BFF endpoints

---

## 1) GlobalExceptionFilter (mobile-safe envelope)

Target response shape:

```json
{
  "error": {
    "code": "FAILED_PRECONDITION",
    "message": "quote expired"
  }
}
```

Use one global filter to normalize every exception:

```ts
import {
  ArgumentsHost,
  Catch,
  ExceptionFilter,
  HttpException,
  HttpStatus,
} from '@nestjs/common';

@Catch()
export class GlobalExceptionFilter implements ExceptionFilter {
  catch(exception: unknown, host: ArgumentsHost): void {
    const ctx = host.switchToHttp();
    const response = ctx.getResponse();

    if (exception instanceof HttpException) {
      const status = exception.getStatus();
      const payload = exception.getResponse() as any;
      const code = payload?.error?.code ?? inferCodeByStatus(status);
      const message = payload?.error?.message ?? payload?.message ?? exception.message;
      response.status(status).json({ error: { code, message } });
      return;
    }

    response
      .status(HttpStatus.INTERNAL_SERVER_ERROR)
      .json({ error: { code: 'INTERNAL_ERROR', message: 'Internal server error' } });
  }
}

function inferCodeByStatus(status: number): string {
  switch (status) {
    case HttpStatus.BAD_REQUEST:
      return 'INVALID_ARGUMENT';
    case HttpStatus.UNAUTHORIZED:
      return 'UNAUTHENTICATED';
    case HttpStatus.NOT_FOUND:
      return 'NOT_FOUND';
    case HttpStatus.CONFLICT:
      return 'CONFLICT';
    case HttpStatus.UNPROCESSABLE_ENTITY:
      return 'FAILED_PRECONDITION';
    default:
      return 'UPSTREAM_ERROR';
  }
}
```

Register in `main.ts`:

```ts
app.useGlobalFilters(new GlobalExceptionFilter());
```

---

## 2) Metadata extraction helper

To avoid repeating header extraction in every controller method:

```ts
import { createParamDecorator, ExecutionContext } from '@nestjs/common';

export const TenantAndClient = createParamDecorator(
  (_: unknown, ctx: ExecutionContext) => {
    const req = ctx.switchToHttp().getRequest();
    return {
      tenantId: req.headers['x-tenant-id'] as string | undefined,
      clientRef: req.headers['x-client-ref'] as string | undefined,
    };
  },
);
```

Controller usage:

```ts
@Post('reserve')
reserve(@TenantAndClient() meta: { tenantId?: string; clientRef?: string }, @Body() body: ReserveOrderDto) {
  return this.ordersService.reserve(meta.tenantId ?? '', meta.clientRef ?? '', body);
}
```

Alternative: use a guard/interceptor that validates required headers early and stores normalized metadata on request context.

---

## 3) Minimal BFF e2e smoke matrix

Run HTTP tests against Nest while upstream gRPC server is running:

1. `POST /orders/reserve` happy path -> `200`
2. `POST /orders/complete` happy path -> `200`
3. `POST /orders/cancel` happy path -> `200`
4. missing `x-tenant-id` -> `401` envelope `{ error: { code: "UNAUTHENTICATED", ... } }`
5. quote not found path -> `404`
6. OCC conflict path -> `409`

Validate JSON shape stability (mobile-safe):
- response has top-level `error` on failures
- `error.code` is stable and machine-friendly
- `error.message` is user/log friendly
