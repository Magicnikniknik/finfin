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
import {
  CancelOrderRequest,
  CancelOrderResponse,
  CompleteOrderRequest,
  CompleteOrderResponse,
  OrderGrpcService,
  ReserveOrderRequest,
  ReserveOrderResponse,
} from './interfaces/order-grpc.interface';

@Injectable()
export class OrdersService implements OnModuleInit {
  private ordersGrpc!: OrderGrpcService;

  constructor(
    @Inject('ORDER_PACKAGE') private readonly grpcClient: ClientGrpc,
  ) {}

  onModuleInit() {
    this.ordersGrpc =
      this.grpcClient.getService<OrderGrpcService>('OrderService');
  }

  async reserve(
    tenantId: string,
    clientRef: string,
    payload: ReserveOrderRequest,
  ): Promise<ReserveOrderResponse> {
    const metadata = buildMetadata(tenantId, clientRef);

    try {
      return await lastValueFrom(
        this.ordersGrpc.ReserveOrder(payload, metadata),
      );
    } catch (error) {
      mapGrpcErrorToHttp(error);
    }
  }

  async complete(
    tenantId: string,
    clientRef: string,
    payload: CompleteOrderRequest,
  ): Promise<CompleteOrderResponse> {
    const metadata = buildMetadata(tenantId, clientRef);

    try {
      return await lastValueFrom(
        this.ordersGrpc.CompleteOrder(payload, metadata),
      );
    } catch (error) {
      mapGrpcErrorToHttp(error);
    }
  }

  async cancel(
    tenantId: string,
    clientRef: string,
    payload: CancelOrderRequest,
  ): Promise<CancelOrderResponse> {
    const metadata = buildMetadata(tenantId, clientRef);

    try {
      return await lastValueFrom(
        this.ordersGrpc.CancelOrder(payload, metadata),
      );
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
