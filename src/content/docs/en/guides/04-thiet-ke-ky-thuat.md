---
title: 4. Technical Design
description: Data schemas, API contracts, and design decisions for the eight modules M1–M8.
---

## Overall architecture

### Reference technology stack

- **Business DB & vault:** PostgreSQL (TDE, pgcrypto, native RLS). Can be SQL Server (Always Encrypted) or Oracle (TDE + Data Redaction).
- **Key management:** self-hosted HashiCorp Vault or cloud KMS/HSM with a region in Vietnam.
- **Access gateway (M2):** stateless gRPC/REST service, horizontally scaled behind a load balancer.
- **Audit & monitoring:** separate append-only store (separate PostgreSQL or OpenSearch); SIEM.
- **Identity:** IdP/SSO issuing JWT for users, mTLS for service-to-service.

### Cross-cutting principles

- Default-deny.
- Separation of duties: data, keys, and audit on three distinct trust boundaries.
- Single entry: all PII access via gateway M2.
- Full traceability: no PII operation without audit.
- Fail-closed for security.

```text
[Apps: Web/App/CRM/OMS]
        | (store only pii_ref)
        v   JWT/mTLS + purpose
[M2 PII Gateway] --(check)--> [M5 Access Control]
        | reveal/store
        v
[M3 Encryption & KMS] <---> [Vault: keys]
        |  ciphertext + blind index
        v
[Vault DB: TDE]      [M4 Masking on return]
        +--> [M6 Immutable Audit] --> [M7 Monitoring/Alert]
        +--> [M8 Compliance/DSAR]
```

## M1 — Subject Management

```sql
TABLE subject (
  pii_ref      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  status       TEXT NOT NULL DEFAULT 'active',  -- active|merged|shredded
  merged_into  UUID NULL REFERENCES subject(pii_ref),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

TABLE subject_field (
  pii_ref      UUID REFERENCES subject(pii_ref),
  field        TEXT NOT NULL,        -- phone|email|address|fullname
  value_enc    BYTEA NOT NULL,       -- AES-256-GCM ciphertext
  value_bidx   TEXT NULL,            -- HMAC blind index
  dek_id       TEXT NOT NULL,        -- key reference in Vault
  PRIMARY KEY (pii_ref, field)
);
CREATE INDEX ON subject_field(value_bidx);
```

**Decision:** `pii_ref` uses UUIDv4 (unguessable). Split `subject_field` per row so each field uses its own DEK and rotates independently.

## M2 — PII Access Gateway

```text
AuthContext { identity, roles[], auth_method(JWT|mTLS) }
RevealField(pii_ref, field, purpose) -> { value | masked_value, audit_id }
StoreSubject(fields{}, purpose)      -> { pii_ref, audit_id }
UpdateSubject(pii_ref, patch{}, purpose) -> { ok, audit_id }
LookupByIndex(field, value, purpose) -> { pii_ref | null, audit_id }
BulkReveal(pii_refs[], field, purpose)  -> { request_id, PENDING_APPROVAL }
```

**RevealField sequence:** authenticate → M5.Authorize → (DENY/PENDING/ALLOW) → M3.Decrypt → M4.Mask → M6 record audit → return.

:::tip[Performance]
Cache authorization decisions (M5) by (role, action, field); do NOT cache plaintext PII. Target gateway p99 latency under tens of ms.
:::

## M3 — Encryption & Key Management

```text
KEK (in Vault/HSM, never exported)
  └─ wrap/unwrap ─> DEK (per subject/field)
                      └─ AES-256-GCM encrypts value
value_enc    = AES_GCM(DEK, nonce, value)
value_bidx   = HMAC-SHA256(IndexKey, normalize(value))
```

**Decision:** AES-256-GCM (AEAD), unique nonce each time. Per subject/field DEK narrows blast radius. Key rotation only re-wraps DEK (no full re-encryption).

:::caution[Security warning]
Blind index reveals that two records share a value. Create it only for fields needing lookup; consider per-field salt to reduce frequency analysis risk.
:::

## M4 — Dynamic Masking

```sql
TABLE mask_policy (
  role TEXT, field TEXT,
  strategy TEXT,  -- FULL|PARTIAL|HIDE
  PRIMARY KEY (role, field)
);
-- Default when no row: HIDE
-- phone 0901234567 -> 09****4567 ; email an@mail.com -> a***@mail.com
```

**Decision:** data-driven configuration (no hard-coding). Multiple roles → apply the least-revealing strategy.

## M5 — Access Control

```sql
TABLE role_grant (role_id TEXT, field TEXT, action TEXT,
  PRIMARY KEY (role_id, field, action));
TABLE purpose_catalog (purpose TEXT PRIMARY KEY, active BOOL);
TABLE fourEyes_rule (action TEXT, condition TEXT);
```

**Authorize algorithm:** check valid purpose → check role_grant (default-deny) → apply RLS row scope → match four-eyes → ALLOW. Every branch returns a reason for M6 to record.

## M6 — Immutable Audit Log

```sql
TABLE pii_audit (
  seq BIGSERIAL PRIMARY KEY,
  ts TIMESTAMPTZ DEFAULT now(),
  actor TEXT, action TEXT, subject_ref UUID,
  field TEXT, purpose TEXT, result TEXT,
  meta JSONB, prev_hash TEXT, row_hash TEXT NOT NULL
);
row_hash = SHA256(prev_hash || ts || actor || action ||
                  subject_ref || field || purpose || result)
```

**Immutability:** revoke UPDATE/DELETE; hash-chain each INSERT; separate instance/credentials; periodic `VerifyChain()`.

:::caution[Integrity]
If audit recording fails, the PII operation should fail-closed (do not proceed without a trail).
:::

## M7 — Monitoring & Detection

| Rule | Example condition | Severity |
|---|---|---|
| Abnormal bulk read | > N subjects/min by one actor | High |
| Off-hours access | READ outside 7am–9pm | Medium |
| High DENY rate | > X% DENY in window | Medium |
| Scattered access | one actor touches many unrelated subjects | High |
| Data export | any EXPORT | High |

## M8 — Compliance & DSAR

```sql
TABLE consent (pii_ref UUID, purpose TEXT, granted BOOL, ts TIMESTAMPTZ,
  PRIMARY KEY(pii_ref, purpose));
TABLE dsar_request (request_id UUID PRIMARY KEY, pii_ref UUID,
  type TEXT, status TEXT, due_date DATE);
TABLE retention_policy (field TEXT, ttl_days INT);
```

**ERASE process:** receive & verify → M5 four-eyes → M1.ShredSubject → M3.ShredKey → mark shredded + M6 audit → issue deletion confirmation.

## Module dependencies

| Module | Depends on | Implemented in |
|---|---|---|
| M6 Audit | Independent (foundation) | P2 — first |
| M2 Gateway | M5, M3, M6 | P2 skeleton → P3 full |
| M3 Encryption | Vault/KMS | P3 |
| M1 Subject | M3, M6 | P1 base → P3 |
| M4 Masking | M5 | P3 |
| M5 Access | IdP/SSO, M6 | P4 |
| M7 Monitoring | M6 | P5 |
| M8 Compliance | M1, M3, M5, M6 | P6 |
