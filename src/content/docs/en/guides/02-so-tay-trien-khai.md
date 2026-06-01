---
title: 2. Implementation Playbook
description: Detailed per-phase how-to — tasks, tools, examples, and completion criteria.
---

The roadmap is split into six phases following the principle "visibility first, then tighten." Each phase delivers independent value.

## Roadmap overview

| Phase | Name | Core output | Depends on |
|---|---|---|---|
| 1 | Survey & data mapping | PII map + classification matrix | None |
| 2 | Centralized audit | Every PII access has an immutable trail | P1 |
| 3 | Encryption & masking | Data meaningless if stolen | P1 |
| 4 | Access control | Least-privilege, default-deny | P2, P3 |
| 5 | Monitoring & detection | Real-time alerts, drills | P2, P4 |
| 6 | Compliance & operations | Decree 13/2023 compliance evidence | P2–P5 |

## Phase 1 — Survey & data mapping

**Goal:** know exactly where PII lives before protecting it.

1. List every storage: databases, shared files, logs, backups, partner data.
2. Run automated PII discovery for name, phone, email, address, ID fields.
3. Classify each field by sensitivity and decide the measure: encryption, masking, or both.
4. Draw the data flow: where PII enters, which systems read it, where it flows out.
5. Map obligations under Decree 13/2023.
6. Build a data processing register.

:::tip[Done when]
You have a complete PII map, classification matrix, data flow diagram, and processing register — confirmed by both business and engineering.
:::

## Phase 2 — Centralized audit (top priority)

**Goal:** achieve visibility immediately. This phase delivers the most value for leak prevention.

1. Design the audit record schema: time, identity, action, record id, purpose, result.
2. Choose tamper-resistance: append-only + hash-chaining.
3. Store audit separately from the business DB.
4. Integrate audit recording at the app / data access layer.
5. Bind purpose to read requests.
6. Build an observation dashboard and basic alerts.
7. Verify integrity (chain verify).

```sql
CREATE TABLE pii_audit (
  id           BIGSERIAL PRIMARY KEY,
  ts           TIMESTAMPTZ NOT NULL DEFAULT now(),
  actor        TEXT NOT NULL,
  action       TEXT NOT NULL,
  subject_ref  TEXT NOT NULL,
  field        TEXT,
  purpose      TEXT NOT NULL,
  result       TEXT NOT NULL,
  prev_hash    TEXT,
  row_hash     TEXT NOT NULL
);
-- INSERT only; revoke UPDATE/DELETE for all app roles
```

:::tip[Done when]
Every main PII read path records audit; the hash chain verifies; the dashboard shows real-time access and raises alerts on anomalies.
:::

## Phase 3 — Encryption & masking

**Goal:** make data meaningless if stolen, and limit staff from seeing full PII.

1. Enable TDE for at-rest encryption.
2. Deploy dedicated key management (KMS/HSM or HashiCorp Vault), separate from data and DBA privileges.
3. Apply column-level encryption to sensitive fields; use blind index for exact lookups.
4. Define Dynamic Data Masking policies by role.
5. Remove PII from logs, error messages, URLs.
6. Test on staging, measure performance, then roll out to production by table groups.

```sql
phone_enc    BYTEA   -- ciphertext
phone_bidx   TEXT    -- = HMAC(key, normalize(phone))
WHERE phone_bidx = hmac_lookup('0901234567')
```

:::caution[Known limitation]
Blind index supports only exact match, not fuzzy or range search. If complex search over encrypted data is needed, that is when to consider searchable encryption / a full vault.
:::

## Phase 4 — Access control

**Goal:** move from "everyone can read" to "least-privilege, default-deny".

1. Design RBAC based on real job roles.
2. Apply least-privilege and default-deny.
3. Enable Row-Level Security for row scope.
4. Require purpose for every PII access.
5. Apply four-eyes approval for high-risk operations.
6. Revoke existing broad privileges.

| Role | Readable fields | Scope (RLS) | Four-eyes? |
|---|---|---|---|
| CS Agent | Name, phone (partly masked) | Assigned customers | On bulk read |
| Order Ops | Name, address, phone | Orders in shift | On file export |
| Marketing | Blind index / segments only | All (anonymized) | Always (export) |
| Data Admin | On demand, with purpose | All | Always |

## Phase 5 — Monitoring, detection & drills

**Goal:** detect early and respond fast.

1. Analyze audit to set a baseline of normal access.
2. Define anomaly detection rules.
3. Wire alerts into ops channels (SIEM/Slack/on-call).
4. Build an incident response process.
5. Run periodic drills with simulated leak scenarios.
6. Verify traceability of a single record.

:::tip[Most valuable outcome]
Time to investigate a suspected leak drops from weeks to minutes — because actor, time, and purpose are already in the audit.
:::

## Phase 6 — Compliance, audit & continuous operations

**Goal:** turn technical effort into compliance evidence, sustained over time.

1. Compile Decree 13/2023 compliance evidence.
2. Establish a DSAR (data subject request) process.
3. Engage third-party pentest.
4. Run periodic access reviews.
5. Operate the encryption key lifecycle (rotation/revocation).
6. Update the data map when new fields/systems appear.
