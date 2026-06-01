---
title: 5. System Analysis & Task Plan
description: Work breakdown by role (BA, Dev, DB, QC, DevOps, Security) for planning tasks.
---

## Conventions

- **FR-xx** = functional requirement; **NFR-xx** = non-functional; **US-xx** = user story.
- Suggested task code: `[Role]-[Module]-[number]`, e.g. `DEV-M2-01`, `DB-M6-02`.
- Priority: **P0** (mandatory, blocking) · **P1** (high) · **P2** (nice to have).

## Audience

| Role | Use for | Main sections |
|---|---|---|
| BA | Finalize requirements, stories, AC | Requirements, BA |
| Tech / Dev | Break module/API tasks | Dev, API |
| DB | Schema, migration, RLS | DB |
| QC | Test plan, test cases | QC |
| DevOps | Infra, CI/CD, HA, KMS | DevOps |
| Security | Threat model, controls | Security |

## Functional requirements (FR)

| Code | Description | Module |
|---|---|---|
| FR-01 | Create subject and generate pii_ref | M1 |
| FR-02 | Read PII via gateway with purpose | M2,M3,M4 |
| FR-03 | Write/update PII, encrypt on store | M2,M3 |
| FR-04 | Exact lookup (blind index) | M3 |
| FR-05 | Per-field encryption; key rotation | M3 |
| FR-06 | Mask data by role | M4 |
| FR-07 | RBAC + RLS + purpose check | M5 |
| FR-08 | Four-eyes approval | M5 |
| FR-09 | Immutable audit logging | M6 |
| FR-10 | Verify audit chain integrity | M6 |
| FR-11 | Anomaly detection & alerting | M7 |
| FR-12 | DSAR per Decree 13/2023 | M8 |
| FR-13 | Erase via crypto-shred | M3,M8 |
| FR-14 | Export compliance evidence | M6,M8 |

## Non-functional requirements (NFR)

| Code | Description | Metric |
|---|---|---|
| NFR-01 | Gateway performance | p99 added latency < ~40ms |
| NFR-02 | Data scale | Stable at > 1M subjects |
| NFR-03 | Availability | HA; defined RTO/RPO |
| NFR-04 | Audit integrity | No lost records; tamper detection |
| NFR-05 | Key security | Keys separate from data & DBA |
| NFR-06 | Data residency | Within Decree 13/2023 scope |
| NFR-08 | Traceability | Reconstruct one record's history in minutes |

## Work breakdown by role

### Business Analyst (BA)

| Task | Content | Priority |
|---|---|---|
| BA-01 | Valid purpose catalog | P0 |
| BA-02 | Roles & permission matrix | P0 |
| BA-03 | Masking rules | P1 |
| BA-04 | Four-eyes rules | P1 |
| BA-05 | DSAR flow & Decree 13/2023 deadlines | P0 |
| BA-06 | Acceptance criteria (AC) for FRs | P0 |

### Tech / Developer

| Task | Content | Priority |
|---|---|---|
| DEV-M6-01 | Audit service: Append + hash-chain | P0 |
| DEV-M2-01 | Gateway skeleton: authn, routing | P0 |
| DEV-M2-02 | Reveal/Store/Update/Lookup | P0 |
| DEV-M3-01 | Envelope encryption, AES-GCM | P0 |
| DEV-M3-02 | Blind index + normalization | P1 |
| DEV-M5-01 | Authorize (RBAC+purpose), default-deny | P0 |
| DEV-M5-02 | RLS hook + four-eyes workflow | P1 |
| DEV-M4-01 | Policy-driven masking engine | P1 |
| DEV-M7-01 | Detector + alerting | P1 |
| DEV-M8-01 | DSAR orchestrator + ExportEvidence | P1 |
| DEV-MIG-01 | Dual-write + backfill tooling | P1 |

:::tip[Suggested order]
Start with DEV-M6-01 and DEV-M2-01 — establish visibility early; then M3 → M5 → M4 → M7 → M8.
:::

### Database (DB)

| Task | Content | Priority |
|---|---|---|
| DB-M1-01 | Schema subject, subject_field + index | P0 |
| DB-M6-01 | Schema pii_audit; revoke UPDATE/DELETE | P0 |
| DB-M6-02 | Separate audit store instance/credentials | P0 |
| DB-M5-01 | Schema role, role_grant, purpose | P0 |
| DB-M5-02 | RLS policies from claims | P1 |
| DB-ENC-01 | Enable TDE; test encrypted-column perf | P0 |
| DB-PERF-01 | Optimize indexes at > 1M records | P1 |
| DB-MIG-01 | Migration + rollback + reconciliation | P1 |

### Quality Control (QC)

| Code | Scenario | Expected |
|---|---|---|
| TC-AUTH-1 | Reveal without purpose | 403 DENY + audit |
| TC-RBAC-1 | Role lacks read permission | 403 DENY |
| TC-RLS-1 | Read subject out of scope | 403 DENY |
| TC-4EYES-1 | Bulk reveal not approved | 202 PENDING |
| TC-MASK-1 | CS agent reads phone | Returns masked value |
| TC-AUD-1 | Manually edit an audit row | VerifyChain reports broken |
| TC-SHRED-1 | DSAR erase | Subsequent decryption fails |
| TC-PERF-1 | 1000 reveals/sec | p99 within NFR-01 |

### DevOps / SRE

| Task | Content | Priority |
|---|---|---|
| OPS-01 | Stand up KMS/HSM or Vault (VN residency) | P0 |
| OPS-02 | HA infra for M2 | P0 |
| OPS-03 | DB cluster + separate audit store | P0 |
| OPS-04 | CI/CD + security scanning | P1 |
| OPS-05 | Secret/mTLS management, auto-rotation | P0 |
| OPS-07 | Encrypted backup & recovery drills | P1 |
| OPS-08 | Test fail-closed behavior | P0 |

### Security

| Task | Content | Priority |
|---|---|---|
| SEC-01 | System-wide threat model | P0 |
| SEC-02 | Review crypto design (nonce, KEK/DEK) | P0 |
| SEC-03 | Review authz model & four-eyes | P0 |
| SEC-04 | Gateway hardening (input, rate-limit, SSRF) | P1 |
| SEC-05 | PII-out-of-log/error/URL policy | P0 |
| SEC-06 | Third-party pentest plan | P1 |
| SEC-07 | Incident response & notification process | P1 |

## Suggested phase planning

| Phase | Main work cluster | Lead roles |
|---|---|---|
| P1 | Survey, BA finalize purpose/roles, DB base schema | BA, DB |
| P2 | Audit (M6) + gateway skeleton (M2) | Dev, DB, DevOps |
| P3 | Encryption (M3), masking (M4), KMS, migration | Dev, DB, DevOps |
| P4 | Access control (M5), RLS, four-eyes | Dev, Security, BA |
| P5 | Monitoring (M7), alerts, drills | Dev, DevOps, Security |
| P6 | DSAR (M8), compliance, pentest, access review | BA, Security, DPO |

:::tip[Control gates]
After each phase, QC accepts against criteria and Security signs off P0 items before moving on.
:::
