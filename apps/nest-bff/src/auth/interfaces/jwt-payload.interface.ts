export type UserRole = 'owner' | 'manager' | 'cashier';

export interface JwtPayload {
  sub: string;
  tenant_id: string;
  office_id?: string | null;
  scope_office_ids?: string[];
  role: UserRole;
  login: string;
}
