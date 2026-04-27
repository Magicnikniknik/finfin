import { IsInt, IsNotEmpty, IsString, Min } from 'class-validator';

export class CompleteOrderDto {
  @IsString()
  @IsNotEmpty()
  idempotency_key!: string;

  @IsString()
  @IsNotEmpty()
  order_id!: string;

  @IsInt()
  @Min(0)
  expected_version!: number;

  @IsString()
  @IsNotEmpty()
  cashier_id!: string;
}
