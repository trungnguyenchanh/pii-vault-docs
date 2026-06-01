---
title: 6. Admin Log Monitoring Portal
description: Detailed breakdown for building the admin portal (React SPA + Go API) for log viewing, dashboards, alerts/investigation, and four-eyes approval.
---

The Admin Portal is the operations UI for the Audit (M6) and Monitoring (M7) layers: view logs, analyze stats, handle alerts, investigate incidents, and manage roles/four-eyes approvals. Stack: **React SPA + Go API**.

:::caution[Core principle]
The admin portal does NOT query real PII and does NOT display PII values. It works only with access metadata (audit) and alerts. Every admin action is self-audited.
:::

## Component architecture

```text
[React SPA Admin]  --HTTPS/JSON-->  [Go Admin API]
      |  (JWT from IdP/SSO, role=admin/dpo/security)
      |                                   |
      |                                   +--> [M6 Audit store (read-only)]
      |                                   +--> [M7 Detection (alerts)]
      |                                   +--> [M5 Access Control (roles, 4-eyes)]
      v                                   +--> writes self-audit for every action
[Recharts/ECharts]                        (Go API is the only layer touching data)
```

## Four function groups

| Code | Group | Summary |
|---|---|---|
| F1 | Log view & search | Log table, filters, keyset pagination, detail, verify-chain |
| F2 | Stats dashboard | Access-over-time charts, top actors, DENY rate |
| F3 | Alerts & investigation | Alert list, subject timeline, incident grouping, containment |
| F4 | Roles & four-eyes | Role management, approve/reject four-eyes requests |

## Database (DB) breakdown

| Task | Content | Priority |
|---|---|---|
| DB-ADM-01 | Log filter indexes: (ts), (actor,ts), (subject_ref,ts), (result,ts) | P0 |
| DB-ADM-02 | Materialized view aggregated hourly/daily for dashboard | P1 |
| DB-ADM-03 | `alert` table (M7) + index by status/severity/ts | P0 |
| DB-ADM-04 | `incident` table grouping alerts | P1 |
| DB-ADM-05 | `approval_request` table for four-eyes flow | P0 |
| DB-ADM-06 | `admin_audit` table (self-audit) | P0 |
| DB-ADM-07 | Read-replica/read-only credentials for Admin API | P1 |
| DB-ADM-08 | Optimize keyset pagination at > 100M rows | P1 |

```sql
TABLE alert (
  alert_id    UUID PRIMARY KEY,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  rule        TEXT NOT NULL,       -- BULK_READ|OFF_HOURS|...
  severity    TEXT NOT NULL,       -- LOW|MEDIUM|HIGH
  actor       TEXT NOT NULL,
  window_from TIMESTAMPTZ, window_to TIMESTAMPTZ,
  evidence    JSONB,               -- related audit seq list
  status      TEXT NOT NULL DEFAULT 'open',
  assignee    TEXT, incident_id UUID NULL
);
CREATE INDEX ON alert(status, severity, created_at DESC);
```

:::tip[Performance]
Use keyset pagination (`WHERE seq < :cursor ORDER BY seq DESC LIMIT n`) instead of OFFSET. The dashboard reads from a materialized view, not the raw log table each time.
:::

## API (Go) breakdown

Suggested: chi-router or gin, pgx/sqlc, JWT middleware via JWKS. The Go API is the only layer touching data.

| Task | Content | Priority |
|---|---|---|
| API-ADM-01 | Service skeleton: router, middleware, config | P0 |
| API-ADM-02 | Authn middleware (JWT/JWKS) + admin RBAC | P0 |
| API-ADM-03 | `GET /logs`: filter + keyset pagination | P0 |
| API-ADM-05 | `POST /logs/verify-chain` | P1 |
| API-ADM-06 | `GET /stats/*` for dashboard | P0 |
| API-ADM-07 | `GET/PATCH /alerts` | P0 |
| API-ADM-08 | `GET /subjects/{ref}/timeline` | P0 |
| API-ADM-10 | `POST /actors/{id}/revoke` (calls M5, needs 4-eyes) | P1 |
| API-ADM-11 | `GET/POST /approvals` (four-eyes) | P0 |
| API-ADM-13 | Write self-audit for every write op | P0 |

```text
GET /api/v1/logs?actor=&result=&from=&to=&cursor=<seq>&limit=50
200 -> { items:[{seq,ts,actor,action,subject_ref,field,purpose,result}],
         next_cursor:<seq|null> }

POST /api/v1/logs/verify-chain { from_seq, to_seq }
200 -> { ok:true } | { ok:false, broken_at:<seq> }

GET /api/v1/stats/access?from=&to=&bucket=hour|day
GET /api/v1/alerts?status=&severity=&cursor=&limit=
PATCH /api/v1/alerts/{id} { status, assignee }
GET /api/v1/subjects/{ref}/timeline
POST /api/v1/approvals/{id}/approve   // approver != requester
```

Suggested source layout:

```text
/cmd/adminapi/main.go    -- bootstrap, router
/internal/auth           -- JWT/JWKS middleware, RBAC
/internal/handler        -- HTTP handlers F1..F4
/internal/service        -- business logic
/internal/repo           -- DB queries (pgx), read-only audit
/internal/audit          -- self-audit (hash-chain)
/internal/model          -- DTO/response
```

## Frontend (React) breakdown

Suggested: Vite + TypeScript, React Router, TanStack Query, Recharts, a UI lib (shadcn/ui or Ant Design). OIDC auth.

| Task | Content | Priority |
|---|---|---|
| FE-ADM-01 | Project skeleton: Vite+TS, router, layout | P0 |
| FE-ADM-02 | SSO/OIDC login + token refresh | P0 |
| FE-ADM-03 | API client + centralized 401/403 handling | P0 |
| FE-ADM-04 | Logs page: table + filters + keyset pagination | P0 |
| FE-ADM-06 | Dashboard: access-over-time charts | P0 |
| FE-ADM-08 | Alerts page: list + status change | P0 |
| FE-ADM-09 | Investigation: subject timeline + incident | P1 |
| FE-ADM-11 | Approvals page: approve/reject four-eyes | P0 |
| FE-ADM-13 | Role-based UI visibility (RBAC UI) | P0 |

**Screen map:**

| Screen | Path | Main components |
|---|---|---|
| Dashboard | `/` | Summary cards, access chart, top actors |
| Logs | `/logs` | Filters, log table, detail drawer |
| Alerts | `/alerts` | Alert list, severity badge |
| Investigation | `/incidents/:id` | Subject timeline, alert list |
| Approvals | `/approvals` | Four-eyes queue, approve/reject |
| Roles | `/roles` | Role × field × action matrix |

:::caution[Key notes]
- Never display real PII — only `pii_ref`, field name, purpose.
- UI RBAC is a secondary layer; real authorization is decided by the API (never trust the client).
- The Approve button is disabled if the approver is the requester (four-eyes).
- Pagination uses keyset (`next_cursor`), not OFFSET.
:::

## Skeleton code

The repo ships real skeleton code under `skeleton/`:

- `skeleton/api-go/` — Go Admin API skeleton (router, middleware, stub handlers, repo, self-audit).
- `skeleton/web-react/` — React SPA skeleton (Vite + TS, router, API client, F1–F4 pages).
- `skeleton/db/` — SQL scripts for tables/indexes/views.

See `skeleton/README.md` to run locally.

## Sprint order

| Sprint | Content |
|---|---|
| S1 | DB index/view + Go API skeleton + authn/RBAC + React skeleton |
| S2 | F1 Log view/search |
| S3 | F2 Dashboard |
| S4 | F3 Alerts & investigation |
| S5 | F4 Roles & four-eyes + self-audit |
| S6 | Hardening, export, realtime, testing & acceptance |

## Definition of Done (DoD)

- Log search correct for all filters; keyset pagination stable at scale.
- `verify-chain` correctly detects a tampered audit record.
- Dashboard figures match verification queries on the DB.
- Alerts change status, group into incidents, build subject timelines.
- Approve button blocked when approver equals requester.
- Every admin write action appears in `admin_audit`.
- No endpoint/screen leaks real PII values.
- API returns 403 when role lacks permission (even if UI hid the button).
