-- =====================================================================
-- Admin Portal — DB schema bổ sung (PostgreSQL)
-- Phục vụ trang admin theo dõi & phân tích log PII.
-- Chạy SAU khi đã có bảng pii_audit (M6) từ hệ thống lõi.
-- =====================================================================

-- ---------------------------------------------------------------------
-- DB-ADM-01: Index phục vụ lọc log nhanh
-- ---------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_audit_ts          ON pii_audit (ts DESC);
CREATE INDEX IF NOT EXISTS idx_audit_actor_ts    ON pii_audit (actor, ts DESC);
CREATE INDEX IF NOT EXISTS idx_audit_subject_ts  ON pii_audit (subject_ref, ts DESC);
CREATE INDEX IF NOT EXISTS idx_audit_result_ts   ON pii_audit (result, ts DESC);

-- ---------------------------------------------------------------------
-- DB-ADM-02: Materialized view tổng hợp theo giờ cho dashboard
-- Refresh định kỳ (vd mỗi 5 phút) bằng job nền hoặc pg_cron.
-- ---------------------------------------------------------------------
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_access_hourly AS
SELECT
  date_trunc('hour', ts)                              AS bucket,
  count(*)                                            AS total_cnt,
  count(*) FILTER (WHERE result = 'ALLOW')            AS allow_cnt,
  count(*) FILTER (WHERE result = 'DENY')             AS deny_cnt,
  count(DISTINCT actor)                               AS distinct_actors,
  count(DISTINCT subject_ref)                         AS distinct_subjects
FROM pii_audit
GROUP BY 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_access_hourly_bucket
  ON mv_access_hourly (bucket);

-- Lệnh refresh (đặt vào job nền):
--   REFRESH MATERIALIZED VIEW CONCURRENTLY mv_access_hourly;

-- ---------------------------------------------------------------------
-- DB-ADM-03: Bảng cảnh báo (do M7 Detection sinh ra)
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS alert (
  alert_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  rule         TEXT NOT NULL,                 -- BULK_READ | OFF_HOURS | HIGH_DENY | SCATTERED | EXPORT
  severity     TEXT NOT NULL,                 -- LOW | MEDIUM | HIGH
  actor        TEXT NOT NULL,
  window_from  TIMESTAMPTZ,
  window_to    TIMESTAMPTZ,
  evidence     JSONB,                         -- { audit_seqs: [...], count: N }
  status       TEXT NOT NULL DEFAULT 'open',  -- open | ack | investigating | closed
  assignee     TEXT,
  incident_id  UUID
);
CREATE INDEX IF NOT EXISTS idx_alert_status_sev_ts
  ON alert (status, severity, created_at DESC);

-- ---------------------------------------------------------------------
-- DB-ADM-04: Vụ điều tra gom nhiều cảnh báo
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS incident (
  incident_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title        TEXT NOT NULL,
  status       TEXT NOT NULL DEFAULT 'open',  -- open | investigating | closed
  opened_by    TEXT NOT NULL,
  opened_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  closed_at    TIMESTAMPTZ,
  summary      TEXT
);

-- ---------------------------------------------------------------------
-- DB-ADM-05: Yêu cầu phê duyệt bốn mắt
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS approval_request (
  request_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  requester    TEXT NOT NULL,
  action       TEXT NOT NULL,                 -- BULK_REVEAL | REVOKE | ...
  scope        JSONB,                         -- mô tả phạm vi thao tác
  purpose      TEXT NOT NULL,
  status       TEXT NOT NULL DEFAULT 'pending', -- pending | approved | rejected | expired
  approver     TEXT,
  decided_at   TIMESTAMPTZ,
  reason       TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  expires_at   TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_approval_status ON approval_request (status, created_at DESC);

-- ---------------------------------------------------------------------
-- DB-ADM-06: Self-audit cho thao tác trên trang admin (hash-chained)
-- ---------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS admin_audit (
  seq        BIGSERIAL PRIMARY KEY,
  ts         TIMESTAMPTZ NOT NULL DEFAULT now(),
  admin      TEXT NOT NULL,
  action     TEXT NOT NULL,                   -- VIEW_LOGS | CHANGE_ALERT | APPROVE | REVOKE | ...
  target     TEXT,
  detail     JSONB,
  prev_hash  TEXT,
  row_hash   TEXT NOT NULL
);
-- Chỉ cho INSERT; thu hồi UPDATE/DELETE với role ứng dụng:
--   REVOKE UPDATE, DELETE ON admin_audit FROM app_role;

-- ---------------------------------------------------------------------
-- (Tham chiếu) Bảng RBAC từ M5 — trang admin đọc/sửa qua API
-- ---------------------------------------------------------------------
-- TABLE role_grant (role_id TEXT, field TEXT, action TEXT,
--   PRIMARY KEY (role_id, field, action));
-- TABLE purpose_catalog (purpose TEXT PRIMARY KEY, active BOOL);
