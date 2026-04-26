import {
  Inject,
  Injectable,
  OnModuleInit,
  UnauthorizedException,
} from '@nestjs/common';
import { Metadata } from '@grpc/grpc-js';
import { ClientGrpc } from '@nestjs/microservices';
import { lastValueFrom } from 'rxjs';

import { mapGrpcErrorToHttp } from '../common/grpc-http-error.mapper';
import { CalculateQuoteDto } from './dto/calculate-quote.dto';
import {
  CalculateQuoteGrpcResponse,
  PricingGrpcService,
} from './interfaces/pricing-grpc.interface';

type CalculateQuoteHttpResponse = {
  quote_id: string;
  side: string;
  give: {
    amount: string;
    currency_id: string;
  };
  get: {
    amount: string;
    currency_id: string;
  };
  fixed_rate: string;
  base_rate: string;
  fee_amount: string;
  expires_at_ts: string;
  source_name: string;
};

@Injectable()
export class PricingService implements OnModuleInit {
  private pricingGrpc!: PricingGrpcService;

  constructor(
    @Inject('PRICING_PACKAGE') private readonly grpcClient: ClientGrpc,
  ) {}

  onModuleInit() {
    this.pricingGrpc =
      this.grpcClient.getService<PricingGrpcService>('PricingService');
  }

  async calculate(
    tenantId: string,
    clientRef: string,
    payload: CalculateQuoteDto,
  ): Promise<CalculateQuoteHttpResponse> {
    const metadata = buildMetadata(tenantId, clientRef);

    const grpcPayload = {
      officeId: payload.office_id,
      giveCurrencyId: payload.give_currency_id,
      getCurrencyId: payload.get_currency_id,
      inputMode: payload.input_mode,
      amount: payload.amount,
    };

    try {
      const response = await lastValueFrom(
        this.pricingGrpc.CalculateQuote(grpcPayload, metadata),
      );
      return mapGrpcResponseToHttp(response);
    } catch (error) {
      mapGrpcErrorToHttp(error);
    }
  }
}

function buildMetadata(tenantId?: string, clientRef?: string): Metadata {
  if (!tenantId) {
    throw new UnauthorizedException({
      error: {
        code: 'MISSING_TENANT',
        message: 'x-tenant-id is required',
      },
    });
  }

  if (!clientRef) {
    throw new UnauthorizedException({
      error: {
        code: 'MISSING_CLIENT_REF',
        message: 'x-client-ref is required',
      },
    });
  }

  const metadata = new Metadata();
  metadata.set('x-tenant-id', tenantId);
  metadata.set('x-client-ref', clientRef);
  return metadata;
}

function mapGrpcResponseToHttp(
  response: CalculateQuoteGrpcResponse,
): CalculateQuoteHttpResponse {
  return {
    quote_id: response.quoteId,
    side: response.side,
    give: {
      amount: response.give.amount,
      currency_id: response.give.currencyId,
    },
    get: {
      amount: response.get.amount,
      currency_id: response.get.currencyId,
    },
    fixed_rate: response.fixedRate,
    base_rate: response.baseRate,
    fee_amount: response.feeAmount,
    expires_at_ts: response.expiresAtTs,
    source_name: response.sourceName,
  };
}
