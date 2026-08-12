# Plan Review — access-data-permission（structure 结构合理性视角）v2

**审查维度**: 服务归属 / 依赖方向 / Proto 变更识别 / 依赖顺序 / spec 间职责重叠冲突
**审查时间**: 2026-08-12
**审查输入**: `request.md` + `proposal.md` + `specs/{scope-model,capability-layering,registered-user,join-auto-authorization,publish-validation}/spec.md` + 设计依据 `docs/specs/access-control-design.md`(§1.4/§3/§5/§8) + `rbac-design.md` §2.5 + 上轮 `spec_review_structure_v1.md`
**对照项**: 现网 `api-proto/api/permission/v1/permission.proto`、`api-proto/api/masterdata/v1/masterdata.proto`、`api-proto/api/user/v1/user.proto`
**复审重点**: 上轮 SHOULD FIX ①②落实核对（祖先链消费方向 / Proto 变更清单补全）

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 8

## 上轮 SHOULD FIX 落实情况核对

| # | 上轮问题 | 落实状态 | 证据 |
|---|---------|:---:|------|
| ① | 祖先链 RPC 消费方向未钉死，design §8 与 spec 冲突（community-hub 直读 master-data） | ✅ 落实（spec+proposal 层） | `specs/scope-model` REQ-1.4 显式钉死「consumption direction SHALL be fixed as community-hub → permission-service → master-data」并声明 community-hub-service SHALL NOT call master-data directly；`proposal.md`「影响范围」master-data 行同步钉死「消费方向钉死为 community-hub → permission-service → master-data」。判据封装回 permission-service 权威，跨服务不重复实现统一规则 ✅ |
| ② | Proto 变更清单不完整（仅 3 项，漏 min_verf_level 透出 / GetDataScopes 三态 / ScopeRef / 祖先链 RPC req-resp） | ✅ 落实 | `proposal.md`「Proto 变更」节扩为 5 项：min_verf_level 透出 / GetDataScopes 三态可区分 / ScopeRef message / AssertPublishScope RPC / ResolveScopeAncestors RPC；`.change.yaml` `proto_changes:` 同步 5 项。对照现网 proto 核实：`GetDataScopesResponse` 现为 `repeated int64 scope_ids`（无法表达 global/空），全 api-proto 无 `ScopeRef`/`AssertPublishScope`/`ResolveScopeAncestors`/`min_verf_level` —— 5 项均为真实增量、无破坏性变更 ✅ |
| ③ | `.change.yaml common_change_required: false` vs proposal「common 可能」声明不一致 | ❌ 未落实 | `.change.yaml` 仍为 `false`，proposal 影响范围表仍写「可能（需全局评估）」+ 风险「提交前全局评估 + 门禁」。见 SHOULD FIX #2 |
| ④ | 多角色 scope 合并优先级未定义 | 🔶 部分落实 | `capability-layering` REQ-2.2 场景「registered_user（空范围）参与命中」确立 empty 不降级、不阻塞 limited；REQ-1.7 确立 global 放行。但「global 支配 / limited 并集 / empty」完整合并优先级仍未以 REQ 显式定义。见 SHOULD FIX #3 |

## 服务归属与依赖方向总评（符合）

| 职责 | 归属 | 判定 |
|------|------|:---:|
| scope 三态 / 祖先链统一判据 / GetDataScopes / AssertPublishScope / min_verf_level / registered_user 角色定义 / 缓存失效 | permission-service（权威） | ✅ |
| CreateUser 自动分配 registered_user / Join/Leave 编排 Assign/RevokeRole（非权威） | user-service | ✅ |
| 写接口挂 AssertPublishScope / publisher_id 取 JWT / 读列表按 GetDataScopes 过滤 | community-hub（消费方） | ✅ |
| 「scope 节点 → 祖先链」RPC / 行政区划整树缓存 | master-data | ✅ |
| 全部 Proto 变更（permission/masterdata 两个 package） | api-proto（阶段 4 全局 Owner） | ✅ |

- 依赖方向 `community-hub → permission-service → master-data` 无环、指向权威；`user-service → permission-service` 复用现有 `AssignRole`/`RevokeRole`。核实现网 `AssignRoleRequest` 已含 `scope_type/scope_id/status(0-4)/verified_at/expires_at`——join 自动授权（owner/tenant + community scope + status=0）与注册自动分配（registered_user + status=2）均可直接复用，无需新增基础设施 RPC ✅。
- 覆盖树拓扑（division/residential_area）与授权语义分离干净：master-data 管数据、permission-service 管判定，无职责重叠 ✅。

## 发现

### 🔴 MUST FIX

无。

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | `docs/specs/access-control-design.md` §8 L273 | **设计依据与钉死的方向仍冲突（上轮①的残留）**。spec REQ-1.4 与 proposal 已把祖先链消费钉死为 community-hub → permission-service → master-data，但权威设计依据 §8 依赖清单仍写 `community-hub ──读配额/祖先链──▶ master-data`（L273），与钉死方向直接矛盾。request.md 声明设计依据「已定稿并提交」，阶段 3 架构设计者按 §8 字面实现有重新拆散统一判据的风险。 | 阶段 3 修订 design §8 L273：该行去掉「祖先链」，仅保留未来 §7 配额（`community-hub ──读配额──▶ master-data`）；祖先链解析明确只归 permission-service 消费。建议在 stage 3 tasks.md 的架构设计 task 描述中显式列出此修订点，防止漏改。 |
| 2 | `.change.yaml` `common_change_required: false` vs `proposal.md` 影响范围「common 可能（需全局评估）」 | **上轮③未落实，声明仍不一致**。proposal 明确留了「AssertPublishScope 客户端辅助 / ScopeRef 类型可能入 common 客户端库」的口子并依赖「提交前全局评估 + 门禁」（硬约束 #6），但 yaml 声明 false，pipeline 可能据此跳过全局评估门禁。 | 二选一：若确实要动 common → 改 `common_change_required: true`；若不动 → 删除 proposal 中「可能入 common」表述，改为「ScopeRef/AssertPublishScope 客户端经 gen/go 生成代码消费，本变更不触碰 common/」（更干净，因为 ScopeRef 与 RPC 桩都在 api-proto，无需 common 包装）。 |
| 3 | `specs/scope-model` REQ-1.1/1.7 | **多角色 scope 合并优先级仍未以 REQ 显式定义（上轮④残留）**。REQ-1.1 说 per scope_type「exactly one of three states」但未定义从多条 grant 收敛到单一状态的计算规则；REQ-1.7 只给「global 放行」，cap-layering REQ-2.2 场景只覆盖 empty+limited。global+limited 并存、多 limited 取并集等跨状态优先级未写死，GetDataScopes（读）与 AssertPublishScope（写）共用同一判据，若实现期两处各自猜测会不一致。 | 在 scope-model 增加一条 REQ 显式定义合并优先级：该 scope_type 下任一 grant 为 global → 有效状态 global；无 global 但存在 limited grant → limited（取并集）；否则 empty。并各配一场景（审核员兼业主 = global+limited → global 生效）。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | spec 伪签名 `targets []scopeRef`（publish-validation REQ-5.1）与契约名 `ScopeRef` 大小写不一致，仅伪代码展示，阶段 4 落 proto 时统一为 `ScopeRef`。 |
| 2 | publish-validation REQ-5.3 把 `UpdateNoticeModerationStatus`/`UpdateLostFoundModerationStatus` 列为受校验的写操作。此类回调若由 moderation-service 服务间发起，其 JWT 身份为服务账号而非内容作者——REQ-5.6「userId 取 JWT」下需确认该身份持有覆盖目标小区的数据范围（或为 global 审核员）。阶段 3 明确回调发起方身份与 scope 归属。 |
| 3 | GetDataScopes 三态表达的具体响应形状（state 枚举 + optional ids / oneof）与「global」作为 scope_type 取值 vs 独立维度，留给阶段 3/4；contract 层只需保证 GetDataScopes 与 AssertPublishScope 走同一合并规则（呼应 SHOULD FIX #3）。 |
| 4 | 实施顺序建议：阶段 4 Proto（ResolveScopeAncestors 是 permission-service 前置）→ permission-service + master-data（权威，可并行）→ user-service + community-hub（编排/消费，可并行）→ 集成。阶段 3 tasks.md 需编码该依赖序。 |
| 5 | master-data 已有 `GetDivisionPath`/`GetDivisionTree`/`GetResidentialArea`（community_div_id），新 `ResolveScopeAncestors` 与之部分重叠。阶段 3 评估复用/组合 vs 新增，避免双套祖先解析。 |
| 6 | 阶段 4 落 proto 时给 `AssertPublishScope`/`ResolveScopeAncestors` 标 `@auth: INTERNAL`、`@idempotent: true`，与现网 `CheckPermission`/`GetDataScopes` 一致。 |
| 7 | scope-model REQ-1.6 读过滤波及 community-hub 所有读/列表接口，爆炸面大。阶段 3/5 应明确挂载端点清单，防编码期范围失控（承接上轮 INFO #3）。 |
| 8 | 缓存失效在 scope-model REQ-1.8 / cap-layering REQ-2.5 / join REQ-4.4 重复出现但触发源不同；建议阶段 3 把失效收敛到 AssignRole/RevokeRole/UpdateUserRoleStatus 处理器内部，避免依赖 user-service 调用方记得失效（承接上轮 INFO #7）。 |

---
VERDICT: APPROVED
---
