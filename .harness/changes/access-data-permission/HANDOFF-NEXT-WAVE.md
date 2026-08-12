# 交接文档：Wave 1 完成 → 后续 Wave 执行基线

> 生成时间：2026-08-12
> 目的：Wave 1（阶段① 数据权限核心）已完整交付，后续 Wave 在此基线上执行阶段 3/4/5/6。

## 1. Wave 1 已完成并提交（后续 Wave 基线）

### 提交记录

| 仓库 | Commit | 说明 |
|------|--------|------|
| 根仓库 | 4c31166 | permission-service T1.1-T1.8 + need_human 修复 + 归档(summary/INDEX/impl) + 4 条新记忆 |
| master-data | 49d476e | 整树缓存跨进程失效 + JWT 认证 + Review CRITICAL 修复 |
| master-data | a42955e | graph-context 刷新 |
| api-proto | c245c09 | permission 错误码 060007 契约同步（仅注释，非破坏） |

### 门禁状态
- permission-service：16 PASS / 0 FAIL / 0 WARN
- master-data-service：16 PASS / 0 FAIL / 0 WARN
- 全链路编译（permission/master-data/community-hub/user）：build exit 0

### 阶段 0-2 完成内容（权威契约已就绪）
- 阶段 0 Proto：031f4e4（min_verf_level/DataScopeState/ScopeRef/AssertPublishScope/ResolveScopeAncestors/CommunityOwnership）+ c245c09（错误码 060007 注释同步）
- 阶段 1 permission：T1.1-T1.8 全部落地
- 阶段 2 master-data：T2.1-T2.3 全部落地（T2.1 复用交接基线实现）

## 2. 后续 Wave 执行范围（tasks.md 阶段 3/4/5/6）

### 阶段 3 · user-service（编排，非权威）— Task 3.1-3.4
- **3.1**: CreateUser 自动分配 registered_user（DB 落库成功后 AssignRole(userId, registered_user, scope='', 0, status=2)，失败仅告警）
- **3.2**: JoinCommunity ownership + 自动授权（校验 ownership∈{OWNED,RENTED}，落库后 AssignRole(owner|tenant, 'community', community_id)）
- **3.3**: LeaveCommunity 撤销授权（双调 RevokeRole owner+tenant 幂等，失败恢复 bind_status）
- **3.4**: 门禁

### 阶段 4 · community-hub-service（消费方）— Task 4.0-4.8
- **4.0**: 前置确认 REST JWT 身份注入通道 + PermMiddleware 顺序
- **4.1**: publisher_id 取 JWT（覆盖客户端 body 值）
- **4.2-4.4**: AssertPublishScope 挂载（lostfound/notice/UpsertContacts 写接口）
- **4.5**: moderation 回调以系统身份（system_user_id=0）校验
- **4.6**: 读列表按 GetDataScopes 过滤（filterByScope）
- **4.7**: 错误码 080006 注册（permission 060007→080006 映射）
- **4.8**: 门禁

### 阶段 5 · 移动端（web/mobile）— Task 5.1
- **5.1**: joinCommunity 携带 ownership + 表单补全（自有/租住选择必填）

### 阶段 6 · 集成验收 — Task 6.1-6.2
- **6.1**: 跨服务端到端验收矩阵（注册→无小区发布❌ / 加入小区→发布✅ / 未认证选举❌ / owner@A发B❌080006 / 抓包改publisher_id❌ / 退出后缓存DEL生效 / 读列表按scope过滤 / moderation回调放行）
- **6.2**: 收尾（devlog 三层体系 + 各服务门禁全绿 + design 一致性复核）

## 3. 交接时需注意

1. **permission 已具备 AssertPublishScope RPC**（T1.7 实现），community-hub 消费时直接复用 gRPC 客户端（`zrpc.MustNewClient(c.PermissionRpc)` → `permissionv1.NewPermissionServiceClient`），无需改 permission。
2. **master-data ResolveScopeAncestors 已就绪**，permission AssertPublishScope 已内嵌调用（祖先链 ∩ ids）。
3. **错误码映射**：permission 060007（目标小区超出数据范围）→ community-hub 消费时映射为 080006。
4. **registered_user 自动分配**依赖 `roleMapper`（user-service 既有）解析 role_id=9；roleMapper 需含 registered_user 映射。
5. **JWT claim 键**：统一 `user_id`（auth-service 签发），consumption 用 `util.ExtractUserID`（user_id 优先 + userId 兜底）——已在 permission/master-data 验证。
6. **敏感权限已标 level-2**（user:read/moderation:read 等），后续 Wave 无需重复标注。
7. **community-hub 读列表过滤**：GLOBAL 不过滤 / LIMITED `IN(ids)` / EMPTY 空列表，须在逻辑层实现（不能用 SQL IN 拼空列表）。
8. **依赖方向**：community-hub 不得直连 master-data 做 scope 解析，祖先链仅经 permission `ResolveScopeAncestors` 消费。

## 4. 后续 Wave 执行入口

```bash
# 按 CLAUDE.md 约束：先 dispatch Skill 做入口判定与工作量分级
# 变更文档：.harness/changes/access-data-permission/
# 阶段：③ user-service(T3.1-T3.4) + ④ community-hub(T4.0-T4.8) 可并行
#        ⑤ web/mobile(T5.1) → ⑥ 集成验收
# 任务类型：feature(taskType:feature, 强制 TDD)
# 注意：子 Claude 禁止改 api-proto/（契约已提交 031f4e4 + c245c09）
```

## 5. 遗留待办（非 access-data-permission 范围）

- master-data REST JWT 历史遗留已修复（10 路由组 WithJwt），但完整认证策略评估待后续
- permission ListPermissions/InvalidateUserCache 透传无直接单测（QA WARNING，非阻塞）
- ai-model 双模块结构 harness-checks 适配（交接基线遗留）
- common permission.go userID==0 fail-open → Owner 收紧为 fail-closed（交接基线遗留）

## 6. 验证清单（后续 Wave 启动后确认）

- [ ] 根仓库 `git log --oneline -1` = 4c31166
- [ ] master-data `git log --oneline -1` = a42955e
- [ ] api-proto `git log --oneline -1` = c245c09
- [ ] permission-service 门禁 16 PASS（`harness-checks.sh --service permission-service`）
- [ ] master-data-service 门禁 16 PASS
