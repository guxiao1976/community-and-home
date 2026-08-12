# Plan Review — access-data-permission（structure 结构合理性视角）

**审查维度**: 服务归属 / 依赖方向 / Proto 变更识别 / 依赖顺序 / spec 间职责重叠冲突
**审查时间**: 2026-08-12
**审查输入**: `request.md` + `proposal.md` + `specs/{scope-model,capability-layering,registered-user,join-auto-authorization,publish-validation}/spec.md` + 设计依据 `docs/specs/access-control-design.md`(§1.4/§3/§5) + `rbac-design.md` §2.5
**对照项**: 现有 `api-proto/api/permission/v1/permission.proto`、`api-proto/api/masterdata/v1/masterdata.proto`、`api-proto/api/user/v1/user.proto`

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 4 / 🔵 INFO: 7

## 服务归属与依赖方向总评（符合）

spec 层的服务归属与 proposal 影响范围、access-control-design §8 一致，依赖方向正确指向权威：

| 职责 | 归属 | 判定 |
|------|------|:---:|
| scope 三态 / 祖先链统一判据 / GetDataScopes / AssertPublishScope / min_verf_level / registered_user 角色定义 / 缓存 | permission-service（权威） | ✅ |
| CreateUser 自动分配 registered_user / Join/Leave 编排 Assign/RevokeRole | user-service（非权威编排） | ✅ |
| 写接口挂 AssertPublishScope / publisher_id 取 JWT / 读列表按 GetDataScopes 过滤 | community-hub（消费方） | ✅ |
| 「scope 节点 → 祖先链」RPC / 整树缓存 | master-data | ✅ |

`permission.proto` 中 `CheckPermission`/`GetDataScopes`/`AssignRole`/`RevokeRole`/`InvalidateUserCache`/`UpdateUserRoleStatus` 均存在，AssignRoleRequest 已含 scope_type/scope_id/status——大部分编排依赖可复用现有 RPC，无需新增基础设施。Proto 变更由全局 Owner 在阶段 4 执行（`request.md` 阶段表、`proposal.md`、`.change.yaml proto_change_required: true` 均已标识）✅。

## 发现

### 🔴 MUST FIX

无。

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | `docs/specs/access-control-design.md` §8 依赖清单；`scope-model REQ-1.3/1.4`、`publish-validation REQ-5.1` | **「祖先链 RPC」的消费方未钉死，design §8 与 spec 冲突**。design §8 写「community-hub ──读配额/祖先链──▶ master-data」，即 community-hub 直读 master-data 祖先链；但 scope-model REQ-1.3 与 REQ-5.1 把覆盖判据（A(t)∩S）封装在 permission-service 的 AssertPublishScope/GetDataScopes 内，这两个 RPC 内部需要祖先链，即 permission-service 才应消费 master-data。若按 §8 字面执行，community-hub 需先向 master-data 解析祖先链再传入 permission-service，等于把「统一判据」拆到两处，违背"permission-service 权威"。 | 在 proposal/spec 显式写明依赖方向：`community-hub → permission-service(AssertPublishScope/GetDataScopes) → master-data(祖先链 RPC)`，community-hub **不直连** master-data；阶段 3 架构设计时同步修订 design §8 的依赖行（祖先链只归 permission-service 消费，community-hub 对 master-data 仅剩未来 §7 配额）。 |
| 2 | `proposal.md`「Proto 变更」节 | **Proto 变更清单不完整，仅列 3 项**。至少遗漏 message 级变更：① `GetDataScopes` 现有响应为 `repeated int64 scope_ids`（permission.proto L216），无法表达「global/限定/空」三态——global 需"全放行"语义、空需"明确为空"而非空列表歧义，proposal 声称的「GetDataScopes 三态语义」是数据层语义，**响应契约变更未列入 Proto 清单**；② `AssertPublishScope` 的 `targets []ScopeRef` 需要新 `ScopeRef` message（proposal 仅在 common 附注提到）；③ master-data 祖先链 RPC 的 req/resp message 未列。 | 阶段 3 架构设计把 Proto 变更清单补全到 message 级（GetDataScopes 三态表达 / ScopeRef / 祖先链 RPC req-resp / Permission.min_verf_level），一次性交阶段 4 Owner 执行，避免 Owner 中途扩列。三处均为增量字段/RPC，无破坏性变更。 |
| 3 | `.change.yaml` `common_change_required: false` vs `proposal.md` 影响范围「common 可能（需全局评估）」 | **声明不一致**。proposal 影响范围表写 common「可能」，风险评估明确缓解依赖「提交前全局评估 + 门禁」（硬约束 #6：修改 common/ 需全局评估）；但 `.change.yaml` 声明 `common_change_required: false`，pipeline 可能据此跳过全局评估门禁。 | 将 `.change.yaml` 改为 `common_change_required: true`（最稳妥），或删除 proposal 中「可能入 common」表述并在阶段 3 明确不触碰 common/。 |
| 4 | `scope-model REQ-1.1`（"per scope_type 恰好三态之一"） | **多角色 scope 合并规则缺失**。REQ-1.1 说每用户 per scope_type 恰好三态之一，但用户可同时持多个社区角色（如 join A+B → limited{A,B}），且可能同时持 global（审核员/超管）与 community 角色——不同 grant 跨状态时取哪种优先级（任一 global → global？否则 limited 取并集？否则 empty？）未定义。若无显式合并规则，实现期 GetDataScopes/AssertPublishScope 会出现歧义。 | 在 scope-model 增加一条 REQ 显式定义合并优先级：任一 global grant → global；无 global 但有限定 grant → limited（取并集）；否则 empty。REQ-1.7 已暗示 global 优先，补全为完整规则即可。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | **实施顺序**：本变更为后端-only（proposal 影响范围无 web/pc、web/mobile，前端阶段不适用）。建议顺序：阶段 4 Proto（先做，且 master-data 祖先链 RPC 是 permission-service 前置）→ permission-service + master-data（权威，可并行）→ user-service（编排）+ community-hub（消费，可并行）→ 集成。阶段 3 生成 tasks.md 时应编码该依赖序（harness-pipeline 每服务一个 Workflow、无依赖并行，但权威→消费的先后需在 task 依赖里体现）。 |
| 2 | master-data 已有 `GetDivisionPath`/`GetDivisionTree`/`ValidateScope`，新「祖先链 RPC」与现有能力部分重叠（行政区划祖先解析可走 GetDivisionPath）。阶段 3 应评估是复用/组合（GetResidentialArea 取 community_div_id → GetDivisionPath）还是新增，避免双套祖先解析能力。 |
| 3 | scope-model REQ-1.6 读过滤波及 community-hub **所有读/列表接口**，爆炸面大（proposal 只写「读列表按 GetDataScopes 过滤」）。阶段 3/5 应明确挂载端点清单与改造范围，否则编码期范围失控。 |
| 4 | 空 scope 的 `rel_user_role` 表示已由 proposal 边界说明 #2 留给架构设计。注意：现有 scope_type 枚举为 community/building/unit/grid/global，空=无 scope 行即可，不构成 Proto 破坏性变更；但 GetDataScopes 的「空」必须返回可区分状态（而非"空列表=全放行"歧义），与 SHOULD FIX #2 同源。 |
| 5 | `tasks.md` 尚未生成（按 owner-agent 阶段表属阶段 3 产出，本阶段 2 评审 spec 属预期）。本次为 spec 级结构评审，task 级分组与依赖序需在阶段 3 tasks.md 生成后复核上述 SHOULD FIX。 |
| 6 | 错误码命名空间（design §5.5 用 `80006`，community-hub 现状为 `08xxxx`）已在 proposal 风险中承认并留给架构对齐——结构上无阻塞，阶段 3 定稿即可。 |
| 7 | 缓存失效在 3 个 spec 重复出现（scope-model REQ-1.8 / capability-layering REQ-2.5 / join REQ-4.4），但触发源不同（join/leave / 认证状态变更），且 cache key（`perm:scopes`/`perm:user`）与失效 RPC（`InvalidateUserCache`）均属 permission-service，归属一致、无冲突。建议阶段 3 把失效逻辑收敛到 AssignRole/RevokeRole/UpdateUserRoleStatus 处理器内部，避免依赖 user-service 调用方记得失效。 |

---
VERDICT: APPROVED
---
