import { IsNotEmpty, IsOptional, IsString, IsUUID } from 'class-validator';

export class OpenShiftDto {
  @IsUUID()
  @IsNotEmpty()
  office_id!: string;

  @IsOptional()
  @IsString()
  note?: string;
}
