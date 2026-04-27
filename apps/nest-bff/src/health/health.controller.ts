import { Controller, Get } from '@nestjs/common';

@Controller()
export class HealthController {
  @Get()
  index() {
    return {
      service: 'finfin-nest-bff',
      status: 'ok',
      endpoints: ['/healthz', '/readyz'],
    };
  }

  @Get('healthz')
  healthz() {
    return { status: 'ok' };
  }

  @Get('readyz')
  readyz() {
    return { status: 'ready' };
  }
}
