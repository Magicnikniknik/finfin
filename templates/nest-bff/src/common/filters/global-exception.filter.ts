import {
  ArgumentsHost,
  BadRequestException,
  Catch,
  ExceptionFilter,
  HttpException,
  HttpStatus,
  UnauthorizedException,
} from '@nestjs/common';
import { Request, Response } from 'express';

type FieldError = {
  field: string;
  code: string;
  message: string;
};

type ErrorBody = {
  error: {
    code: string;
    message: string;
    field_errors?: FieldError[];
    request_id?: string;
  };
};

@Catch()
export class GlobalExceptionFilter implements ExceptionFilter {
  catch(exception: unknown, host: ArgumentsHost): void {
    const ctx = host.switchToHttp();
    const response = ctx.getResponse<Response>();
    const request = ctx.getRequest<Request>();

    const requestId =
      (request.headers['x-request-id'] as string | undefined) ??
      (request.headers['x-correlation-id'] as string | undefined);

    const { statusCode, body } = this.normalizeException(exception, requestId);
    response.status(statusCode).json(body);
  }

  private normalizeException(
    exception: unknown,
    requestId?: string,
  ): { statusCode: number; body: ErrorBody } {
    if (exception instanceof HttpException) {
      return this.fromHttpException(exception, requestId);
    }

    return {
      statusCode: HttpStatus.INTERNAL_SERVER_ERROR,
      body: {
        error: {
          code: 'INTERNAL_ERROR',
          message: 'Internal server error',
          ...(requestId ? { request_id: requestId } : {}),
        },
      },
    };
  }

  private fromHttpException(
    exception: HttpException,
    requestId?: string,
  ): { statusCode: number; body: ErrorBody } {
    const statusCode = exception.getStatus();
    const raw = exception.getResponse();

    if (this.isAlreadyWrapped(raw)) {
      const wrapped = raw as ErrorBody;
      if (requestId && !wrapped.error.request_id) {
        wrapped.error.request_id = requestId;
      }
      return { statusCode, body: wrapped };
    }

    if (exception instanceof BadRequestException) {
      return { statusCode, body: this.fromBadRequest(raw, requestId) };
    }

    if (exception instanceof UnauthorizedException) {
      return {
        statusCode,
        body: {
          error: {
            code: 'UNAUTHORIZED',
            message: this.extractMessage(raw, 'Unauthorized'),
            ...(requestId ? { request_id: requestId } : {}),
          },
        },
      };
    }

    return {
      statusCode,
      body: {
        error: {
          code: this.defaultCodeByStatus(statusCode),
          message: this.extractMessage(raw, exception.message),
          ...(requestId ? { request_id: requestId } : {}),
        },
      },
    };
  }

  private fromBadRequest(raw: unknown, requestId?: string): ErrorBody {
    const fieldErrors = this.extractFieldErrors(raw);
    if (fieldErrors.length > 0) {
      return {
        error: {
          code: 'VALIDATION_FAILED',
          message: 'Request validation failed',
          field_errors: fieldErrors,
          ...(requestId ? { request_id: requestId } : {}),
        },
      };
    }

    return {
      error: {
        code: 'BAD_REQUEST',
        message: this.extractMessage(raw, 'Bad request'),
        ...(requestId ? { request_id: requestId } : {}),
      },
    };
  }

  private extractFieldErrors(raw: unknown): FieldError[] {
    if (!raw || typeof raw !== 'object') return [];
    const maybeMessage = (raw as { message?: unknown }).message;
    if (!Array.isArray(maybeMessage)) return [];

    return maybeMessage.map((item) => {
      if (typeof item === 'string') {
        return { field: 'request', code: 'INVALID', message: item };
      }
      if (item && typeof item === 'object') {
        const obj = item as Partial<FieldError>;
        return {
          field: obj.field ?? 'request',
          code: obj.code ?? 'INVALID',
          message: obj.message ?? 'Invalid value',
        };
      }
      return { field: 'request', code: 'INVALID', message: 'Invalid value' };
    });
  }

  private extractMessage(raw: unknown, fallback: string): string {
    if (typeof raw === 'string') return raw;
    if (raw && typeof raw === 'object') {
      const obj = raw as { message?: unknown; error?: unknown };
      if (typeof obj.message === 'string') return obj.message;
      if (Array.isArray(obj.message) && typeof obj.message[0] === 'string') return obj.message[0];
      if (typeof obj.error === 'string') return obj.error;
    }
    return fallback;
  }

  private defaultCodeByStatus(statusCode: number): string {
    switch (statusCode) {
      case HttpStatus.BAD_REQUEST:
        return 'BAD_REQUEST';
      case HttpStatus.UNAUTHORIZED:
        return 'UNAUTHORIZED';
      case HttpStatus.FORBIDDEN:
        return 'FORBIDDEN';
      case HttpStatus.NOT_FOUND:
        return 'NOT_FOUND';
      case HttpStatus.CONFLICT:
        return 'CONFLICT';
      case HttpStatus.UNPROCESSABLE_ENTITY:
        return 'FAILED_PRECONDITION';
      default:
        return 'HTTP_ERROR';
    }
  }

  private isAlreadyWrapped(raw: unknown): raw is ErrorBody {
    if (!raw || typeof raw !== 'object') return false;
    const maybeError = (raw as { error?: unknown }).error;
    return !!maybeError && typeof maybeError === 'object';
  }
}
