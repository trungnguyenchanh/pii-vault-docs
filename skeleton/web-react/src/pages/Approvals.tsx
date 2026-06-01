import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { endpoints } from '../api/endpoints';

// F4 — Phê duyệt bốn mắt: duyệt/từ chối yêu cầu.
// LƯU Ý: nút Approve bị chặn nếu người duyệt trùng người yêu cầu (API kiểm tra thật).
const CURRENT_USER = 'dev-admin'; // TODO(FE-ADM-02): lấy từ phiên OIDC.

export default function Approvals() {
  const qc = useQueryClient();
  const q = useQuery({ queryKey: ['approvals'], queryFn: () => endpoints.listApprovals('pending') });

  const approve = useMutation({
    mutationFn: (id: string) => endpoints.approve(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['approvals'] }),
    onError: () => alert('Không duyệt được (API chưa hiện thực hoặc vi phạm bốn mắt).'),
  });
  const reject = useMutation({
    mutationFn: (id: string) => endpoints.reject(id, 'không hợp lệ'),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['approvals'] }),
  });

  return (
    <div>
      <h1>Phê duyệt bốn mắt</h1>
      {q.isError && <p className="error">Không tải được hàng chờ (API chưa hiện thực).</p>}
      <table>
        <thead>
          <tr><th>thời gian</th><th>người yêu cầu</th><th>thao tác</th><th>mục đích</th><th></th></tr>
        </thead>
        <tbody>
          {(q.data?.items ?? []).map((a) => {
            const selfRequest = a.requester === CURRENT_USER; // bốn mắt
            return (
              <tr key={a.request_id}>
                <td>{new Date(a.created_at).toLocaleString()}</td>
                <td>{a.requester}</td>
                <td>{a.action}</td>
                <td>{a.purpose}</td>
                <td style={{ display: 'flex', gap: 6 }}>
                  <button disabled={selfRequest} title={selfRequest ? 'Không thể tự duyệt yêu cầu của mình' : ''}
                    onClick={() => approve.mutate(a.request_id)}>
                    Duyệt
                  </button>
                  <button style={{ background: '#b91c1c', borderColor: '#b91c1c' }}
                    onClick={() => reject.mutate(a.request_id)}>
                    Từ chối
                  </button>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
