import { NavLink, Route, Routes } from 'react-router-dom';
import Dashboard from './pages/Dashboard';
import Logs from './pages/Logs';
import Alerts from './pages/Alerts';
import Investigation from './pages/Investigation';
import Approvals from './pages/Approvals';
import Roles from './pages/Roles';

const nav = [
  { to: '/', label: 'Dashboard', end: true },
  { to: '/logs', label: 'Nhật ký' },
  { to: '/alerts', label: 'Cảnh báo' },
  { to: '/approvals', label: 'Phê duyệt' },
  { to: '/roles', label: 'Phân quyền' },
];

export default function App() {
  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="brand">PII Admin</div>
        <nav>
          {nav.map((n) => (
            <NavLink key={n.to} to={n.to} end={n.end}
              className={({ isActive }) => (isActive ? 'active' : '')}>
              {n.label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <main className="content">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/logs" element={<Logs />} />
          <Route path="/alerts" element={<Alerts />} />
          <Route path="/incidents/:id" element={<Investigation />} />
          <Route path="/approvals" element={<Approvals />} />
          <Route path="/roles" element={<Roles />} />
        </Routes>
      </main>
    </div>
  );
}
