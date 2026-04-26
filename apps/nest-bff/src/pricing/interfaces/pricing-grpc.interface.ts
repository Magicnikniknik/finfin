import { Metadata } from '@grpc/grpc-js';
import { Observable } from 'rxjs';

export type CalculateQuoteGrpcRequest = {
  officeId: string;
  giveCurrencyId: string;
  getCurrencyId: string;
  inputMode: 'GIVE' | 'GET';
  amount: string;
};

export type QuoteMoneyGrpc = {
  amount: string;
  currencyId: string;
};

export type CalculateQuoteGrpcResponse = {
  quoteId: string;
  side: 'BUY' | 'SELL' | 'QUOTE_SIDE_UNSPECIFIED';
  give: QuoteMoneyGrpc;
  get: QuoteMoneyGrpc;
  fixedRate: string;
  baseRate: string;
  feeAmount: string;
  expiresAtTs: string;
  sourceName: string;
};

export type PricingGrpcService = {
  CalculateQuote(
    payload: CalculateQuoteGrpcRequest,
    metadata?: Metadata,
  ): Observable<CalculateQuoteGrpcResponse>;
};
