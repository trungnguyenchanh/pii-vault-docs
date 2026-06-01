---
title: 3. Functional Specification
description: Eight functional modules and end-to-end business flows of the PII system.
---

## Module map

| Code | Module | Primary role | Phase |
|---|---|---|---|
| M1 | Subject Management | Manage subjects & pii_ref | P1, P3 |
| M2 | PII Access Gateway | Single entry point for read/write | P2, P3 |
| M3 | Encryption & KMS | Column encryption, blind index, KMS | P3 |
| M4 | Dynamic Masking | Mask data by role | P3 |
| M5 | Access Control | RBAC, RLS, purpose, four-eyes | P4 |
| M6 | Immutable Audit | Immutable trail of every access | P2 |
| M7 | Monitoring & Detection | Real-time anomaly detection | P5 |
| M8 | Compliance & DSAR | Subject rights, evidence | P6 |

## M1 — Subject Management

**Purpose:** manage each data subject's lifecycle and the `pii_ref` used to replace PII in applications.

**Functions:** create subject, link reference, update PII, merge/split, disable/delete (crypto-shred).

**Rules:**
1. Each subject has exactly one stable `pii_ref`; apps store only `pii_ref`.
2. All operations go through gateway M2 and record audit M6.
3. Deletion is done by destroying the encryption key (ciphertext permanently meaningless).

## M2 — PII Access Gateway

**Purpose:** single entry point to read/write PII, enforcing security policy, purpose, and audit consistently.

**Functions:** Reveal (read), Store (write), Update, Lookup by blind index, call authentication.

```text
RevealField(pii_ref, field, purpose) -> { value | masked }
StoreSubject(pii_payload)            -> { pii_ref }
UpdateSubject(pii_ref, patch)        -> { ok }
LookupByIndex(field, value)          -> { pii_ref | null }
```

**Rules:** every request carries a valid purpose; missing purpose → deny. Each call creates exactly one audit record. Never return PII to logs/errors/URLs.

## M3 — Encryption & Key Management

**Purpose:** encrypt PII values, generate blind index, manage key lifecycle via KMS/HSM.

**Rules:** keys never co-located with data and not under DBA privilege. Each subject/field uses its own DEK. Blind index only serves exact match.

## M4 — Dynamic Masking

**Purpose:** mask part/all of PII by role when returning values.

**Rules:** applied after M5 determines the role, before M2 returns. Default role gets fully hidden value (safe by default).

## M5 — Access Control

**Purpose:** decide who can do what with which data, based on role, row scope, purpose, and approval.

**Functions:** RBAC, RLS, purpose check, four-eyes approval, privilege lifecycle (SSO/IdP).

**Rules:** default-deny. Bulk/sensitive operations require four-eyes. Privilege revocation takes effect immediately.

## M6 — Immutable Audit Log

**Purpose:** immutably record every PII access/operation, foundation for tracing and compliance. **The highest-value module for leak prevention.**

**Rules:** append-only, no update/delete; hash-chaining; separated from business DB; contains no real PII, only `pii_ref` and field name.

## M7 — Monitoring & Detection

**Purpose:** analyze audit in real time, detect anomalies, and alert.

**Rules:** alerts carry full context (actor, time, purpose, count). Thresholds tuned to baseline to reduce noise.

## M8 — Compliance & Subject Rights (DSAR)

**Purpose:** enforce subject rights under Decree 13/2023 and compile compliance evidence.

**Rules:** deletion via crypto-shred (M3) and audit (M6). Sensitive DSAR operations require four-eyes.

## Main business flows

### Controlled PII read flow
1. Call gateway M2 with `pii_ref` + identity + purpose.
2. M2 authenticates; M5 checks permission, scope, purpose.
3. If high-risk → M5 requires four-eyes approval.
4. M3 decrypts; M4 applies masking by role.
5. M2 returns result; M6 records immutable audit.
6. M7 analyzes; alerts on anomaly.

### Deletion (DSAR) flow
1. M8 receives and verifies requester identity.
2. M5 requires four-eyes approval.
3. M3 performs crypto-shred (destroy key).
4. M6 records audit; M8 stores processing evidence.

### Incident investigation flow
1. M7 raises an alert (e.g., abnormal read of hundreds of records).
2. The response team queries M6 by actor/time.
3. M5 revokes the suspect's privileges to contain.
4. M8 compiles the incident file and notifies if required.

:::tip[Core value]
Thanks to M6 (immutable audit) and M7 (monitoring), "where did the leak come from?" has an answer with actor, time, and purpose — instead of "unknown" today.
:::
