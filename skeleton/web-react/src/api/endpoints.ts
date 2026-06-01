import { api } from './client';

// ---- Types (khớp với DTO của Go API) ----
export interface LogEntry {
  seq: number;
  ts: string;
  actor: string;
  action: string;
  subject_ref: string;
  field: string;
  purpose: string;
  result: 'ALLOW' | 'DENY' | 'APPROVAL';
}
export interface LogPage {
  items: LogEntry[];
  next_cursor: number | null;
}
export interface AccessPoint {
  bucket: string;
  allow_cnt: number;
  deny_cnt: number;
  actors: number;
}
export interface Summary {
  total: number;
  allow: number;
  deny: number;
  distinct_actors: number;
  distinct_subjects: number;
}
export interface Alert {
  alert_id: string;
  created_at: string;
  rule: string;
  severity: 'LOW' | 'MEDIUM' | 'HIGH';
  actor: string;
  status: string;
  assignee?: string;
}
export interface TimelineEntry {
  ts: string;
  actor: string;
  action: string;
  purpose: string;
  result: string;
}
export interface ApprovalRequest {
  request_id: string;
  requester: string;
  action: string;
  purpose: string;
  status: string;
  created_at: string;
}

// ---- Endpoint functions ----
export interface LogQuery {
  actor?: string;
  result?: string;
  from?: string;
  to?: string;
  cursor?: number;
  limit?: number;
}

function qs(params: Record<string, unknown>): string {
  const sp = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== null && v !== '') sp.set(k, String(v));
  });
  const s = sp.toString();
  return s ? `?${s}` : '';
}

export const endpoints = {
  // F1
  listLogs: (q: LogQuery) => api.get<LogPage>(`/logs${qs(q)}`),
  verifyChain: (from_seq: number, to_seq: number) =>
    api.post<{ ok: boolean; broken_at?: number }>('/logs/verify-chain', { from_seq, to_seq }),
  // F2
  statsAccess: (from?: string, to?: string, bucket = 'hour') =>
    api.get<{ series: AccessPoint[] }>(`/stats/access${qs({ from, to, bucket })}`),
  statsSummary: (from?: string, to?: string) =>
    api.get<Summary>(`/stats/summary${qs({ from, to })}`),
  // F3
  listAlerts: (status?: string, severity?: string) =>
    api.get<{ items: Alert[] }>(`/alerts${qs({ status, severity })}`),
  updateAlert: (id: string, status: string, assignee?: string) =>
    api.patch<{ ok: boolean }>(`/alerts/${id}`, { status, assignee }),
  subjectTimeline: (ref: string) =>
    api.get<{ items: TimelineEntry[] }>(`/subjects/${ref}/timeline`),
  // F4
  listApprovals: (status = 'pending') =>
    api.get<{ items: ApprovalRequest[] }>(`/approvals${qs({ status })}`),
  approve: (id: string) => api.post<{ ok: boolean }>(`/approvals/${id}/approve`),
  reject: (id: string, reason: string) =>
    api.post<{ ok: boolean }>(`/approvals/${id}/reject`, { reason }),
};
