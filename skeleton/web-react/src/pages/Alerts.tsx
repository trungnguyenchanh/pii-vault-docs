import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { endpoints } from '../api/endpoints';

// F3 — Cảnh báo: danh sách + đổi trạng thái + mở điều tra.
export default function Alerts() {
  const qc = useQueryClient();
  const nav = useNavigate();
  const q = useQuery({ queryKey: ['alerts'], queryFn: () => endpoints.listAlerts('open') });

  const update = useMutation({
    mutationFn: (v: { id: string; status: string }) => endpoints.updateAlert(v.id, v.status),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['alerts'] }),
  });

  return (
    <div>
      <h1>Cảnh báo</h1>
      {q.isError && <p className="error">Không tải được cảnh báo (API chưa hiện thực).</p>}
      <table>
        <thead>
          <tr><th>thời gian</th><th>luật</th><th>mức</th><th>actor</th><th>trạng thái</th><th></th></tr>
        </thead>
        <tbody>
          {(q.data?.items ?? []).map((a) => (
            <tr key={a.alert_id}>
              <td>{new Date(a.created_at).toLocaleString()}</td>
              <td>{a.rule}</td>
              <td><span className={`badge ${a.severity}`}>{a.severity}</span></td>
              <td>{a.actor}</td>
              <td>{a.status}</td>
              <td style={{ display: 'flex', gap: 6 }}>
                <button onClick={() => update.mutate({ id: a.alert_id, status: 'investigating' })}>
                  Điều tra
                </button>
                <button style={{ background: '#475569', borderColor: '#475569' }}
                  onClick={() => nav(`/incidents/${a.alert_id}`)}>
                  Timeline
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
