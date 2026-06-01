// F4 — Phân quyền: xem/sửa ma trận role × field × action.
// Skeleton hiển thị ma trận tĩnh; nối API GET/PUT /roles khi hiện thực (API-ADM-12).
const ROWS = [
  { role: 'CSKH', fields: 'Tên, SĐT (che)', scope: 'Khách được phân công', fourEyes: 'Khi đọc hàng loạt' },
  { role: 'Vận hành đơn', fields: 'Tên, địa chỉ, SĐT', scope: 'Đơn trong ca', fourEyes: 'Khi xuất file' },
  { role: 'Marketing', fields: 'Blind index / phân khúc', scope: 'Toàn bộ (ẩn danh)', fourEyes: 'Luôn (xuất tệp)' },
  { role: 'Quản trị dữ liệu', fields: 'Theo yêu cầu, có mục đích', scope: 'Toàn bộ', fourEyes: 'Luôn' },
];

export default function Roles() {
  return (
    <div>
      <h1>Phân quyền</h1>
      <p style={{ color: '#6b7280' }}>
        Ma trận tham chiếu. Khi nối API, cho phép sửa grant và lưu lại (có self-audit).
      </p>
      <table>
        <thead><tr><th>Vai trò</th><th>Trường được đọc</th><th>Phạm vi (RLS)</th><th>Bốn mắt?</th></tr></thead>
        <tbody>
          {ROWS.map((r) => (
            <tr key={r.role}>
              <td><b>{r.role}</b></td><td>{r.fields}</td><td>{r.scope}</td><td>{r.fourEyes}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
