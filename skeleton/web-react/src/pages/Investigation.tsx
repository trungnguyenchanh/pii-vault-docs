import { useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { endpoints } from '../api/endpoints';

// F3 — Điều tra: dựng timeline truy cập của một subject (hoặc liên quan alert).
export default function Investigation() {
  const { id } = useParams();
  // Trong thực tế, từ alert -> subject_ref liên quan. Skeleton dùng id trực tiếp.
  const q = useQuery({
    queryKey: ['timeline', id],
    queryFn: () => endpoints.subjectTimeline(id ?? ''),
    enabled: !!id,
  });

  return (
    <div>
      <h1>Điều tra: {id}</h1>
      {q.isError && <p className="error">Không tải được timeline (API chưa hiện thực).</p>}
      <table>
        <thead><tr><th>thời gian</th><th>actor</th><th>action</th><th>purpose</th><th>result</th></tr></thead>
        <tbody>
          {(q.data?.items ?? []).map((t, i) => (
            <tr key={i}>
              <td>{new Date(t.ts).toLocaleString()}</td>
              <td>{t.actor}</td><td>{t.action}</td><td>{t.purpose}</td><td>{t.result}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
