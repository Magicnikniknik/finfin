import {
  BadRequestException,
  ConflictException,
  InternalServerErrorException,
  NotFoundException,
  UnauthorizedException,
  UnprocessableEntityException,
} from '@nestjs/common';
import { status as GrpcStatus } from '@grpc/grpc-js';

type GrpcLikeError = {
  code?: number;
  details?: string;
  message?: string;
};

export function mapGrpcErrorToHttp(error: unknown): never {
  const err = (error ?? {}) as GrpcLikeError;
  const message = err.details || err.message || 'Unknown upstream error';
  const normalized = message.toLowerCase();

  switch (err.code) {
    case GrpcStatus.INVALID_ARGUMENT:
      if (normalized.includes('invalid quote input')) {
        throw new BadRequestException(buildBody('INVALID_QUOTE_INPUT', message));
      }
      throw new BadRequestException(buildBody('INVALID_ARGUMENT', message));
    case GrpcStatus.UNAUTHENTICATED:
      throw new UnauthorizedException(buildBody('UNAUTHENTICATED', message));
    case GrpcStatus.NOT_FOUND:
      if (normalized.includes('base rate not found')) {
        throw new NotFoundException(buildBody('BASE_RATE_NOT_FOUND', message));
      }
      if (normalized.includes('no margin rule found')) {
        throw new NotFoundException(buildBody('NO_MARGIN_RULE_FOUND', message));
      }
      throw new NotFoundException(buildBody('NOT_FOUND', message));
    case GrpcStatus.ALREADY_EXISTS:
      throw new ConflictException(buildBody('ALREADY_EXISTS', message));
    case GrpcStatus.ABORTED:
      throw new ConflictException(buildBody('VERSION_CONFLICT', message));
    case GrpcStatus.RESOURCE_EXHAUSTED:
      throw new ConflictException(buildBody('INSUFFICIENT_LIQUIDITY', message));
    case GrpcStatus.FAILED_PRECONDITION:
      if (normalized.includes('rate stale')) {
        throw new UnprocessableEntityException(buildBody('RATE_STALE', message));
      }
      if (normalized.includes('rate guardrail triggered')) {
        throw new UnprocessableEntityException(buildBody('RATE_GUARDRAIL_TRIGGERED', message));
      }
      throw new UnprocessableEntityException(buildBody('FAILED_PRECONDITION', message));
    default:
      throw new InternalServerErrorException(buildBody('UPSTREAM_ERROR', message));
  }
}

function buildBody(code: string, message: string) {
  return { error: { code, message } };
}
