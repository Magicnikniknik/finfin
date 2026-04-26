import { IsIn, IsNotEmpty, IsString } from 'class-validator';

export class CalculateQuoteDto {
  @IsString()
  @IsNotEmpty()
  office_id!: string;

  @IsString()
  @IsNotEmpty()
  give_currency_id!: string;

  @IsString()
  @IsNotEmpty()
  get_currency_id!: string;

  @IsIn(['GIVE', 'GET'])
  input_mode!: 'GIVE' | 'GET';

  @IsString()
  @IsNotEmpty()
  amount!: string;
}
