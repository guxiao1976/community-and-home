# Implementation Plan: Identity and Masterdata Microservices

**Branch**: `001-identity-masterdata` | **Date**: 2026-04-13 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-identity-masterdata/spec.md`

## Summary

Implement two foundational go-zero microservices (identity and masterdata) providing authentication, authorization, role-based access control, and master data management for a community service platform. The system enforces a two-tier governance model where headquarters controls master data and provincial/municipal administrators manage regional operations within their scope.

## Technical Context

**Language/Version**: Go 1.21+  
**Primary Dependencies**: go-zero 1.6+, gRPC, JWT (golang-jwt/jwt), bcrypt (golang.org/x/crypto/bcrypt)  
**Storage**: MySQL 8.0 (relational data), Redis 7.0 (cache/session), MinIO (object storage for images)  
**Testing**: go test, testify/assert  
**Target Platform**: Linux server (Docker containerized deployment)  
**Project Type**: Microservices (2 services: identity, masterdata)  
**Performance Goals**: API P99 ≤ 200ms, RPC P99 ≤ 100ms, 10k concurrent users  
**Constraints**: Soft delete only for core data, all changes audited, scope-based data isolation  
**Scale/Scope**: 100k communities, 1M users, 5-tier administrative hierarchy

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### ✅ I. Spec is the Single Source of Truth
- All business logic documented in spec.md before implementation
- No code modifications outside of Logic layer
- Spec-driven development workflow enforced

### ✅ II. Microservices by Business Capability
- Two independent microservices: identity (auth/users) and masterdata (admin divisions/config)
- Each service has independent database schema (auth_* and md_* table prefixes)
- Services communicate via gRPC for internal calls
- Etcd for service discovery

### ✅ III. go-zero Three-Layer Architecture
- Strict API → Logic → Model layering
- goctl-generated code for API, RPC, and Model layers
- All database operations through Model layer only
- Built-in go-zero features: cache, circuit breaker, rate limiting, logging, monitoring

### ✅ IV. Two-Tier Governance Architecture
- masterdata service is sole provider of administrative divisions
- Headquarters-only control for core master data (CRUD operations restricted by role)
- Provincial/municipal submission + headquarters approval workflow
- Soft delete with audit logging for all core data

### ✅ V. AI as Development Assistant
- This plan generated from approved spec.md
- Code generation will use goctl + Logic layer implementation
- All generated code subject to human review

**Gate Status**: ✅ PASSED - All constitutional requirements satisfied

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
services/
├── identity/
│   ├── api/
│   │   ├── etc/
│   │   │   └── identity-api.yaml
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── handler/      # goctl generated
│   │   │   ├── logic/        # Business logic here
│   │   │   ├── svc/
│   │   │   └── types/        # goctl generated
│   │   ├── identity.api      # API definition
│   │   └── identity.go
│   ├── rpc/
│   │   ├── etc/
│   │   │   └── identity-rpc.yaml
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── logic/        # Business logic here
│   │   │   ├── server/       # goctl generated
│   │   │   └── svc/
│   │   ├── pb/               # goctl generated
│   │   ├── identity.proto    # RPC definition
│   │   └── identity.go
│   ├── model/                # goctl generated from DB
│   │   ├── authuser.go
│   │   ├── authrole.go
│   │   ├── authpermission.go
│   │   └── ...
│   └── Dockerfile
│
└── masterdata/
    ├── api/
    │   ├── etc/
    │   │   └── masterdata-api.yaml
    │   ├── internal/
    │   │   ├── config/
    │   │   ├── handler/      # goctl generated
    │   │   ├── logic/        # Business logic here
    │   │   ├── svc/
    │   │   └── types/        # goctl generated
    │   ├── masterdata.api    # API definition
    │   └── masterdata.go
    ├── rpc/
    │   ├── etc/
    │   │   └── masterdata-rpc.yaml
    │   ├── internal/
    │   │   ├── config/
    │   │   ├── logic/        # Business logic here
    │   │   ├── server/       # goctl generated
    │   │   └── svc/
    │   ├── pb/               # goctl generated
    │   ├── masterdata.proto  # RPC definition
    │   └── masterdata.go
    ├── model/                # goctl generated from DB
    │   ├── mdadministrativedivision.go
    │   ├── mdcommunity.go
    │   ├── mdconfiguration.go
    │   └── ...
    └── Dockerfile

common/
├── errorx/               # Common error handling
├── jwtx/                 # JWT utilities
├── responsex/            # Unified response format
└── miniox/               # MinIO client wrapper

deploy/
├── docker-compose.yml    # Already exists (MySQL, Redis, Etcd, MinIO)
└── k8s/                  # Kubernetes manifests (future)
```

**Structure Decision**: Two-service microservices architecture following go-zero conventions. Each service has independent API (external HTTP), RPC (internal gRPC), and Model (database) layers. Common utilities shared via `common/` package. Services deployed independently via Docker containers. Infrastructure (MySQL, Redis, Etcd, MinIO) already deployed via docker-compose.yml.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
