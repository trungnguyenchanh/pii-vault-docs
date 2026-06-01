import { useQuery } from '@tanstack/react-query';
import {
  LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, Legend,
} from 'recharts';
import { endpoints } from '../api/endpoints';

// F2 — Dashboard thống kê: thẻ tổng quan + biểu đồ access theo thời gian.
export default function Dashboard() {
  const summary = useQuery({ queryKey: ['summary'], queryFn: () => endpoints.statsSummary() });
  const access = useQuery({ queryKey: ['access'], queryFn: () => endpoints.statsAccess() });

  return (
    <div>
      <h1>Dashboard</h1>

      <div className="cards">
        <Card label="Tổng truy cập" value={summary.data?.total} />
        <Card label="ALLOW" value={summary.data?.allow} />
        <Card label="DENY" value={summary.data?.deny} />
        <Card label="Số actor" value={summary.data?.distinct_actors} />
        <Card label="Số subject" value={summary.data?.distinct_subjects} />
      </div>

      <div className="card">
        <h3>Truy cập theo thời gian</h3>
        {access.isError && <p className="error">Không tải được dữ liệu (API chưa hiện thực).</p>}
        <ResponsiveContainer width="100%" height={320}>
          <LineChart data={access.data?.series ?? []}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="bucket" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="allow_cnt" name="ALLOW" stroke="#2e7d32" />
            <Line type="monotone" dataKey="deny_cnt" name="DENY" stroke="#b91c1c" />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

function Card({ label, value }: { label: string; value?: number }) {
  return (
    <div className="card">
      <div style={{ color: '#6b7280', fontSize: 13 }}>{label}</div>
      <div style={{ fontSize: 26, fontWeight: 700 }}>{value ?? '—'}</div>
    </div>
  );
}
