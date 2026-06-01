---
title: 1. Architecture Proposal
description: Positioning a PII protection solution using the hybrid DB-native + Masking and centralized Audit/Access-Control approach.
---

## Executive summary

This document presents a pragmatic approach to protecting customer Personally Identifiable Information (PII), designed for a retail/e-commerce business with over one million customer records, focused on preventing data leaks and tracing their source when incidents occur.

Unlike building a full PII Vault (costly, high-risk, long timeline) or buying a commercial product (license cost, vendor lock-in), this hybrid approach leverages **the database's built-in security features** (encryption, masking, access control) and adds **a centralized audit and access-control layer** to address the core pain point: today nobody knows who read which customer's data, when, and why.

:::tip[Core philosophy]
"You cannot patch a leak you cannot see." The first and most important investment is **visibility** — recording every access to PII in an immutable way.
:::

:::note[Scope]
This is a defense-in-depth risk-mitigation solution, NOT a cryptographic data-in-use protection like searchable encryption. If that capability is needed later, this hybrid approach remains a solid foundation to evolve toward a full vault without starting over.
:::

## The problem

Three observed symptoms share one root cause:

- **Exposed and scattered data:** PII (name, phone, email, address) stored in plaintext, scattered across CRM, order DB, logs, manual Excel exports, backups, partner services.
- **No access log:** When a leak occurs, there is no evidence to identify the source. Every investigation is "groping in the dark."
- **No purpose-based access control:** Too many people and services can read almost everything; no record of why a record was accessed.

**Legal implication:** Decree 13/2023/ND-CP on personal data protection requires processing data for the correct purpose, with control and traceability. The lack of logs and access control makes compliance hard to demonstrate during audits.

## Hybrid architecture overview

The solution has two complementary pillars.

### Pillar A — DB-native + Masking (data layer)

- **TDE (Transparent Data Encryption):** encrypts all data at rest.
- **Column / Field-level Encryption:** encrypts sensitive columns with separately managed keys.
- **Dynamic Data Masking:** masks data by role at query time.
- **Row-Level Security (RLS):** restricts each role to rows in its scope.

### Pillar B — Centralized Audit & Access-Control layer

The biggest differentiator. Every query touching PII is recorded by a central layer, leaving an immutable trail:

- **Access gateway / Data Access Layer:** apps read PII via a shared service instead of querying sensitive tables directly.
- **Purpose binding:** every PII read request must include a reason.
- **Immutable audit log (append-only, hash-chained):** each access records who · what · when · why · result.
- **RBAC + least-privilege, default-deny:** sensitive operations require "four-eyes" approval.
- **Real-time anomaly detection & alerting.**

## Why this hybrid approach

| Criterion | Build Vault | Buy product | Hybrid |
|---|---|---|---|
| Upfront cost | High | Medium–high | Low |
| Time to deploy | Slow (12+ months) | Medium | Fast (per quarter) |
| Crypto error risk | High (self-borne) | Low | Low |
| Leak prevention & tracing | Yes | Yes | Yes (focus) |
| Data-in-use protection | Yes (if done right) | Strong | Limited |
| Vendor lock-in | No | High | Low |
| Data residency (VN) | Self-controlled | Needs confirmation | Self-controlled |

**Positioning conclusion:** for the top goal of leak prevention and tracing, the hybrid achieves most value at the lowest cost and risk, while keeping full control of data domestically.

## Phased roadmap summary

| Phase | Focus | Key outcome |
|---|---|---|
| P1 | Survey & data mapping | PII map, classification matrix |
| P2 | Centralized audit (top priority) | Every PII access leaves an immutable trail |
| P3 | Encryption (TDE + column) & masking | Leaked DB/backup no longer leaks real data |
| P4 | Tighten access control | Least-privilege, default-deny |
| P5 | Monitoring & anomaly detection | Early detection; minutes to investigate |
| P6 | Compliance, audit, operations | Decree 13/2023 compliance evidence |
