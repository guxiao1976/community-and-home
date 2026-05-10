<!--
Sync Impact Report:
- Version change: [UNVERSIONED] → 1.0.0
- Modified principles: N/A (initial version)
- Added sections: All core principles (5), Code Generation Rules, Data & Security Rules, Development Workflow, Quality Standards
- Removed sections: N/A
- Templates requiring updates:
  ✅ spec-template.md - reviewed, aligns with Spec-first principle
  ✅ plan-template.md - reviewed, Constitution Check section present
  ✅ tasks-template.md - reviewed, aligns with microservice development workflow
- Follow-up TODOs: None
-->

# Community and Home Management Platform Constitution

## Core Principles

### I. Spec is the Single Source of Truth (NON-NEGOTIABLE)

All business logic MUST be written in Spec files before code generation. Code MUST NOT contain any logic not present in the Spec. Any requirement changes MUST modify the Spec first, then regenerate code.

**Rationale**: Prevents code drift, ensures documentation accuracy, and maintains traceability between requirements and implementation.

**Enforcement**:

- ✅ All business logic written in Spec before code generation
- ✅ Code generated from approved Specs only
- ✅ Requirement changes modify Spec first, then regenerate code
- ❌ Direct code modification prohibited
- ❌ Adding business logic to code prohibited
- ❌ Bypassing Spec for feature development prohibited

### II. Microservices by Business Capability

Each business capability MUST be an independent go-zero microservice with its own deployment, database, and cache. Microservices communicate via gRPC only. Etcd serves as the service registry and discovery center.

**Rationale**: Enables independent scaling, deployment, and team ownership. Prevents tight coupling and supports the target scale of 100k communities and 1M users.

**Service Catalog**:

- identity: User authentication and authorization
- masterdata: Master data management (headquarters-controlled)
- community: Community/village management
- repair: Maintenance and repair workflows
- finance: Financial transactions and billing
- content: Content management and publishing
- notification: Notification delivery
- operation: Operational workflows
- security: Security monitoring and access control

**Enforcement**:

- ✅ Each business capability as independent go-zero microservice
- ✅ Independent deployment, database, and cache per service
- ✅ gRPC for inter-service communication
- ✅ Etcd for service registration and discovery
- ❌ Multiple unrelated capabilities in one service prohibited
- ❌ Cross-database direct access prohibited

### III. go-zero Three-Layer Architecture (NON-NEGOTIABLE)

Every microservice MUST follow go-zero's strict three-layer architecture: API → Logic → Model. External exposure uses go-zero API services, internal calls use go-zero RPC services. All database operations MUST use goctl-generated Model layer. Built-in go-zero features (cache, circuit breaker, rate limiting, logging, monitoring, tracing) MUST be used.

**Rationale**: Ensures consistency, leverages framework capabilities, and prevents architectural violations that lead to maintenance issues.

**Enforcement**:

- ✅ Strict API → Logic → Model layering
- ✅ go-zero API for external exposure, RPC for internal calls
- ✅ goctl-generated Model layer for all database operations
- ✅ go-zero built-in cache, circuit breaker, rate limiting, logging, monitoring, tracing
- ✅ Dependency injection for all external dependencies
- ❌ Direct SQL in Logic layer prohibited
- ❌ Other ORM frameworks prohibited
- ❌ Cross-layer calls prohibited

### IV. Two-Tier Governance Architecture

The masterdata service is the sole master data provider. All static master data is maintained exclusively by headquarters. Community/village data is submitted by provincial/municipal levels through masterdata service and becomes effective after headquarters approval. Core data deletion rights are centralized at headquarters via masterdata service. All public welfare content, contract templates, and legal documents are published centrally by headquarters through the content service.

**Rationale**: Ensures data consistency, regulatory compliance, and centralized control over critical business data across 100k communities.

**Enforcement**:

- ✅ masterdata service as sole master data provider
- ✅ Static master data maintained exclusively by headquarters
- ✅ Community/village data submitted by provinces/cities, approved by headquarters
- ✅ Core data deletion centralized at headquarters via masterdata service
- ✅ Public content, templates, legal documents published by headquarters via content service
- ❌ Other services modifying master data prohibited
- ❌ Provincial/municipal deletion of core data prohibited

### V. AI as Development Assistant, Not Decision Maker

Claude is a development assistant that generates code based on approved Specs only. All AI-generated code MUST undergo human review. Claude MUST NOT write Specs, make architectural decisions, or receive sensitive information.

**Rationale**: Maintains human oversight, prevents unauthorized architectural changes, and protects sensitive data.

**Enforcement**:

- ✅ Claude generates code from approved Specs only
- ✅ All AI-generated code undergoes human review
- ❌ Claude writing Specs prohibited
- ❌ Claude making architectural decisions prohibited
- ❌ Inputting sensitive information to Claude prohibited

## Code Generation Rules

### goctl Generation Standards

All RPC services MUST use `goctl rpc new`. All API services MUST use `goctl api new`. All Model code MUST use `goctl model mysql datasource -cache=true`. Generated code MUST be placed in `services/{service-name}/`. Generated handler, server, model, and types code MUST NOT be modified. Business logic MUST only be written in Logic layer.###

### Claude Code Generation Standards

Generated code MUST strictly follow go-zero microservice specifications. Generated code MUST include complete error handling. Generated code MUST include unit tests. Generated code MUST comply with all constitution requirements.

Frontend API, types, and page code MUST be generated based on backend Spec files and keep fields consistent.

**Prohibitions**:

- ❌ Claude adding features not in Spec prohibited
- ❌ Claude modifying this constitution prohibited

## Data & Security Rules

**Data Management**:

- ✅ All core data uses soft delete (delete_time field)
- ✅ All important data changes MUST be logged
- ✅ Sensitive data MUST be encrypted at rest
- ✅ Passwords MUST use bcrypt encryption
- ❌ Physical deletion of core data prohibited
- ❌ Plaintext storage of sensitive information prohibited

**Security Requirements**:

- ✅ All APIs MUST enforce authentication and authorization
- ✅ All user input MUST be validated and sanitized
- ✅ Microservice calls MUST use gRPC interceptors for authentication

## Development Workflow

**Process**:

1. Create `spec/xxx` branch
2. Write microservice feature spec
3. Spec review and approval, merge to main
4. Create `feature/xxx` branch
5. Claude generates code from Spec
6. Code review
7. Run all tests
8. Merge to main branch
9. Deploy microservice independently
10. Frontend development follows backend API definitions, synchronizes Specs first, then generates code.

**Commit Standards**:

- ✅ All commits MUST follow format: `[type]: [description]\n\n[spec-file-path]`
- ✅ Type MUST be one of: spec, code, test, fix, docs
- ✅ All tests MUST pass before merge
- ✅ Each microservice developed, tested, and deployed independently
- ❌ Committing untested code prohibited
- ❌ Direct commits to main branch prohibited

## Quality Standards

**Code Quality**:

- ✅ All code MUST pass `go fmt` and `go vet`
- ✅ All exported functions and types MUST have comments
- ✅ Function length MUST NOT exceed 50 lines
- ✅ Variable and function names MUST be descriptive (no pinyin or abbreviations)
- ✅ Core business logic unit test coverage ≥ 80%

**Performance Standards**:

- ✅ API response time P99 ≤ 500ms
- ✅ RPC call response time P99 ≤ 100ms
- ✅ Database query response time ≤ 100ms
- ✅ System availability ≥ 99.9%

## Governance

This constitution supersedes all other development practices. All amendments require documentation, approval, and migration plan. All PRs and code reviews MUST verify compliance with this constitution. Any complexity or deviation MUST be explicitly justified in the implementation plan.

**Amendment Process**:

1. Propose amendment with rationale
2. Technical review and impact analysis
3. Approval by project leadership
4. Update constitution version
5. Communicate changes to all teams
6. Update dependent templates and documentation

**Compliance**:

- All Specs MUST align with constitutional principles
- All generated code MUST comply with architectural rules
- All deployments MUST meet quality standards
- Violations MUST be documented and remediated

**Version**: 1.0.0 | **Ratified**: 2026-04-13 | **Last Amended**: 2026-04-13

## VI. Frontend Development Specifications (FRONTEND RULES)

### 1. Tech Stack (Mandatory)

#### PC Management End (For Headquarters/Operation Staff)

- Framework: Vue3 + TypeScript
- Core Coding Pattern: **Composition API (Mandatory)**
- SFC Syntax: **`<script setup>` syntax sugar (Mandatory)**
- Reactivity API: Vue3 official reactivity (ref / reactive / computed / watch)
- UI Component Library: Element Plus
- Request Library: Axios (unified encapsulation, consistent with common layer)
- State Management: Pinia
- Style: SCSS / CSS Modules
- Application Scenarios: Backend management, data configuration, permission control, and operational workflows
  
  #### Mobile End (For Residents/End Users)
- Framework: Vue3 + TypeScript
- Core Coding Pattern: **Composition API (Mandatory)**
- SFC Syntax: **`<script setup>` syntax sugar (Mandatory)**
- Reactivity API: Vue3 official reactivity (ref / reactive / computed / watch)
- UI Component Library: Vant (mobile-adapted)
- Request Library: Axios (unified encapsulation, consistent with common layer)
- State Management: Pinia
- Style: SCSS / CSS Modules (mobile-adapted, responsive design)
- Application Scenarios: Community services, maintenance repair, information browsing, and user-related operations
  
  ### 2. Frontend Directory Structure (PC & Mobile Separation, Mandatory)
  
  All frontend code MUST be placed in the `web/` directory of the root project. PC and mobile ends are independent in engineering, directory, and deployment, while sharing common resources.
  Standard Directory Structure:
  
  ```
  web/
  ├── pc/ # PC Management End (independent project)
  │ ├── src/api # PC-specific API request functions (extended from common API)
  │ ├── src/views # PC page components
  │ ├── src/components # PC-specific public components
  │ ├── src/stores # PC Pinia state management
  │ ├── src/utils # PC-specific tool functions
  │ └── src/types # PC-specific TS type definitions (extended from common types)
  │
  ├── mobile/ # Mobile End (independent project)
  │ ├── src/api # Mobile-specific API request functions (extended from common API)
  │ ├── src/views # Mobile page components
  │ ├── src/components # Mobile-specific public components
  │ ├── src/stores # Mobile Pinia state management
  │ ├── src/utils # Mobile-specific tool functions
  │ └── src/types # Mobile-specific TS type definitions (extended from common types)
  │
  └── common/ # Cross-end shared resources (unified, non-modifiable)
  ├── api/ # Unified API definitions (aligned with backend Spec)
  ├── types/ # Unified TS type definitions (synchronized with backend proto/struct)
  ├── utils/ # Cross-end shared tool functions (e.g., authentication, desensitization)
  └── constants/ # Unified constants (error codes, enums, consistent with backend)
  ```
  
  ### 3. Directory Separation Rules (NON-NEGOTIABLE)
  
  **Enforcement**:
- ✅ PC and mobile business code MUST be completely isolated, no cross-dependency
- ✅ Cross-end shared resources (API, types, constants) MUST be maintained in `web/common/` and unified across both ends
- ✅ PC and mobile ends MUST be developed, built, and deployed independently
- ✅ UI components of PC (Element Plus) and mobile (Vant) MUST NOT be mixed
- ❌ Reusing PC business code in mobile end or vice versa is prohibited
- ❌ Modifying shared resources in `web/common/` manually is prohibited (generated from backend Spec only)
- ❌ Merging PC and mobile code into a single directory is prohibited
  
  ### 4. API & Data Specifications (NON-NEGOTIABLE)
  
  **Enforcement**:
- ✅ All API request/response fields, error codes, and enums MUST be completely consistent with backend Spec and Go structs
- ✅ Unified TS type definitions in `web/common/types/` MUST be synchronized with backend pb/struct definitions
- ✅ API request headers (token, userId, etc.) MUST follow backend authentication rules for both ends
- ✅ PC and mobile ends MUST use API functions extended from `web/common/api/`; no independent API definitions are allowed
- ❌ Modifying field names, data types, or error codes independently is prohibited
- ❌ Creating separate API/type definitions outside `web/common/` for common business is prohibited
  
  ### 5. Vue3 Coding Rules (NON-NEGOTIABLE)
  
  **Enforcement**:
- ✅ All Vue components MUST use `<script setup>` syntax
- ✅ All business logic MUST be written with Composition API
- ✅ All reactive data MUST use Vue3 official APIs: `ref`, `reactive`, `computed`, `watch`
- ✅ Props and emits MUST be strictly typed with TypeScript
- ❌ Options API is prohibited in all components
- ❌ Mixed use of Composition API and Options API is prohibited
- ❌ Uncontrolled responsive data operations are prohibited
  
  ### 6. Code Generation Rules for Frontend
  
  **Enforcement**:
- ✅ Claude MUST generate cross-end shared resources (`web/common/api/`, `web/common/types/`, `web/common/constants/`) based on backend Spec files
- ✅ Claude MUST generate PC/mobile-specific API extensions and page code based on shared resources
- ✅ All generated Vue components MUST comply with `<script setup>` + Composition API rules
- ❌ Manual modification of auto-generated code (shared API/types/constants, end-specific API extensions) is prohibited
- ❌ Claude generating end-specific code that deviates from shared resources is prohibited
  
  ### 7. Security & Permissions (Mandatory)
  
  **Enforcement**:
- ✅ All pages (PC and mobile) MUST implement login authentication and permission verification
- ✅ Sensitive data (mobile phone number, ID card, etc.) MUST be desensitized on the frontend (unified desensitization rules in `web/common/utils/`)
- ✅ Token storage (localStorage/sessionStorage) and transmission MUST follow backend specifications, consistent across both ends
- ✅ PC end MUST implement role-based permission control (aligned with backend identity service)
- ✅ Mobile end MUST only expose user-related functions; no administrative permissions are allowed
- ❌ Plaintext display/storage of sensitive information is prohibited
- ❌ Bypassing authentication to access restricted pages is prohibited
  
  ### 8. Frontend Development Workflow (Mandatory)
  
  **Process**:
1. Synchronize backend API Spec and field definitions
2. Backend API documentation storage directory: docs/api/
3. Claude generates cross-end shared resources (`web/common/`) based on backend Spec
4. Create separate branches for PC and mobile development (`feature/pc-xxx`, `feature/mobile-xxx`)
5. Claude generates end-specific code (API extensions, page components) based on shared resources
6. Develop end-specific business logic with Composition API + `<script setup>`
7. Connect interfaces, debug, and conduct end-specific tests
8. Code review (verify compliance with this constitution)
9. Merge to main branch and deploy PC/mobile ends independently
   **Commit Standards (Extended)**:
- ✅ All frontend commits MUST follow the format: `[type]: [description]\n\n[spec-file-path]`
- ✅ Type MUST include: `frontend-pc`, `frontend-mobile`, `frontend-common`
- ✅ All tests MUST pass before merge
- ❌ Committing untested code is prohibited
- ❌ Direct commits to main branch are prohibited
