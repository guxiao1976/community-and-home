# 核心业务流程

> 从各服务 design.md 提取的端到端业务流程。Agent 编码前理解业务上下文用。
> 最后更新: 2026-06-10

## 一、用户入驻全链路

这是项目最核心的端到端流程，横跨 auth / user / permission 三个服务。

```
注册 → 加入小区 → 申请角色 → 提交认证 → 管理员审核 → 角色生效 → 权限激活

auth-service:    ①注册        ────────────────────────────────→ ⑦签发JWT(带角色)
user-service:    ①创建用户    ②加入小区  ③申请角色  ④提交认证  ⑤审核  ⑥角色生效
permission:                                                             ⑦权限匹配
```

**详细步骤**:

1. **注册** (auth-service + user-service): SMS 验证码 → `CreateUser(phone, nickname)` → user 表插入（AES 加密手机号、status=1、credit_score=100）→ bcrypt 密码 → 签发 AT+RT
2. **加入小区** (user-service): `JoinCommunity(user_id, community_id)` → 检查上限（最多 5 个）→ INSERT membership(bind_status=1) → 首个小区设为默认
3. **申请角色** (user-service): `ApplyRole(user_id, community_id, role='owner', building, unit, room)` → 查 membership 有效 → 查无重复 → INSERT role(verf_status=0)
4. **提交认证** (user-service): `SubmitCertification(role_id, documents, real_name, id_card, building, unit, room)` → verf_status ∈ {0,3,4} → AES 加密身份证 → JSON 存储 → INSERT certification(status=1) → role verf_status=1
5. **管理员审核** (user-service): `ReviewCertification(cert_id, result=approve/reject)` → 审批通过 → 回填 real_name/id_card → INSERT residence → role verf_status=2 → 通过 → role verf_status=3（可重新提交）

**角色过期**: 定时任务（每日 0 点）→ `verf_status=2 + expires_at<now` → verf_status=4 → 用户重新认证

## 二、认证与 Token 生命周期

```
登录(密码/SMS) → AT(15min)+RT(15day) → API访问(Gateway验证AT+查权限)
                    ↓ AT过期
                 RT刷新(RT轮转,新AT+新RT)
                    ↓ 登出
                 AT黑名单 + RT删除
```

- **密码登录**: RSA 解密 → bcrypt 验密 → 查用户状态 → 查角色(Cache-Aside Redis) → 签发 AT+RT
- **SMS 登录**: 验证码从 Redis 取后立即删除（防重放）→ 其余同密码登录
- **Token 刷新**: 旧 RT 验证 → Lua 原子轮转（删旧写新）→ 重新查角色（5min 内角色变更生效）
- **登出**: AT 加入黑名单(TTL=剩余有效期) + 删除当前设备 RT + 可选踢所有设备

## 三、内容审核管线

```
文本/图片 → 规范化 → AC自动机(敏感词) → 白名单过滤 → 小模型(Ollama) → 大模型(RemoteLLM)
              │                    │              │               │
              └─ 命中severity=1 ──→ 直接拒绝      └─ confidence≥0.9 → 最终判定
```

- **combined 模式**: 全管线（规范化→AC→白名单→小模型→大模型）
- **ac_only 模式**: 仅关键词匹配
- **敏感词同步**: 每 5 分钟从 master-data-service 拉取（gRPC 优先，直读 DB 兜底），版本号增量更新
- **输出**: `{is_safe, risk_level, need_review, check_layer}`

## 四、主数据审批工作流

适用于 4 种实体类型（行政区划、小区、系统配置、敏感词），统一状态机：

```
Draft(0) → Pending(1) → Approved(2) / Rejected(3)
              ↑              │              │
              └── resubmit ──┘              │
                                             ↓
                              DeleteRequest → Draft(0) → Pending(1) → [已删除]
                                                         ↑
                                          CancelDelete ──┘
```

- 修改时保存 `change_snapshot` JSON，驳回时回滚
- 敏感词变更 → moderation-service 5 分钟内同步

## 五、RBAC 权限检查

```
请求 → Gateway → ValidateToken(auth) → CheckPermission(perm)
                                            │
                         Redis SISMEMBER ←──┴── MISS → DB查角色/权限 → 写Redis(30min)
                                            │
                                     HIT → allowed=true
```

- 系统角色(is_system=1)直接放行
- 角色权限变更 → 批量删 Redis 缓存 → 下次请求重建
- 数据权限：`GetDataScopes` 返回 scope_id 列表 → 调用方 `WHERE x IN (scope_ids)`

## 六、文件上传/下载

```
上传: 客户端 → 请求预签名URL → 服务端返回PUT URL(15min有效) → 客户端直传MinIO → ConfirmUpload写DB
下载: 客户端 → 请求文件 → 服务端生成GET URL(1h有效) → 客户端直取MinIO
删除: 软删除(MinIO对象不物理删除)
```

## 七、社区枢纽（通知/联络/寻失）

- **通知**: 创建 → 列表(置顶优先) → 详情 → 软删除。角色验证和审核集成尚未实现
- **联络**: 全量替换式更新 → 分类展示 → 一键拨号
- **寻失**: 发布(active) → 列表(按类型筛选+分页) → 标记解决(resolved)

## 八、已知空白

| 空白 | 影响 | 说明 |
|------|------|------|
| 通知/寻失无审核集成 | 🔴 高 | community-hub 和 moderation 未对接 |
| 人工审核 stub | 🟡 中 | moderation SubmitReview 未实现 |
| 图片哈希 stub | 🟡 中 | 违规图片库比对未实现 |
| permission vs user 双角色体系未梳理 | 🟡 中 | sys_role 和 user_membership_role 关系不明 |
| MQ 审批事件未实现 | 🟢 低 | master-data outbox 仍为 stub |
| community-dynamics 和 community-hub 重复 | 🔴 高 | 两个规格文档定义了相同功能（通知/联络/寻失） |
