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

  switch (err.code) {
    case GrpcStatus.INVALID_ARGUMENT:
      throw new BadRequestException(buildBody('INVALID_ARGUMENT', message));
    case GrpcStatus.UNAUTHENTICATED:
      throw new UnauthorizedException(buildBody('UNAUTHENTICATED', message));
    case GrpcStatus.NOT_FOUND:
      throw new NotFoundException(buildBody('NOT_FOUND', message));
    case GrpcStatus.ALREADY_EXISTS:
      throw new ConflictException(buildBody('ALREADY_EXISTS', message));
    case GrpcStatus.ABORTED:
      throw new ConflictException(buildBody('VERSION_CONFLICT', message));
    case GrpcStatus.RESOURCE_EXHAUSTED:
      throw new ConflictException(buildBody('INSUFFICIENT_LIQUIDITY', message));
    case GrpcStatus.FAILED_PRECONDITION:
      throw new UnprocessableEntityException(buildBody('FAILED_PRECONDITION', message));
    default:
      throw new InternalServerErrorException(buildBody('UPSTREAM_ERROR', message));
  }
}

function buildBody(code: string, message: string) {
  return { error: { code, message } };
}
