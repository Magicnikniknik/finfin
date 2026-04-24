# Nest BFF starter templates

Copy these files into a NestJS project to bootstrap:

- global exception envelope filter
- current tenant/client decorators
- orders module/service/controller gRPC bridge
- baseline e2e controller smoke spec

## Quick transfer into a real Nest project

```bash
chmod +x templates/nest-bff/apply.sh
./templates/nest-bff/apply.sh /path/to/your/nest-bff
```

Recommended order after copy:

1. verify `src/orders/orders.module.ts` gRPC client options
2. wire envs:
   - `ORDER_GRPC_URL`
   - `ORDER_GRPC_PACKAGE`
   - `ORDER_GRPC_PROTO_PATH`
3. adjust DTO imports in `src/orders/orders.controller.ts` to your actual dto files
4. run the e2e skeleton in `test/orders.e2e-spec.ts`
5. replace mocked service path with real gRPC-backed module tests

See also:
- `templates/nest-bff/POST_COPY_CHECKLIST.md` (quick verification list)

## Typical dev dependencies for the e2e skeleton

```json
{
  "devDependencies": {
    "@nestjs/testing": "^10.0.0",
    "supertest": "^7.0.0",
    "@types/supertest": "^6.0.0",
    "jest": "^29.0.0",
    "ts-jest": "^29.0.0"
  }
}
```
