import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { endpoints, type LogQuery } from '../api/endpoints';

// F1 — Xem & tìm kiếm log: bộ lọc + bảng + phân trang keyset + verify-chain.
export default function Logs() {
  const [filter, setFilter] = useState<LogQuery>({ limit: 50 });
  const [cursor, setCursor] = useState<number | undefined>(undefined);

  const q = useQuery({
    queryKey: ['logs', filter, cursor],
    queryFn: () => endpoints.listLogs({ ...filter, cursor }),
  });

  async function onVerify() {
    const items = q.data?.items ?? [];
    if (items.length === 0) return;
    const from = items[items.length - 1].seq;
    const to = items[0].seq;
    try {
      const res = await endpoints.verifyChain(from, to);
      alert(res.ok ? 'Chuỗi hash hợp lệ ✓' : `Phát hiện sửa đổi tại seq ${res.broken_at}`);
    } catch {
      alert('verify-chain chưa được hiện thực ở API.');
    }
  }

  return (
    <div>
      <h1>Nhật ký truy cập</h1>

      <div className="filters">
        <input placeholder="actor" onChange={(e) => setFilter((f) => ({ ...f, actor: e.target.value }))} />
        <select onChange={(e) => setFilter((f) => ({ ...f, result: e.target.value }))}>
          <option value="">— result —</option>
          <option value="ALLOW">ALLOW</option>
          <option value="DENY">DENY</option>
        </select>
        <input type="datetime-local" onChange={(e) => setFilter((f) => ({ ...f, from: e.target.value }))} />
        <input type="datetime-local" onChange={(e) => setFilter((f) => ({ ...f, to: e.target.value }))} />
        <button onClick={() => setCursor(undefined)}>Lọc</button>
        <button onClick={onVerify} style={{ background: '#475569', borderColor: '#475569' }}>
          Xác minh chuỗi
        </button>
      </div>

      {q.isError && <p className="error">Không tải được log (API chưa hiện thực).</p>}

      <table>
        <thead>
          <tr>
            <th>seq</th><th>thời gian</th><th>actor</th><th>action</th>
            <th>subject_ref</th><th>field</th><th>purpose</th><th>result</th>
          </tr>
        </thead>
        <tbody>
          {(q.data?.items ?? []).map((r) => (
            <tr key={r.seq}>
              <td>{r.seq}</td>
              <td>{new Date(r.ts).toLocaleString()}</td>
              <td>{r.actor}</td>
              <td>{r.action}</td>
              <td>{r.subject_ref}</td>
              <td>{r.field}</td>
              <td>{r.purpose}</td>
              <td>{r.result}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <div style={{ marginTop: 12 }}>
        <button
          disabled={!q.data?.next_cursor}
          onClick={() => setCursor(q.data?.next_cursor ?? undefined)}
        >
          Tải thêm
        </button>
      </div>
    </div>
  );
}
