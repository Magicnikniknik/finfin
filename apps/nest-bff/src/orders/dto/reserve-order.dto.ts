import { Type } from 'class-transformer';
import {
  IsIn,
  IsNotEmpty,
  IsObject,
  IsString,
  ValidateNested,
} from 'class-validator';

class CurrencyDto {
  @IsString()
  @IsNotEmpty()
  code!: string;

  @IsString()
  @IsNotEmpty()
  network!: string;
}

class MoneyDto {
  @IsString()
  @IsNotEmpty()
  amount!: string;

  @IsObject()
  @ValidateNested()
  @Type(() => CurrencyDto)
  currency!: CurrencyDto;
}

export class ReserveOrderDto {
  @IsString()
  @IsNotEmpty()
  idempotency_key!: string;

  @IsString()
  @IsNotEmpty()
  office_id!: string;

  @IsString()
  @IsNotEmpty()
  quote_id!: string;

  @IsIn(['BUY', 'SELL'])
  side!: 'BUY' | 'SELL';

  @IsObject()
  @ValidateNested()
  @Type(() => MoneyDto)
  give!: MoneyDto;

  @IsObject()
  @ValidateNested()
  @Type(() => MoneyDto)
  get!: MoneyDto;
}
