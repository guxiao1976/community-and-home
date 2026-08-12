# 新会话启动说明：access-data-permission 后续 Wave（阶段 3/4/5/6）

> **给新会话的第一份任务安排。** 启动本会话时,先完整读本文件,然后按 §5 执行顺序逐步推进。
> 技术交接基线详见同目录 [`HANDOFF-NEXT-WAVE.md`](./HANDOFF-NEXT-WAVE.md);权威契约见 [`tasks.md`](./tasks.md) 阶段 3/4/5/6。

---

## 1. 当前基线（已提交,可直接在此之上开发）

| 仓库 | HEAD | 说明 |
|------|------|------|
| 根仓库 | `15f23c2` | Wave 1 交付 + 交接文档 |
| master-data | `a42955e` | 整树缓存跨进程失效 + JWT 认证 |
| api-proto | `c245c09` | 错误码 060007 契约同步 |

**已就绪的基础设施**（后续 Wave 直接复用,不要重复造）:
- `permission-service` 已实现 `AssertPublishScope` RPC(T1.7),community-hub 直接 gRPC 调用
- `master-data` 已实现 `ResolveScopeAncestors`,permission 已内嵌调用
- `user-service` 已有 `PermissionClient`(svc/servicecontext.go:33) + `roleMapper`(role_mapper.go,动态加载 role_code↔role_id)
- `community-hub-service` 已有 `PermClient`(svc/servicecontext.go:17)
- `registered_user`(role_id=9)种子 + browse 权限已在 Wave 1 落地

## 2. 本次任务范围（对应 tasks.md 阶段 3/4/5/6）

| 阶段 | 服务 | 任务 | taskType |
|------|------|------|:---:|
| 3 | user-service | T3.1 CreateUser 自动分配 registered_user / T3.2 JoinCommunity ownership+授权 / T3.3 LeaveCommunity 撤销 / T3.4 门禁 | feature |
| 4 | community-hub-service | T4.0 前置 JWT 通道 / T4.1 publisher_id 取 JWT / T4.2-4.4 AssertPublishScope 挂载 / T4.5 moderation 回调系统身份 / T4.6 读列表 GetDataScopes 过滤 / T4.7 错误码 080006 / T4.8 门禁 | feature |
| 5 | web/mobile | T5.1 joinCommunity 携带 ownership | feature |
| 6 | 集成验收 | T6.1 跨服务端到端矩阵 / T6.2 收尾 | — |

**阶段 3 与阶段 4 无依赖,可并行启动两个 Workflow。** 阶段 5 依赖阶段 3(ownership 由 user 落库),阶段 6 依赖全部。

## 3. 执行前验证清单

```bash
git status                              # 应仅剩无关子模块污染(my-ralph-project/ai-model/moderation)
git log --oneline -1                    # 期望 15f23c2
git -C services/master-data-service log --oneline -1   # a42955e
git -C api-proto log --oneline -1                     # c245c09
# 门禁基线:
bash .harness/skills/qa/scripts/harness-checks.sh --service permission-service   # 16 PASS
bash .harness/skills/qa/scripts/harness-checks.sh --service master-data-service   # 16 PASS
```

## 4. 关键约束（违反即回退）

1. **禁止改 `api-proto/`**（契约已提交 031f4e4 + c245c09,子 Agent 无权限）
2. **禁止改 `common/`**（本变更 common_change_required=false）
3. **taskType: feature 强制 TDD**（RED 摘录必须真实,参考 Wave 1 教训 `[[tdd-red-evidence-requires-fail-excerpt]]`）
4. **依赖方向**:community-hub 不得直连 master-data 做 scope 解析,祖先链仅经 permission `ResolveScopeAncestors` 消费
5. **错误码映射**:permission `060007`(目标小区超出数据范围)→ community-hub 消费时映射为 `080006`
6. **JWT claim 键**:统一 `user_id`,消费用 `util.ExtractUserID`(user_id 优先 + userId 兜底)
7. **registered_user 自动分配**:DB 落库成功后 AssignRole(userId, role_id=9, scope_type='', scope_id=0, status=2),失败仅告警不阻塞注册
8. **读列表过滤**:GLOBAL 不过滤 / LIMITED `IN(ids)` / EMPTY 空列表,须逻辑层实现,SQL IN 不能拼空列表

## 5. 执行顺序

### 5.0 启动 dispatch（首条响应输出工作量分级）

按 CLAUDE.md 硬约束 #7,先调 dispatch 分级。本变更整体为 L 级(跨 3 服务 + 前端),OpenSpec 已定稿,直接进入 N×Workflow 并行:

```
## 工作量分级
- 分级: L（后续 Wave 阶段 3/4/5/6，跨 user/community-hub/web-mobile）
- 路由: N×Workflow 并行（阶段3+4 无依赖）→ 阶段5 → 阶段6 集成
- QA: ✅15项 | Review: ✅3视角 | taskType: feature
- 涉及服务: user-service / community-hub-service / web-mobile
```

### 5.1 启动阶段 3 + 4 两个 Workflow（并行）

**Workflow A — user-service**:
```javascript
Workflow({ scriptPath: ".harness/workflows/harness-pipeline.js",
  args: {
    serviceName: "用户服务", serviceDir: "services/user-service",
    task: "type: feature | access-data-permission 阶段③ user-service T3.1-T3.4（严格按 tasks.md，TDD）"
  }})
```

**Workflow B — community-hub-service**:
```javascript
Workflow({ scriptPath: ".harness/workflows/harness-pipeline.js",
  args: {
    serviceName: "社区枢纽服务", serviceDir: "services/community-hub-service",
    task: "type: feature | access-data-permission 阶段④ community-hub T4.0-T4.8（严格按 tasks.md，TDD）"
  }})
```

> 任务文本建议直接引用 tasks.md 对应阶段全部内容（含文件路径 + TDD 用例矩阵）,确保 Generator 精确执行。

### 5.2 阶段 5 web/mobile（依赖阶段 3,串行）
```javascript
Workflow({ scriptPath: ".harness/workflows/harness-pipeline.js",
  args: {
    serviceName: "移动端", serviceDir: "web/mobile",
    task: "type: feature | access-data-permission 阶段⑤ web/mobile T5.1 joinCommunity 携带 ownership（严格按 tasks.md）"
  }})
```

### 5.3 阶段 6 集成验收（Owner 内联执行）
- `scripts/start.sh` 全栈启动
- 按 tasks.md T6.1 端到端验收矩阵逐步验证:
  注册→无小区发布❌ / 加入小区(自有)→owner+scope出现→发布✅ / 未认证选举❌ / 认证后选举✅ / owner@A发B❌080006 / 抓包改publisher_id❌ / 退出B后立刻发布❌(缓存DEL生效) / 读列表按scope过滤 / moderation回调放行、内容不存在拒绝
- T6.2 收尾:devlog 三层体系 + 各服务门禁全绿 + design 一致性复核

### 5.4 阶段 6 归档（Owner 内联）
- 移动 QA/Review → `.harness/changes/access-data-permission/impl/<service>/`
- 更新 summary.md + INDEX.md
- 处理 Memory 建议

## 6. 注意事项（Wave 1 踩坑,避免重蹈）

1. **TDD 证据必须真实摘录**：行为型断言 RED(如 `assert.Equal`)需 revert 生产代码→跑测试→复制 FAIL 文本→恢复,不能只写"RED→GREEN"。Wave 1 master-data 因此超轮次。
2. **Generator 禁止改 api-proto/**:Wave 1 修复轮曾越界改 proto,已被 Owner 回滚。若 Generator 想改契约,通知 Owner 亲自执行。
3. **错误码先 grep 查重**:新增错误码前 `grep -rn "NewBaseRespWithError" services/<name>/rpc/`,避免同码异义。
4. **迁移先查重**:对既有表加唯一索引前先查重,否则 ALTER 阻塞部署。
5. **敏感权限标 level-2 已做完**(Wave 1),不要重复。
6. **memory 建议落地**:Workflow 返回的 memorySuggestions 需创建记忆文件 + 更新 MEMORY.md + 重建索引。

## 7. 完成后交付

- 各服务门禁 16 PASS / 0 FAIL
- 阶段 3/4/5 代码 + 测试提交到对应仓库
- summary.md 阶段追踪全 ✅ + INDEX 更新
- 跨服务集成测试通过（T6.1 矩阵全绿）
