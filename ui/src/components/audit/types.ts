export interface AuditLog {
  id: number;
  user_id: number;
  action: string;
  resource: string;
  detail: string;
  ip: string;
  created_at: string;
}