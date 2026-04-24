import { Observable } from 'rxjs';

export type CurrencyRef = {
  code: string;
  network: string;
};

export type Money = {
  amount: string;
  currency: CurrencyRef;
};

export type ReserveOrderRequest = {
  idempotency_key: string;
  office_id: string;
  quote_id: string;
  side: 'BUY' | 'SELL';
  give: Money;
  get: Money;
};

export type ReserveOrderResponse = {
  order_id: string;
  status: string;
  expires_at_ts: string | number;
  version: string | number;
};

export type CompleteOrderRequest = {
  idempotency_key: string;
  order_id: string;
  expected_version: number;
  cashier_id: string;
};

export type CompleteOrderResponse = {
  order_id: string;
  status: string;
  completed_at_ts: string | number;
  version: string | number;
};

export type CancelOrderRequest = {
  idempotency_key: string;
  order_id: string;
  expected_version: number;
  reason: string;
};

export type CancelOrderResponse = {
  order_id: string;
  status: string;
  version: string | number;
};

export interface OrderGrpcService {
  ReserveOrder(data: ReserveOrderRequest, metadata?: unknown): Observable<ReserveOrderResponse>;
  CompleteOrder(data: CompleteOrderRequest, metadata?: unknown): Observable<CompleteOrderResponse>;
  CancelOrder(data: CancelOrderRequest, metadata?: unknown): Observable<CancelOrderResponse>;
}
