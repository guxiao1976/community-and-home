# Implementation Plan: Web PC Admin Frontend Development

**Branch**: `002-web-pc-admin` | **Date**: 2026-05-03 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/002-web-pc-admin/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Develop a comprehensive PC-based administrative frontend for the Community and Home Management Platform using Vue3 + TypeScript + Element Plus. The system provides eight core management modules: authentication, administrative divisions, communities, users, roles/permissions, homeowner verification, system configuration, and sensitive words. The frontend strictly follows the constitution's two-tier governance model (headquarters vs provincial/municipal), implements role-based access control, and maintains complete field consistency with backend APIs documented in docs/api/.

## Technical Context

**Language/Version**: Vue 3.4+ with TypeScript 5.0+  
**Primary Dependencies**: Element Plus (UI), Axios (HTTP), Pinia (state), Vue Router (routing), Vite (build)  
**Storage**: Browser localStorage for JWT tokens (access 24h, refresh 7d), sessionStorage for temporary state  
**Testing**: Vitest (unit tests), Playwright (E2E tests)  
**Target Platform**: Modern desktop browsers (Chrome 90+, Firefox 88+, Edge 90+, Safari 14+)  
**Project Type**: Single-page web application (SPA) for administrative management  
**Performance Goals**: Page load <2s, API response rendering <500ms, form submission <1s, tree rendering (10k nodes) <3s  
**Constraints**: Desktop-only (no mobile responsive), Chinese UI only (no i18n), no offline mode, no real-time updates (polling/refresh required)  
**Scale/Scope**: 8 user stories, 28 Identity API endpoints + 18 Masterdata API endpoints, ~50 pages/dialogs, support 10k divisions + 100k communities

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Spec is the Single Source of Truth
- ✅ **PASS**: All business logic defined in spec.md (8 user stories, 72 functional requirements)
- ✅ **PASS**: Backend APIs fully documented in docs/api/ (identity-service.md, masterdata-service.md)
- ✅ **PASS**: Frontend will be generated from approved spec and backend API definitions
- ✅ **PASS**: No direct code modification; all changes go through spec updates

### Principle II: Microservices by Business Capability
- ✅ **PASS**: Frontend consumes two backend microservices via REST APIs:
  - Identity Service (authentication, users, roles, permissions, verification)
  - Masterdata Service (divisions, communities, configuration, sensitive words)
- ✅ **PASS**: No direct database access; all data via API calls
- ✅ **PASS**: Services communicate independently; frontend does not couple services

### Principle III: go-zero Three-Layer Architecture
- ✅ **PASS**: Backend follows API → Logic → Model; frontend consumes API layer only
- ✅ **PASS**: Frontend respects backend error codes, validation rules, and response formats
- ✅ **PASS**: No business logic duplication; frontend performs UI validation only

### Principle IV: Two-Tier Governance Architecture
- ✅ **PASS**: UI enforces headquarters vs provincial/municipal role separation
- ✅ **PASS**: Masterdata (divisions, communities) follows submission → approval workflow
- ✅ **PASS**: Deletion rights restricted to headquarters roles via permission checks
- ✅ **PASS**: Administrative scope filtering applied to all list views

### Principle V: AI as Development Assistant
- ✅ **PASS**: Claude generates code from approved spec.md and docs/api/ only
- ✅ **PASS**: All generated code subject to human review before merge
- ✅ **PASS**: No architectural decisions made by Claude; follows constitution rules

### Principle VI: Frontend Development Specifications
- ✅ **PASS**: Vue3 + TypeScript with Composition API + `<script setup>` syntax (mandatory)
- ✅ **PASS**: Element Plus for PC admin UI components
- ✅ **PASS**: Code placed in `web/pc/` directory (PC/mobile separation enforced)
- ✅ **PASS**: Shared resources (API types, constants) in `web/common/` synchronized with backend
- ✅ **PASS**: All API fields, error codes, enums consistent with backend Spec
- ✅ **PASS**: Role-based permission control aligned with backend identity service
- ✅ **PASS**: Sensitive data desensitization (phone, ID card) via `web/common/utils/`

**Constitution Compliance**: ✅ ALL GATES PASSED

## Project Structure

### Documentation (this feature)

```text
specs/002-web-pc-admin/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
└── contracts/           # Phase 1 output (/speckit.plan command)
    ├── api-contracts.md      # API request/response contracts
    ├── routing-contracts.md  # Route definitions and navigation
    └── state-contracts.md    # Pinia store contracts
```

### Source Code (repository root)

```text
web/
├── pc/                          # PC Management End (this feature)
│   ├── src/
│   │   ├── api/                 # PC-specific API extensions
│   │   │   ├── identity.ts      # Identity service API calls
│   │   │   ├── masterdata.ts    # Masterdata service API calls
│   │   │   └── index.ts         # API aggregation
│   │   ├── views/               # Page components (8 modules)
│   │   │   ├── auth/            # Login, register, logout
│   │   │   ├── divisions/       # Administrative division management
│   │   │   ├── communities/     # Community management & review
│   │   │   ├── users/           # User management
│   │   │   ├── roles/           # Role & permission management
│   │   │   ├── verification/    # Homeowner verification review
│   │   │   ├── config/          # System configuration
│   │   │   └── sensitive-words/ # Sensitive word management
│   │   ├── components/          # PC-specific reusable components
│   │   │   ├── layout/          # Layout components (header, sidebar, breadcrumb)
│   │   │   ├── common/          # Common UI components (table, form, dialog)
│   │   │   └── business/        # Business components (division-tree, permission-tree)
│   │   ├── stores/              # Pinia state management
│   │   │   ├── auth.ts          # Authentication state (token, user info)
│   │   │   ├── permission.ts    # Permission state (roles, menus, buttons)
│   │   │   └── app.ts           # App state (loading, breadcrumb, sidebar)
│   │   ├── router/              # Vue Router configuration
│   │   │   ├── index.ts         # Router instance
│   │   │   ├── routes.ts        # Route definitions
│   │   │   └── guards.ts        # Navigation guards (auth, permission)
│   │   ├── utils/               # PC-specific utilities
│   │   │   ├── request.ts       # Axios instance with interceptors
│   │   │   ├── validation.ts    # Form validation rules
│   │   │   └── permission.ts    # Permission check helpers
│   │   ├── types/               # PC-specific TS types (extended from common)
│   │   │   ├── views.d.ts       # View-specific types
│   │   │   └── components.d.ts  # Component prop types
│   │   ├── styles/              # Global styles
│   │   │   ├── variables.scss   # SCSS variables
│   │   │   ├── mixins.scss      # SCSS mixins
│   │   │   └── global.scss      # Global styles
│   │   ├── App.vue              # Root component
│   │   └── main.ts              # Entry point
│   ├── public/                  # Static assets
│   ├── index.html               # HTML template
│   ├── vite.config.ts           # Vite configuration
│   ├── tsconfig.json            # TypeScript configuration
│   └── package.json             # Dependencies
│
├── mobile/                      # Mobile End (future feature, not in scope)
│
└── common/                      # Cross-end shared resources
    ├── api/                     # Unified API definitions (generated from backend Spec)
    │   ├── identity/            # Identity service API types
    │   ├── masterdata/          # Masterdata service API types
    │   └── base.ts              # Base API types (request/response wrappers)
    ├── types/                   # Unified TS types (synchronized with backend pb/struct)
    │   ├── identity.d.ts        # Identity service types
    │   ├── masterdata.d.ts      # Masterdata service types
    │   └── common.d.ts          # Common types (pagination, error codes)
    ├── utils/                   # Cross-end shared utilities
    │   ├── auth.ts              # Authentication helpers (token storage)
    │   ├── desensitize.ts       # Data desensitization (phone, ID card)
    │   └── format.ts            # Data formatting (date, number)
    └── constants/               # Unified constants (synchronized with backend)
        ├── error-codes.ts       # Error code enums
        ├── enums.ts             # Business enums (user type, status, etc.)
        └── config.ts            # App configuration (API base URL)

tests/
├── unit/                        # Vitest unit tests
│   ├── components/              # Component tests
│   ├── stores/                  # Store tests
│   └── utils/                   # Utility tests
└── e2e/                         # Playwright E2E tests
    ├── auth.spec.ts             # Authentication flows
    ├── divisions.spec.ts        # Division management
    ├── communities.spec.ts      # Community management
    └── users.spec.ts            # User management
```

**Structure Decision**: Selected "Web application" structure with PC/mobile separation as mandated by Constitution Principle VI. The `web/pc/` directory contains all PC admin code, `web/common/` contains shared resources synchronized with backend APIs, and `web/mobile/` is reserved for future mobile development. This structure enforces complete isolation between PC and mobile business code while maintaining unified API/type definitions.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations. All constitutional requirements are satisfied.
