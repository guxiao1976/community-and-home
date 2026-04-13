# Tasks: Identity and Masterdata Microservices

**Input**: Design documents from `/specs/001-identity-masterdata/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Tests are OPTIONAL - only include them if explicitly requested in the feature specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Microservices**: `services/identity/`, `services/masterdata/`
- **Common utilities**: `common/`
- **SQL scripts**: `scripts/sql/`
- **Deployment**: `deploy/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create project directory structure (services/identity, services/masterdata, common, scripts/sql, deploy)
- [ ] T002 Initialize Go modules for identity service in services/identity/go.mod
- [ ] T003 Initialize Go modules for masterdata service in services/masterdata/go.mod
- [ ] T004 [P] Install goctl code generator (`go install github.com/zeromicro/go-zero/tools/goctl@latest`)
- [ ] T005 [P] Create .gitignore file in repository root
- [ ] T006 [P] Create common/errorx package for error handling in common/errorx/errors.go
- [ ] T007 [P] Create common/responsex package for unified responses in common/responsex/response.go
- [ ] T008 [P] Create common/jwtx package for JWT utilities in common/jwtx/jwt.go
- [ ] T009 [P] Create common/miniox package for MinIO client in common/miniox/client.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Database Setup

- [ ] T010 Create identity_db database schema DDL script in scripts/sql/identity_schema.sql
- [ ] T011 Create masterdata_db database schema DDL script in scripts/sql/masterdata_schema.sql
- [ ] T012 Create identity_db seed data script (super admin, default roles, permissions) in scripts/sql/identity_seed.sql
- [ ] T013 Create masterdata_db seed data script (national administrative divisions) in scripts/sql/masterdata_seed.sql
- [ ] T014 Execute identity_schema.sql against MySQL identity_db database
- [ ] T015 Execute masterdata_schema.sql against MySQL masterdata_db database
- [ ] T016 Execute identity_seed.sql to populate initial data
- [ ] T017 Execute masterdata_seed.sql to populate administrative divisions

### Model Generation (Identity Service)

- [ ] T018 [P] Generate auth_user model using goctl in services/identity/model/authusermodel.go
- [ ] T019 [P] Generate auth_role model using goctl in services/identity/model/authrolemodel.go
- [ ] T020 [P] Generate auth_permission model using goctl in services/identity/model/authpermissionmodel.go
- [ ] T021 [P] Generate auth_role_permission model using goctl in services/identity/model/authrolepermissionmodel.go
- [ ] T022 [P] Generate auth_user_role model using goctl in services/identity/model/authuserrolemodel.go
- [ ] T023 [P] Generate auth_property_unit model using goctl in services/identity/model/authpropertyunitmodel.go
- [ ] T024 [P] Generate auth_property_binding model using goctl in services/identity/model/authpropertybindingmodel.go
- [ ] T025 [P] Generate auth_homeowner_verification model using goctl in services/identity/model/authhomeownerverificationmodel.go
- [ ] T026 [P] Generate auth_family model using goctl in services/identity/model/authfamilymodel.go
- [ ] T027 [P] Generate auth_family_member model using goctl in services/identity/model/authfamilymembermodel.go
- [ ] T028 [P] Generate auth_uploaded_file model using goctl in services/identity/model/authuploadedfilemodel.go

### Model Generation (Masterdata Service)

- [ ] T029 [P] Generate md_administrative_division model using goctl in services/masterdata/model/mdadministrativedivisionmodel.go
- [ ] T030 [P] Generate md_community model using goctl in services/masterdata/model/mdcommunitymodel.go
- [ ] T031 [P] Generate md_district_economic_data model using goctl in services/masterdata/model/mddistricteconomicdatamodel.go
- [ ] T032 [P] Generate md_configuration model using goctl in services/masterdata/model/mdconfigurationmodel.go
- [ ] T033 [P] Generate md_sensitive_word model using goctl in services/masterdata/model/mdsensitivewordmodel.go
- [ ] T034 [P] Generate md_audit_log model using goctl in services/masterdata/model/mdauditlogmodel.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Headquarters Administrator Manages Master Data (Priority: P1) 🎯 MVP

**Goal**: Enable headquarters administrators to manage five-tier administrative divisions and review community submissions from provincial/municipal administrators

**Independent Test**: Create administrative divisions, have provincial admins submit community data, headquarters approve/reject submissions

### API Definition (Masterdata)

- [ ] T035 [US1] Define masterdata API contract in services/masterdata/api/masterdata.api
- [ ] T036 [US1] Generate masterdata API code using goctl in services/masterdata/api/

### RPC Definition (Masterdata)

- [ ] T037 [US1] Define masterdata RPC proto in services/masterdata/rpc/masterdata.proto
- [ ] T038 [US1] Generate masterdata RPC code using goctl in services/masterdata/rpc/

### Configuration

- [ ] T039 [US1] Create masterdata API configuration in services/masterdata/api/etc/masterdata-api.yaml
- [ ] T040 [US1] Create masterdata RPC configuration in services/masterdata/rpc/etc/masterdata-rpc.yaml

### Logic Implementation (Administrative Divisions)

- [ ] T041 [US1] Implement GetDivisionsLogic (list/tree) in services/masterdata/api/internal/logic/getdivisionslogic.go
- [ ] T042 [US1] Implement GetDivisionLogic (details) in services/masterdata/api/internal/logic/getdivisionlogic.go
- [ ] T043 [US1] Implement CreateDivisionLogic in services/masterdata/api/internal/logic/createdivisionlogic.go
- [ ] T044 [US1] Implement UpdateDivisionLogic in services/masterdata/api/internal/logic/updatedivisionlogic.go
- [ ] T045 [US1] Implement DeleteDivisionLogic (soft delete) in services/masterdata/api/internal/logic/deletedivisionlogic.go
- [ ] T046 [US1] Implement path materialization trigger logic in services/masterdata/api/internal/logic/divisionpathlogic.go

### Logic Implementation (Community Management)

- [ ] T047 [US1] Implement GetCommunitiesLogic (list with scope filter) in services/masterdata/api/internal/logic/getcommunitieslogic.go
- [ ] T048 [US1] Implement GetCommunityLogic (details) in services/masterdata/api/internal/logic/getcommunitylogic.go
- [ ] T049 [US1] Implement CreateCommunityLogic (provincial/municipal) in services/masterdata/api/internal/logic/createcommunitylogic.go
- [ ] T050 [US1] Implement UpdateCommunityLogic in services/masterdata/api/internal/logic/updatecommunitylogic.go
- [ ] T051 [US1] Implement SubmitCommunityLogic (change status to submitted) in services/masterdata/api/internal/logic/submitcommunitylogic.go
- [ ] T052 [US1] Implement ReviewCommunityLogic (approve/reject, headquarters only) in services/masterdata/api/internal/logic/reviewcommunitylogic.go
- [ ] T053 [US1] Implement DeleteCommunityLogic (soft delete, headquarters only) in services/masterdata/api/internal/logic/deletecommunitylogic.go

### RPC Implementation (Masterdata)

- [ ] T054 [P] [US1] Implement GetDivision RPC in services/masterdata/rpc/internal/logic/getdivisionlogic.go
- [ ] T055 [P] [US1] Implement GetDivisionsByIds RPC in services/masterdata/rpc/internal/logic/getdivisionsbyidslogic.go
- [ ] T056 [P] [US1] Implement GetDivisionTree RPC in services/masterdata/rpc/internal/logic/getdivisiontreelogic.go
- [ ] T057 [P] [US1] Implement GetDivisionPath RPC in services/masterdata/rpc/internal/logic/getdivisionpathlogic.go
- [ ] T058 [P] [US1] Implement ValidateScope RPC in services/masterdata/rpc/internal/logic/validatescopelogic.go
- [ ] T059 [P] [US1] Implement GetCommunity RPC in services/masterdata/rpc/internal/logic/getcommunitylogic.go
- [ ] T060 [P] [US1] Implement GetCommunitiesByIds RPC in services/masterdata/rpc/internal/logic/getcommunitiesbyidslogic.go
- [ ] T061 [P] [US1] Implement GetCommunitiesByDivision RPC in services/masterdata/rpc/internal/logic/getcommunitiesbydivisionlogic.go

### Audit Logging

- [ ] T062 [US1] Implement audit log middleware in services/masterdata/api/internal/middleware/auditlogmiddleware.go
- [ ] T063 [US1] Integrate audit logging for all division and community changes

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - Backend User Authentication and Authorization (Priority: P1) 🎯 MVP

**Goal**: Enable backend users to register, login, and have permissions enforced based on role and administrative scope

**Independent Test**: Register users with different roles, login, verify scope-based data access enforcement

### API Definition (Identity)

- [ ] T064 [US2] Define identity API contract in services/identity/api/identity.api
- [ ] T065 [US2] Generate identity API code using goctl in services/identity/api/

### RPC Definition (Identity)

- [ ] T066 [US2] Define identity RPC proto in services/identity/rpc/identity.proto
- [ ] T067 [US2] Generate identity RPC code using goctl in services/identity/rpc/

### Configuration

- [ ] T068 [US2] Create identity API configuration in services/identity/api/etc/identity-api.yaml
- [ ] T069 [US2] Create identity RPC configuration in services/identity/rpc/etc/identity-rpc.yaml

### JWT & Authentication Logic

- [ ] T070 [US2] Implement JWT token generation in common/jwtx/generate.go
- [ ] T071 [US2] Implement JWT token validation in common/jwtx/validate.go
- [ ] T072 [US2] Implement JWT middleware in services/identity/api/internal/middleware/jwtmiddleware.go
- [ ] T073 [US2] Implement SMS verification code service in services/identity/api/internal/logic/smslogic.go

### User Management Logic

- [ ] T074 [US2] Implement RegisterLogic (phone + SMS code) in services/identity/api/internal/logic/registerlogic.go
- [ ] T075 [US2] Implement LoginLogic (phone + password) in services/identity/api/internal/logic/loginlogic.go
- [ ] T076 [US2] Implement LoginSmsLogic (phone + SMS code) in services/identity/api/internal/logic/loginsmslogic.go
- [ ] T077 [US2] Implement RefreshTokenLogic in services/identity/api/internal/logic/refreshtokenlogic.go
- [ ] T078 [US2] Implement LogoutLogic (token blacklist) in services/identity/api/internal/logic/logoutlogic.go
- [ ] T079 [US2] Implement SendSmsCodeLogic (rate limiting) in services/identity/api/internal/logic/sendsmscodelogic.go
- [ ] T080 [US2] Implement GetUsersLogic (list with scope filter) in services/identity/api/internal/logic/getuserslogic.go
- [ ] T081 [US2] Implement GetUserLogic (details) in services/identity/api/internal/logic/getuserlogic.go
- [ ] T082 [US2] Implement CreateUserLogic (backend users) in services/identity/api/internal/logic/createuserlogic.go
- [ ] T083 [US2] Implement UpdateUserLogic in services/identity/api/internal/logic/updateuserlogic.go
- [ ] T084 [US2] Implement DeleteUserLogic (soft delete, headquarters only) in services/identity/api/internal/logic/deleteuserlogic.go

### Permission Enforcement

- [ ] T085 [US2] Install and configure Casbin in services/identity/api/internal/svc/servicecontext.go
- [ ] T086 [US2] Implement permission check middleware in services/identity/api/internal/middleware/permissionmiddleware.go
- [ ] T087 [US2] Implement scope validation middleware in services/identity/api/internal/middleware/scopemiddleware.go
- [ ] T088 [US2] Implement GetUserPermissionsLogic in services/identity/api/internal/logic/getuserpermissionslogic.go

### RPC Implementation (Identity)

- [ ] T089 [P] [US2] Implement GetUser RPC in services/identity/rpc/internal/logic/getuserlogic.go
- [ ] T090 [P] [US2] Implement GetUsersByIds RPC in services/identity/rpc/internal/logic/getusersbyidslogic.go
- [ ] T091 [P] [US2] Implement ValidateToken RPC in services/identity/rpc/internal/logic/validatetokenlogic.go
- [ ] T092 [P] [US2] Implement CheckPermission RPC in services/identity/rpc/internal/logic/checkpermissionlogic.go
- [ ] T093 [P] [US2] Implement GetUserPermissions RPC in services/identity/rpc/internal/logic/getuserpermissionslogic.go

**Checkpoint**: At this point, User Story 2 should be fully functional and testable independently

---

## Phase 5: User Story 5 - Role and Permission Management (Priority: P2)

**Goal**: Enable headquarters administrators to create custom roles, assign permissions, and manage RBAC

**Independent Test**: Create custom roles with specific permissions, assign to users, verify permission enforcement

### Role Management Logic

- [ ] T094 [US5] Implement GetRolesLogic (list) in services/identity/api/internal/logic/getroleslogic.go
- [ ] T095 [US5] Implement GetRoleLogic (details with permissions) in services/identity/api/internal/logic/getrolelogic.go
- [ ] T096 [US5] Implement CreateRoleLogic in services/identity/api/internal/logic/createrolelogic.go
- [ ] T097 [US5] Implement UpdateRoleLogic (update permissions) in services/identity/api/internal/logic/updaterolelogic.go
- [ ] T098 [US5] Implement DeleteRoleLogic (prevent system role deletion) in services/identity/api/internal/logic/deleterolelogic.go

### Permission Management Logic

- [ ] T099 [US5] Implement GetPermissionsLogic (tree structure) in services/identity/api/internal/logic/getpermissionslogic.go
- [ ] T100 [US5] Implement permission tree builder in services/identity/api/internal/logic/permissiontreelogic.go

### Casbin Policy Management

- [ ] T101 [US5] Implement Casbin policy sync on role changes in services/identity/api/internal/logic/casbinpolicylogic.go
- [ ] T102 [US5] Implement Redis pub/sub for policy reload in services/identity/api/internal/logic/policyreloadlogic.go

### RPC Implementation

- [ ] T103 [P] [US5] Implement GetRolesByIds RPC in services/identity/rpc/internal/logic/getrolesbyidslogic.go
- [ ] T104 [P] [US5] Implement GetUserRoles RPC in services/identity/rpc/internal/logic/getuserroleslogic.go

**Checkpoint**: At this point, User Story 5 should be fully functional and testable independently

---

## Phase 6: User Story 3 - Homeowner Identity Verification and Property Binding (Priority: P2)

**Goal**: Enable homeowners to upload property certificates, get verified, and bind to property units

**Independent Test**: Homeowner uploads documents, headquarters reviews and approves, homeowner binds to property

### MinIO Integration

- [ ] T105 [US3] Implement MinIO client initialization in common/miniox/client.go
- [ ] T106 [US3] Implement presigned URL generation in common/miniox/presigned.go

### Property Management Logic

- [ ] T107 [US3] Implement GetPropertyUnitsLogic (list) in services/identity/api/internal/logic/getpropertyunitslogic.go
- [ ] T108 [US3] Implement GetPropertyUnitLogic (details) in services/identity/api/internal/logic/getpropertyunitlogic.go

### Homeowner Verification Logic

- [ ] T109 [US3] Implement GenerateUploadUrlLogic (presigned URL for MinIO) in services/identity/api/internal/logic/generateuploadurllogic.go
- [ ] T110 [US3] Implement SubmitVerificationLogic (with document URLs) in services/identity/api/internal/logic/submitverificationlogic.go
- [ ] T111 [US3] Implement GetVerificationsLogic (list for headquarters) in services/identity/api/internal/logic/getverificationslogic.go
- [ ] T112 [US3] Implement ReviewVerificationLogic (approve/reject) in services/identity/api/internal/logic/reviewverificationlogic.go

### Property Binding Logic

- [ ] T113 [US3] Implement CreatePropertyBindingLogic (bind user to property) in services/identity/api/internal/logic/createpropertybindinglogic.go
- [ ] T114 [US3] Implement DeletePropertyBindingLogic (remove user from property) in services/identity/api/internal/logic/deletepropertybindinglogic.go
- [ ] T115 [US3] Implement primary user validation logic in services/identity/api/internal/logic/primaryuserlogic.go

### RPC Implementation

- [ ] T116 [P] [US3] Implement GetPropertyUnit RPC in services/identity/rpc/internal/logic/getpropertyunitlogic.go
- [ ] T117 [P] [US3] Implement GetUserProperties RPC in services/identity/rpc/internal/logic/getuserpropertieslogic.go
- [ ] T118 [P] [US3] Implement GetVerificationStatus RPC in services/identity/rpc/internal/logic/getverificationstatuslogic.go

**Checkpoint**: At this point, User Story 3 should be fully functional and testable independently

---

## Phase 7: User Story 4 - Family Center Management (Priority: P3)

**Goal**: Enable verified homeowners to create family profiles and manage family members

**Independent Test**: Verified homeowner creates family, adds members, updates information

### Family Management Logic

- [ ] T119 [US4] Implement CreateFamilyLogic in services/identity/api/internal/logic/createfamilylogic.go
- [ ] T120 [US4] Implement GetFamilyLogic (details with members) in services/identity/api/internal/logic/getfamilylogic.go
- [ ] T121 [US4] Implement AddFamilyMemberLogic in services/identity/api/internal/logic/addfamilymemberlogic.go
- [ ] T122 [US4] Implement UpdateFamilyMemberLogic in services/identity/api/internal/logic/updatefamilymemberlogic.go
- [ ] T123 [US4] Implement DeleteFamilyMemberLogic in services/identity/api/internal/logic/deletefamilymemberlogic.go
- [ ] T124 [US4] Implement family head validation logic in services/identity/api/internal/logic/familyheadlogic.go

**Checkpoint**: At this point, User Story 4 should be fully functional and testable independently

---

## Phase 8: User Story 6 - Platform Configuration Management (Priority: P3)

**Goal**: Enable headquarters administrators to manage platform-wide configurations and sensitive words

**Independent Test**: Create and modify configuration parameters, add sensitive words, verify changes take effect

### Configuration Management Logic

- [ ] T125 [US6] Implement GetConfigurationsLogic (list by module) in services/masterdata/api/internal/logic/getconfigurationslogic.go
- [ ] T126 [US6] Implement GetConfigurationLogic (details) in services/masterdata/api/internal/logic/getconfigurationlogic.go
- [ ] T127 [US6] Implement CreateConfigurationLogic in services/masterdata/api/internal/logic/createconfigurationlogic.go
- [ ] T128 [US6] Implement UpdateConfigurationLogic in services/masterdata/api/internal/logic/updateconfigurationlogic.go
- [ ] T129 [US6] Implement ApproveConfigurationLogic in services/masterdata/api/internal/logic/approveconfigurationlogic.go

### Sensitive Word Management Logic

- [ ] T130 [US6] Implement GetSensitiveWordsLogic (list) in services/masterdata/api/internal/logic/getsensitivewordslogic.go
- [ ] T131 [US6] Implement CreateSensitiveWordLogic in services/masterdata/api/internal/logic/createsensitivewordlogic.go
- [ ] T132 [US6] Implement UpdateSensitiveWordLogic in services/masterdata/api/internal/logic/updatesensitivewordlogic.go
- [ ] T133 [US6] Implement DeleteSensitiveWordLogic in services/masterdata/api/internal/logic/deletesensitivewordlogic.go
- [ ] T134 [US6] Implement Aho-Corasick algorithm for sensitive word matching in services/masterdata/api/internal/logic/sensitivewordmatchlogic.go

### RPC Implementation

- [ ] T135 [P] [US6] Implement GetConfig RPC in services/masterdata/rpc/internal/logic/getconfiglogic.go
- [ ] T136 [P] [US6] Implement GetConfigsByModule RPC in services/masterdata/rpc/internal/logic/getconfigsbymodulelogic.go
- [ ] T137 [P] [US6] Implement CheckSensitiveWords RPC in services/masterdata/rpc/internal/logic/checksensitivewordslogic.go

**Checkpoint**: All user stories should now be independently functional

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T138 [P] Implement rate limiting middleware in services/identity/api/internal/middleware/ratelimitmiddleware.go
- [ ] T139 [P] Implement rate limiting middleware in services/masterdata/api/internal/middleware/ratelimitmiddleware.go
- [ ] T140 [P] Add structured logging with request ID in common/logx/logger.go
- [ ] T141 [P] Implement error code standardization in common/errorx/codes.go
- [ ] T142 [P] Create Dockerfile for identity service in services/identity/Dockerfile
- [ ] T143 [P] Create Dockerfile for masterdata service in services/masterdata/Dockerfile
- [ ] T144 [P] Update docker-compose.yml with service definitions in deploy/docker-compose.yml
- [ ] T145 [P] Create Kubernetes manifests in deploy/k8s/
- [ ] T146 [P] Add Prometheus metrics endpoints in services/identity/api/identity.go
- [ ] T147 [P] Add Prometheus metrics endpoints in services/masterdata/api/masterdata.go
- [ ] T148 [P] Create API documentation (Swagger/OpenAPI) in docs/api/
- [ ] T149 Run quickstart.md validation (verify all steps work)
- [ ] T150 Performance testing (verify P99 ≤ 200ms for API, ≤ 100ms for RPC)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-8)**: All depend on Foundational phase completion
  - US1 (Masterdata Management): Can start after Foundational
  - US2 (Authentication): Can start after Foundational
  - US5 (Role Management): Depends on US2 (needs authentication)
  - US3 (Homeowner Verification): Depends on US2 (needs authentication)
  - US4 (Family Management): Depends on US3 (needs verified homeowners)
  - US6 (Configuration): Can start after Foundational
- **Polish (Phase 9)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational - No dependencies on other stories
- **User Story 5 (P2)**: Depends on US2 (needs authentication framework)
- **User Story 3 (P2)**: Depends on US2 (needs authentication)
- **User Story 4 (P3)**: Depends on US3 (needs verified homeowners)
- **User Story 6 (P3)**: Can start after Foundational - No dependencies on other stories

### Recommended Execution Order

1. **Phase 1 + Phase 2**: Complete Setup and Foundational (required for all)
2. **Phase 3 + Phase 4**: US1 and US2 in parallel (both P1, independent)
3. **Phase 5**: US5 after US2 completes
4. **Phase 6**: US3 after US2 completes
5. **Phase 7**: US4 after US3 completes
6. **Phase 8**: US6 (can run in parallel with US5/US3/US4)
7. **Phase 9**: Polish after all desired stories complete

### Parallel Opportunities

- **Setup Phase**: T006-T009 can run in parallel (different packages)
- **Model Generation**: T018-T028 (identity) and T029-T034 (masterdata) can all run in parallel
- **User Stories**: US1 and US2 can be developed in parallel by different team members
- **RPC Methods**: Within each story, RPC implementations marked [P] can run in parallel
- **Polish Phase**: T138-T148 can run in parallel (different concerns)

---

## Parallel Example: User Story 2 (Authentication)

```bash
# Launch all RPC implementations for US2 together:
Task: "T089 [P] [US2] Implement GetUser RPC"
Task: "T090 [P] [US2] Implement GetUsersByIds RPC"
Task: "T091 [P] [US2] Implement ValidateToken RPC"
Task: "T092 [P] [US2] Implement CheckPermission RPC"
Task: "T093 [P] [US2] Implement GetUserPermissions RPC"
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Masterdata Management)
4. Complete Phase 4: User Story 2 (Authentication)
5. **STOP and VALIDATE**: Test US1 and US2 independently
6. Deploy/demo if ready

**MVP Delivers**:

- Complete master data management (administrative divisions, communities)
- Full authentication and authorization system
- Scope-based data access control
- Foundation for all other features

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add US1 + US2 → Test independently → Deploy/Demo (MVP!)
3. Add US5 (Role Management) → Test independently → Deploy/Demo
4. Add US3 (Homeowner Verification) → Test independently → Deploy/Demo
5. Add US4 (Family Management) → Test independently → Deploy/Demo
6. Add US6 (Configuration) → Test independently → Deploy/Demo
7. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1 (Masterdata)
   - Developer B: User Story 2 (Authentication)
3. After US2 completes:
   - Developer C: User Story 5 (Role Management)
   - Developer D: User Story 3 (Homeowner Verification)
4. After US3 completes:
   - Developer E: User Story 4 (Family Management)
5. Anytime after Foundational:
   - Developer F: User Story 6 (Configuration)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
- All goctl commands should use `--cache=true` for Model generation
- All Logic layer implementations must follow go-zero conventions
- All database operations must go through Model layer (no direct SQL)
- All API responses must use common/responsex format
- All errors must use common/errorx codes
