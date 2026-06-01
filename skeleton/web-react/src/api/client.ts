// Lớp API client tập trung: gắn token, xử lý 401/403, gọi Go Admin API.
const BASE = import.meta.env.VITE_API_BASE ?? '/api/v1';

// TODO(FE-ADM-02): lấy access token thật từ OIDC/SSO.
// Skeleton: dùng header debug để chạy với Go API ở dev mode.
function authHeaders(): Record<string, string> {
  const token = sessionStorage.getItem('access_token');
  if (token) return { Authorization: `Bearer ${token}` };
  // dev fallback
  return { 'X-Debug-User': 'dev-admin', 'X-Debug-Roles': 'admin,dpo,security' };
}

export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...authHeaders(),
      ...(init?.headers ?? {}),
    },
  });

  if (res.status === 401) {
    // TODO(FE-ADM-03): điều hướng về /login (OIDC).
    throw new ApiError(401, 'Chưa đăng nhập hoặc phiên hết hạn');
  }
  if (res.status === 403) {
    throw new ApiError(403, 'Không đủ quyền truy cập');
  }
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new ApiError(res.status, text || `Lỗi ${res.status}`);
  }
  return (await res.json()) as T;
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: body ? JSON.stringify(body) : undefined }),
  patch: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'PATCH', body: body ? JSON.stringify(body) : undefined }),
};
