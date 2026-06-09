# 「用户认证全流程」功能测试计划

## 测试信息

- **测试范围**：用户注册 → 加入小区 → 选楼号/房号 → 申请角色 → 提交认证（房产证） → 审核通过 → JWT 签发与刷新 → 全流程
- **涉及服务**：auth-service、user-service
- **测试策略**：分层独立测试（Mock gRPC 依赖）
- **测试环境**：miniredis（Go 内存 Redis）、SQLite 内存库 或 Mock DB
- **设计文档参考**：
  - [auth-design.md](auth-design.md) — 认证中心设计方案
  - [user-design.md](user-design.md) — 用户中心设计方案
  - [auth.md](auth.md) — 认证服务简要规格
  - [user.md](user.md) — 用户服务简要规格
- **编写日期**：2026-06-03

---

## 一、测试策略

### 1.1 测试分层

```
┌─────────────────────────────────────────────┐
│ 六、端到端串联用例（关键路径冒烟）             │
├─────────────────────────────────────────────┤
│ 五、JWT 与角色数据范围专项                    │
├─────────────────────────────────────────────┤
│ 四、Redis 专项测试                           │
├──────────────────┬──────────────────────────┤
│ 二、auth-service  │ 三、user-service          │
│     测试用例      │     测试用例              │
├──────────────────┴──────────────────────────┤
│ Mock 层：miniredis / Mock gRPC Client        │
└─────────────────────────────────────────────┘
```

### 1.2 测试范围

| 范围 | 内容 |
|------|------|
| **要测** | Register、Login、LoginSms、RefreshToken、Logout、ValidateToken、JoinCommunity、ApplyRole、SubmitCertification、ReviewCertification、GetUserRoles、CheckAccess、BindResidence |
| **要测** | Redis 缓存读写/失效、JWT claims 结构（roles/r/c）、角色数据范围（community_id） |
| **不测** | API Gateway 鉴权逻辑（配置层面）、短信发送（第三方）、文件上传（file-service）、approval-service（审批流）、定时任务（Cron） |

### 1.3 Mock 策略

```
auth-service 测试：
  ├── user-service gRPC → Mock（可控返回各种成功/失败场景）
  ├── Redis → miniredis（真实 Redis 语义，内存运行）
  └── MySQL → Mock DB（通过 mock model 层）

user-service 测试：
  ├── Redis → miniredis
  └── MySQL → Mock DB 或 SQLite 内存库
```

### 1.4 测试角色矩阵

| 角色 | role_code | community_id | 绑定小区 | 有期限 | 需要房屋 |
|------|-----------|:---:|:---:|:---:|:---:|
| 普通用户(无角色) | — | — | — | — | — |
| 业主 | `owner` | 小区ID | ✅ | 永久 | ✅ |
| 租户 | `tenant` | 小区ID | ✅ | 租约到期 | ✅ |
| 网格员 | `grid_worker` | 小区ID | ✅ | 1年 | — |
| 社区管理员 | `community_admin` | 小区ID | ✅ | 2年 | — |
| 物业管理员 | `property_admin` | 小区ID | ✅ | 1年 | — |
| 业委会 | `committee` | 小区ID | ✅ | 2年 | — |
| 商家 | `merchant` | 0（全局） | ❌ | 永久 | — |

---

## 二、auth-service 测试用例

### 2.1 注册 (Register)

参考：[auth-design.md §3.3](auth-design.md) | Proto: `auth.v1.AuthService.Register`

| 编号 | 场景 | 输入 | 预期结果 | Redis 验证 |
|------|------|------|---------|-----------|
| **A-R-01** | 正常注册（带密码） | `phone=13800138000, sms_code=123456, encrypted_password=<RSA密文>, nickname=张三, device_id=web_001` | `code=0, user_id>0, access_token 非空, refresh_token 非空, expires_at=now+900` | `sms:code:13800138000` 已删除；`auth:rt:{uid}:web_001` 存在且值=jti(TTL=1296000)；`auth:rt:{uid}:devices` 含 `web_001` |
| **A-R-02** | 正常注册（无密码，仅短信登录） | `phone=13800138001, sms_code=654321, device_id=ios_001, encrypted_password=""` | `code=0`，credential 表中 bcrypt 为空字符串 | 同 A-R-01 |
| **A-R-03** | 验证码已过期（Redis 中不存在） | sms_code 正确但 Redis key 不存在 | `code=50004, msg="验证码已过期，请重新获取"` | `sms:code` key 本就不存在 |
| **A-R-04** | 验证码错误 | sms_code 与 Redis 中的值不匹配 | `code=50004, msg="验证码错误"` | `sms:code` 未被删除（仍然存在） |
| **A-R-05** | 手机号已注册 | phone 已存在 `user_base` 中 | `code=100002`（user-service 返回） | — |
| **A-R-06** | RSA 密码解密失败 — Saga 补偿 | `encrypted_password` 格式错误 | `code=500400, msg="密码格式错误"` | user_base 被软删除(`status=3`) |
| **A-R-07** | Credential 写入 DB 失败 — Saga 补偿 | Mock `CredentialModel.Insert` 返回错误 | `code=100002` | user_base 被软删除(`status=3`) |
| **A-R-08** | 空手机号 | `phone=""` | `code=50004` | — |
| **A-R-09** | 新用户 JWT AT 的 roles 为空数组 | 正常注册后解析 AT | `AT.roles == []`（空数组，非 null） | — |
| **A-R-10** | 注册时 GetUserRoles 调用失败不阻塞 | Mock `GetUserRoles` 返回 gRPC 错误 | `code=0`，注册成功，AT 的 `roles=[]` | — |
| **A-R-11** | 双 Token 签名密钥隔离 | 正常注册 | AT 用 `AccessSecret` 验证通过；RT 用 `RefreshSecret` 验证通过；交叉验证失败 | — |
| **A-R-12** | 设备和 JTI 格式 | 正常注册 | `jti` 格式为 `{user_id}-{unix_nano}`；`device_id` 正确写入 RT claims | — |

### 2.2 账密登录 (Login)

参考：[auth-design.md §3.1](auth-design.md) | Proto: `auth.v1.AuthService.Login`

| 编号 | 场景 | 预期结果 | JWT / Redis 验证 |
|------|------|---------|-----------------|
| **A-L-01** | 正常登录（用户有 owner 角色） | `code=0, AT+RT+user_id+expires_at` | AT claims: `user_id, jti, roles=[{r:"owner",c:1001}], exp, iat`；RT 写入 Redis `auth:rt:{uid}:{device}` |
| **A-L-02** | 正常登录（用户有多角色） | `code=0` | AT roles 包含全部已认证角色：`[{r:"owner",c:1001},{r:"committee",c:1001}]` |
| **A-L-03** | 手机号未注册（凭证不存在） | `code=50001, msg="手机号或密码错误"` | — |
| **A-L-04** | 密码错误 | `code=50001, msg="手机号或密码错误"` | — |
| **A-L-05** | RSA 解密 phone 失败 | `code=50001` | — |
| **A-L-06** | RSA 解密 password 失败 | `code=50001` | — |
| **A-L-07** | 账号被禁用（status=2）| Mock user-service 返回 `status=2` | `code=50005`（当前代码未实现此校验——**待确认**）。注：当前 loginlogic.go 调用了 `GetUserByPhone` 但未校验 status |
| **A-L-08** | GetUserRoles gRPC 调用失败 | Mock `GetUserRoles` 返回错误 | `code=509504, msg="获取用户角色失败"` |
| **A-L-09** | GetUserByPhone gRPC 调用失败 | Mock `GetUserByPhone` 返回错误 | `code=509503, msg="获取用户信息失败"` |
| **A-L-10** | 空手机号 | `encrypted_phone=""` | `code=500400` |
| **A-L-11** | 空密码 | `encrypted_password=""` | `code=500400` |
| **A-L-12** | RT 持久化成功 | 登录成功 | `auth:rt:{uid}:{device}` = jti, TTL=1296000 秒 |
| **A-L-13** | 设备集合更新 | 登录成功 | `auth:rt:{uid}:devices` Set 包含当前 device_id |
| **A-L-14** | merchant 角色 c=0 | 商家用户登录 | AT roles 中 merchant 条目：`{r:"merchant",c:0}` |

### 2.3 短信验证码登录 (LoginSms)

参考：[auth-design.md §3.2](auth-design.md) | Proto: `auth.v1.AuthService.LoginSms`

| 编号 | 场景 | 预期结果 | Redis 验证 |
|------|------|---------|-----------|
| **A-S-01** | 正常短信登录 | `code=0, AT+RT` | sms:code 校验后已删除；AT roles 正确 |
| **A-S-02** | 验证码错误 | `code=50004` | sms:code 未被删除 |
| **A-S-03** | 验证码已过期 | `code=50004` | — |
| **A-S-04** | 手机号未注册（无凭证） | `code=50001` | sms:code 已删除（先校验验证码） |
| **A-S-05** | 空手机号 | `code=50004` | — |
| **A-S-06** | roles 正确写入 AT | 有角色的用户登录 | AT claims 包含 `roles` 数组 |

### 2.4 Token 刷新 (RefreshToken)

参考：[auth-design.md §3.4](auth-design.md) | Proto: `auth.v1.AuthService.RefreshToken`

| 编号 | 场景 | 预期结果 | Redis / JWT 验证 |
|------|------|---------|-----------------|
| **A-T-01** | 正常刷新 | `code=0`，返回新 AT+新 RT | 旧 RT 已删除，新 RT 已写入；新 AT jti ≠ 旧 AT jti |
| **A-T-02** | RT 已过期（Redis 中 key 不存在） | `code=50003, msg="RT已失效，请重新登录"` | — |
| **A-T-03** | RT jti 不匹配（已被旋转，即 RT 泄露场景） | `code=50003` | — |
| **A-T-04** | RT 签名无效（用错误 secret 签发的 RT） | `code=50003` | — |
| **A-T-05** | RT 格式错误（非 JWT） | `code=50003` | — |
| **A-T-06** | **刷新后角色更新**：用户认证通过后刷新 Token | 新 AT 包含最新角色 | 角色从缓存/gRPC 重新拉取，不再使用旧 AT 的快照 |
| **A-T-07** | **角色被撤销后刷新**：DEL `auth:roles:{uid}` 后刷新 | 新 AT 不包含被撤销的角色 | 缓存 MISS → gRPC 穿透 → 获取最新角色 |
| **A-T-08** | Lua 原子旋转：并发刷新安全性 | 同一 RT 并发 RefreshToken，仅一次成功 | Lua 脚本 CAS：`GET + compare + SET` 原子。只有第一个请求的 oldJti 匹配，其余返回 nil |
| **A-T-09** | 刷新时 GetUserRoles 失败 | Mock `GetUserRoles` 返回 gRPC 错误 | `code=509504`，旧 RT 在 Lua 中已被删除（⚠️ 潜在问题：旋转已执行但角色拉取失败，用户 Token 丢失） |
| **A-T-10** | RT TTL 重置为完整 15 天 | 旧 RT 还剩 1 天时刷新 | 新 RT TTL = 1296000 秒（完整 15 天） |

### 2.5 注销 (Logout)

参考：[auth-design.md §3.5](auth-design.md) | Proto: `auth.v1.AuthService.Logout`

| 编号 | 场景 | 预期结果 | Redis 验证 |
|------|------|---------|-----------|
| **A-O-01** | 正常注销 | `code=0` | `auth:at:blacklist:{jti}=1` (TTL=AT剩余有效期)；`auth:rt:{uid}:{device}` 已删除 |
| **A-O-02** | 注销后 AT 加入黑名单 | `ValidateToken` 返回 `valid=false` | `auth:at:blacklist:{jti}` 存在 |
| **A-O-03** | 注销后 RT 不可刷新 | `RefreshToken` 返回 `50003` | `auth:rt:{uid}:{device}` 已删除 |
| **A-O-04** | 强踢全设备 (`kick_all_devices=true`) | `code=0` | 所有 `auth:rt:{uid}:*` 已删除（通过 `auth:rt:{uid}:devices` SMEMBERS → 批量 DEL） |
| **A-O-05** | AT 黑名单 TTL = AT 剩余有效期 | 注销时 AT 还剩 10 分钟过期 | `auth:at:blacklist:{jti}` TTL = 600 秒 |
| **A-O-06** | 黑名单 TTL 到期自动清除 | 等待 TTL 过期 | `auth:at:blacklist:{jti}` key 不存在，AT 自然过期而非主动拉黑 |

### 2.6 Token 验证 (ValidateToken)

参考：[auth-design.md §3.6](auth-design.md) | Proto: `auth.v1.AuthService.ValidateToken`

| 编号 | 场景 | 预期结果 |
|------|------|---------|
| **A-V-01** | 有效 AT 验证 | `valid=true, user_id=正确值, expires_at=正确值` |
| **A-V-02** | AT 已过期 | `valid=false` |
| **A-V-03** | AT 在黑名单中 | `valid=false` |
| **A-V-04** | AT 签名无效（篡改） | `valid=false` |
| **A-V-05** | 空 AT | `valid=false` |
| **A-V-06** | 格式错误的 AT（非 JWT） | `valid=false` |

---

## 三、user-service 测试用例

### 3.1 加入小区 (JoinCommunity)

参考：[user-design.md §3.2](user-design.md) | Proto: `user.v1.UserService.JoinCommunity`

| 编号 | 场景 | 输入 | 预期结果 |
|------|------|------|---------|
| **U-J-01** | 正常加入小区（首次） | `user_id=1001, community_id=2001` | `code=0`，membership 创建，`bind_status=1`；`user_base.preferences.default_community_id=2001` |
| **U-J-02** | 已加入同一小区（幂等） | 重复调用 U-J-01 | `code=0`，返回已有 membership，不创建新记录 |
| **U-J-03** | 退出后重新加入 | 先 LeaveCommunity 再 JoinCommunity | `code=0`，`bind_status` 重新置为 1，`leave_time` 清除 |
| **U-J-04** | 已达 5 个小区上限 | 用户已加入 5 个小区 | `code=100006, msg="最多加入 5 个小区"` |
| **U-J-05** | 首次加入设置 default_community_id | 用户 preferences 中无 default_community_id | preferences 自动设置为当前 `community_id` |
| **U-J-06** | 已有 default_community_id 不覆盖 | 用户 preferences 已有 default_community_id=1001 | 加入小区 2001 后，default_community_id 仍为 1001 |
| **U-J-07** | 用户不存在 | `user_id=99999` | membership 创建成功（user_id 仅作外键，由调用方保证存在）⚠️ 当前实现无 user 存在性校验 |

### 3.2 申请角色 (ApplyRole)

参考：[user-design.md §3.3](user-design.md) | Proto: `user.v1.UserService.ApplyRole`

| 编号 | 场景 | 输入 | 预期结果 |
|------|------|------|---------|
| **U-A-01** | 申请业主角色（owner） | `user_id, community_id=2001, role_code=owner, building=3, unit=2, room=1501` | `code=0`，role 创建，`verf_status=0, membership_id 有值, community_id=2001`。⚠️ building/unit/room 不在 ApplyRole 时存储 |
| **U-A-02** | 申请租户角色（tenant） | 同上，role_code=tenant | `code=0, verf_status=0` |
| **U-A-03** | 申请商家角色（merchant） | `role_code=merchant` | `code=0, membership_id=NULL, community_id=0` |
| **U-A-04** | 未加入小区直接申请角色 | 无 membership 记录 | `code=100005, msg="小区成员关系不存在或已退出"` |
| **U-A-05** | 已退出小区后申请角色 | membership.bind_status=0 | `code=100005` |
| **U-A-06** | 同一角色重复申请 | 已有 owner 角色 | `code=100008, msg="该角色已存在"` |
| **U-A-07** | 同一小区申请不同角色（先 owner 再 committee） | role_code=committee | 需确认：`uk_member_role (membership_id, role_code)` 阻止相同 role_code，但不同 role_code 应允许。——**待确认当前实现** |
| **U-A-08** | 申请网格员角色 | `role_code=grid_worker` | `code=0, verf_status=0` |
| **U-A-09** | 申请社区管理员角色 | `role_code=community_admin` | `code=0, verf_status=0` |
| **U-A-10** | 申请物业管理员角色 | `role_code=property_admin` | `code=0, verf_status=0` |
| **U-A-11** | 申请业委会角色 | `role_code=committee` | `code=0, verf_status=0` |

### 3.3 提交认证材料 (SubmitCertification)

参考：[user-design.md §3.4](user-design.md) | Proto: `user.v1.UserService.SubmitCertification`

| 编号 | 场景 | 输入 | 预期结果 |
|------|------|------|---------|
| **U-S-01** | 正常提交业主认证（含房产证 URL） | `role_id, user_id, document_urls=["http://minio/deed.jpg"], real_name=张三, id_card_number=110101..., building=3, unit=2, room=1501` | `code=0, certification created (status=1)`；role.verf_status 0→1；document_urls 字段存储完整 JSON（urls+real_name+AES(id_card)+building/unit/room） |
| **U-S-02** | 重新提交（之前被驳回 verf_status=3） | 同上 | `code=0`，INSERT 新 certification 记录；role.verf_status 3→1 |
| **U-S-03** | 重新提交（之前已过期 verf_status=4） | 同上 | `code=0`，INSERT 新 certification；role.verf_status 4→1 |
| **U-S-04** | 待审核中重复提交（verf_status=1） | role.verf_status=1 | `code=100003, msg="该角色已提交认证申请，请勿重复提交"` |
| **U-S-05** | 已通过重复提交（verf_status=2） | role.verf_status=2 | `code=100003` |
| **U-S-06** | role 不存在 | role_id=99999 | `code=100007` |
| **U-S-07** | 身份证号 AES 加密存储 | 正常提交 | `document_urls` JSON 中 `id_card_number` ≠ 明文 |
| **U-S-08** | 商家认证（无房屋信息） | role_code=merchant | `document_urls` JSON 中无 building/unit/room 字段 |
| **U-S-09** | 网格员认证（无房屋信息） | role_code=grid_worker | `document_urls` JSON 中无 building/unit/room 字段 |
| **U-S-10** | 租户认证（含房屋信息+租约） | role_code=tenant, 含 building/unit/room | `document_urls` JSON 包含房屋信息 |

### 3.4 审核认证 (ReviewCertification)

参考：[user-design.md §3.5](user-design.md) | Proto: `user.v1.UserService.ReviewCertification`

| 编号 | 场景 | 输入 | 预期结果 |
|------|------|------|---------|
| **U-R-01** | 审核通过 — owner 完整流程 | `certification_id, reviewer_id=9001, result=2(通过)` | `code=0`；cert.status 1→2；role.verf_status 1→2, role.verified_at=now, role.expires_at=NULL(永久)；user_base.real_name/ID_card 回填(COALESCE)；**residence 创建**（house_id="3-2-1501"）；**`auth:roles:{uid}` 缓存已删除** |
| **U-R-02** | 审核通过 — tenant（有时限） | `result=2, expires_at="2027-06-03"` | role.expires_at = 指定日期；residence 创建（含 start_date/end_date） |
| **U-R-03** | 审核通过 — grid_worker（默认 1 年） | `result=2, 无 expires_at` | role.expires_at = now + 365天 |
| **U-R-04** | 审核通过 — community_admin（默认 2 年） | `result=2` | role.expires_at = now + 2*365天 |
| **U-R-05** | 审核通过 — merchant（永久） | `result=2` | role.expires_at = NULL；**不创建 residence**（merchant 无 membership_id） |
| **U-R-06** | 审核驳回 | `result=3(驳回), review_notes="材料不清晰"` | `code=0`；cert.status 1→3；role.verf_status 1→3；**`auth:roles:{uid}` 缓存已删除**；不创建 residence；不回填实名信息 |
| **U-R-07** | certification 不存在 | `certification_id=99999` | `code=100007` |
| **U-R-08** | certification 状态不是待审核 | cert 已通过(status=2) | `code=100007` |
| **U-R-09** | 实名信息回填 — 首次（COALESCE） | user_base.real_name 为 NULL | 回填成功 |
| **U-R-10** | 实名信息回填 — 已有不覆盖 | user_base.real_name="李四" | 回填不改变已有值（COALESCE） |
| **U-R-11** | owner 认证通过创建 residence（首次） | house_id="3-2-1501" | INSERT residence，is_primary=1 |
| **U-R-12** | owner 认证通过 — 同 house_id 已存在 | 已有 residence(house_id="3-2-1501") | 不重复创建 |
| **U-R-13** | 审核人信息记录 | `reviewer_id=9001` | cert.reviewer_id=9001, cert.review_time 非空 |
| **U-R-14** | **缓存失效：审核通过后 GetUserRoles 穿透 DB** | 审核通过后调用 GetUserRoles(verf_status=2) | 缓存 MISS → 查 DB → 返回最新角色（含新通过的角色） |

### 3.5 角色查询 (GetUserRoles)

参考：[user-design.md §十一](user-design.md) | Proto: `user.v1.UserService.GetUserRoles`

| 编号 | 场景 | 输入 | 预期结果 |
|------|------|------|---------|
| **U-G-01** | 查询已认证角色（verf_status=2）— 缓存命中 | `user_id, verf_status=2` | 返回已认证角色列表，从 Redis 缓存读取（<1ms） |
| **U-G-02** | 查询已认证角色 — 缓存未命中，穿透 DB | `user_id, verf_status=2`，Redis 中无缓存 | 查 DB → 返回 → 回填 Redis（TTL=300s） |
| **U-G-03** | 查询全部角色（无 verf_status 过滤） | `user_id` | 直接查 DB，不走缓存 |
| **U-G-04** | 用户无任何角色 | new user | `roles=[]` |
| **U-G-05** | 用户多角色（owner+committee） | 同一小区两个角色 | 返回两条记录 |
| **U-G-06** | 按小区过滤 | `community_id=2001` | 仅返回该小区的角色 |
| **U-G-07** | 仅返回 verf_status=2 的角色（auth-service 调用语义） | `verf_status=2` | 不返回 verf_status=0/1/3/4 的角色 |

### 3.6 权限校验 (CheckAccess)

参考：[user-design.md §9.4](user-design.md) | Proto: `user.v1.UserService.CheckAccess`

| 编号 | 场景 | 输入 | 预期结果 |
|------|------|------|---------|
| **U-C-01** | owner 角色校验通过 | `user_id, role_codes=["owner"], community_id=2001` | `allowed=true, matched_role="owner", matched_community_id=2001` |
| **U-C-02** | 角色不匹配 | `user_id, role_codes=["community_admin"], community_id=2001`（用户是 owner） | `allowed=false` |
| **U-C-03** | 小区不匹配 | `user_id, role_codes=["owner"], community_id=9999`（用户在 2001 是 owner） | `allowed=false` |
| **U-C-04** | 未认证角色（verf_status≠2）不通过 | 用户有 owner 角色但 verf_status=0 | `allowed=false` |
| **U-C-05** | merchant 全局角色（community_id=0） | `role_codes=["merchant"], community_id=0` | `allowed=true, matched_role="merchant", matched_community_id=0` |
| **U-C-06** | 多角色任一命中即通过 | user 有 owner+committee，`role_codes=["committee"]` | `allowed=true, matched_role="committee"` |
| **U-C-07** | 角色已过期（verf_status=4） | 用户 role 已过期 | `allowed=false` |

### 3.7 房屋绑定 (BindResidence)

参考：[user-design.md §2.5](user-design.md) | Proto: `user.v1.UserService.BindResidence`

| 编号 | 场景 | 预期结果 |
|------|------|---------|
| **U-B-01** | owner 认证通过后绑定房屋 | house_id="{building}-{unit}-{room}" 格式正确 |
| **U-B-02** | 同一 membership 下同 house_id 唯一 | uk_member_house 约束生效 |
| **U-B-03** | 多套房标记主房产 | is_primary=1 |

---

## 四、Redis 专项测试

### 4.1 角色缓存 Cache-Aside 流程

参考：[auth-design.md §五](auth-design.md)

| 编号 | 场景 | 操作 | 验证 |
|------|------|------|------|
| **R-C-01** | 首次读取 → 缓存 MISS → 穿透 gRPC → 回填 | `getUserRolesWithCache(uid)` | Redis `auth:roles:{uid}` 被 SET，TTL=300；返回正确的 roles |
| **R-C-02** | 缓存命中 → 直接返回 | 连续两次调用 | 第二次不触发 gRPC 调用（Mock 验证调用次数=1）；返回数据一致 |
| **R-C-03** | 缓存 JSON 解析失败 → 穿透兜底 | 手动写入损坏的 JSON 到 `auth:roles:{uid}` | 忽略缓存，gRPC 穿透拉取，回填新正确数据 |
| **R-C-04** | 缓存 TTL 到期自然失效 | 等待 300s 后调用 | 缓存 MISS → gRPC 穿透 → 重新回填 |

### 4.2 缓存失效时机验证

参考：[auth-design.md §5.2](auth-design.md) | [user-design.md §11.3](user-design.md)

| 编号 | 场景 | 触发动作 | 验证 |
|------|------|---------|------|
| **R-I-01** | 审核通过 → 缓存失效 | `ReviewCertification(result=2)` | `auth:roles:{uid}` key 被 DEL；下次读取触发 MISS |
| **R-I-02** | 审核驳回 → 缓存失效 | `ReviewCertification(result=3)` | 同上 |
| **R-I-03** | user-service DEL 失败 → TTL 兜底 | Mock Redis DEL 返回错误 | 缓存最多 300s 后自动过期，下次读取正确穿透 |
| **R-I-04** | AT 刷新时角色更新生效延迟 | 审核通过 → 用户 RefreshToken | 新 AT 包含最新角色（缓存已被 DEL，穿透获取）。最长延迟 ≤ 5min(TTL) + 15min(旧AT) |

### 4.3 RT 持久化与旋转

参考：[auth-design.md §3.4](auth-design.md)

| 编号 | 场景 | 验证 |
|------|------|------|
| **R-T-01** | RT 写入 Redis | `auth:rt:{uid}:{device}` key 存在，TTL=1296000，值=jti |
| **R-T-02** | RT 旋转：旧 RT 删除 + 新 RT 写入 | 刷新后旧 jti 的 key 不存在（或被覆盖为新 jti） |
| **R-T-03** | RT 旋转 Lua 原子性 | 两次并发刷新，仅一次成功；无数据竞争 |
| **R-T-04** | RT TTL 重置 | 旧 RT 剩 1 天，刷新后新 RT TTL=1296000 |

### 4.4 AT 黑名单

参考：[auth-design.md §3.5](auth-design.md)

| 编号 | 场景 | 验证 |
|------|------|------|
| **R-B-01** | 注销时 AT 加入黑名单 | `auth:at:blacklist:{jti}=1`，TTL=AT剩余秒数 |
| **R-B-02** | ValidateToken 检查黑名单 | 黑名单中的 AT 返回 `valid=false` |
| **R-B-03** | 黑名单 TTL 到期自动清理 | TTL 过后 key 不存在 |
| **R-B-04** | 不同设备注销互不影响 | 设备 A 注销 → 设备 A 的 AT 拉黑 + RT 删除；设备 B 的 RT 和 AT 不受影响 |

### 4.5 Redis 不可用降级

| 编号 | 场景 | 验证 |
|------|------|------|
| **R-D-01** | auth-service：Redis 不可用时角色查询 | `getUserRolesWithCache` 中 Redis GET 返回 error → 直接穿透 gRPC，不阻塞 |
| **R-D-02** | user-service：Redis 不可用时 GetUserRoles | `getRolesFromCache` 中 `rds==nil` → 返回 nil（缓存 MISS），走 DB |
| **R-D-03** | user-service：Redis 不可用时缓存失效 | `invalidateRolesCache` 中 `rds==nil` → 直接 return，不报错 |
| **R-D-04** | auth-service：Redis 不可用时登录 | RT 持久化失败 → 应返回错误（当前实现会返回 error）**待确认** |

---

## 五、JWT 与角色数据范围专项

### 5.1 AT Payload 结构验证

| 编号 | 验证点 | 预期 |
|------|--------|------|
| **J-S-01** | AT 必须包含的字段 | `user_id`(int64), `jti`(string), `roles`(array), `exp`(int64), `iat`(int64) |
| **J-S-02** | roles 中每个条目的字段 | `r`(string, role_code), `c`(int64, community_id) |
| **J-S-03** | AT 不含多余敏感字段 | 不含 `phone`, `password`, `real_name`, `id_card_number` |
| **J-S-04** | jti 格式 | `{user_id}-{unix_nano}`，如 `1234567890-1717430400123456789` |
| **J-S-05** | AT 过期时间 = iat + 900s | `exp - iat == 900` |
| **J-S-06** | AT 签名算法 | `HS256` (HMAC-SHA256) |
| **J-S-07** | RT 不含 roles | RT payload 仅含 `user_id, device_id, jti, exp, iat` |

### 5.2 roles 字段正确性

| 编号 | 场景 | 用户角色 | 预期 AT roles |
|------|------|---------|---------------|
| **J-R-01** | 单角色单小区 | owner in community 1001 | `[{"r":"owner","c":1001}]` |
| **J-R-02** | 多角色同小区 | owner + committee in community 1001 | `[{"r":"owner","c":1001},{"r":"committee","c":1001}]` |
| **J-R-03** | 同角色多小区（网格员）| grid_worker in [1001, 1002, 1003] | `[{"r":"grid_worker","c":1001},{"r":"grid_worker","c":1002},{"r":"grid_worker","c":1003}]` |
| **J-R-04** | 商家角色（全局） | merchant | `[{"r":"merchant","c":0}]` |
| **J-R-05** | 无角色用户 | 新注册用户 | `[]` (空数组) |
| **J-R-06** | 未认证角色不进 JWT | owner verf_status=0 | 该角色不在 roles 中出现 |
| **J-R-07** | 已过期角色不进 JWT | grid_worker verf_status=4 | 该角色不在 roles 中出现 |
| **J-R-08** | roles 仅含 verf_status=2 的角色 | 混合状态 | 仅出现 verf_status=2 的条目 |
| **J-R-09** | **roles 上限 50 条**（极端场景） | 网格员跨 100+ 小区 | 超过 50 条时 c 省略，Gateway 降级 CheckAccess（当前代码未实现此限制——**待确认**） |

### 5.3 各角色 community_id 数据范围

| 编号 | 场景 | 验证 |
|------|------|------|
| **J-C-01** | owner 只能访问本小区数据 | CheckAccess(community_id=9999) → false |
| **J-C-02** | tenant 只能访问本小区数据 | 同 owner |
| **J-C-03** | grid_worker 可访问管辖小区 | CheckAccess(community_id=1001) → true；CheckAccess(community_id=9999) → false |
| **J-C-04** | community_admin 本小区范围 | 仅所在小区 |
| **J-C-05** | property_admin 本小区范围 | 仅所在小区 |
| **J-C-06** | merchant 全局范围 | c=0，CheckAccess(community_id=0) → true |
| **J-C-07** | 用户退出小区后角色失效 | membership.bind_status=0，CheckAccess → false（JOIN membership ON bind_status=1） |

### 5.4 RT 刷新后角色更新验证

| 编号 | 场景 | 验证 |
|------|------|------|
| **J-T-01** | 认证通过前 AT roles=[] | 注册→登录→解析 AT，roles=[] |
| **J-T-02** | 认证通过后刷新 AT → roles 更新 | 审核通过 → RefreshToken → 新 AT 包含新角色 |
| **J-T-03** | 角色被撤销后刷新 AT → roles 移除 | 管理员撤销角色 → RefreshToken → 新 AT 不包含该角色 |
| **J-T-04** | 不刷新 Token → 旧 AT 仍含旧角色（最多 15 分钟） | 审核通过后不刷新，旧 AT 在 15 分钟内仍不含新角色（预期行为） |

---

## 六、端到端串联用例（关键路径冒烟）

### 6.1 核心流程：普通用户注册 → 业主认证通过

```
步骤 1: Register(phone=13800138000, sms_code, password, nickname="张三")
  → ✓ code=0, 获得 user_id, AT(roles=[]), RT

步骤 2: JoinCommunity(user_id, community_id=2001)
  → ✓ code=0, membership 创建, bind_status=1

步骤 3: ApplyRole(user_id, community_id=2001, role_code=owner, building=3, unit=2, room=1501)
  → ✓ code=0, role 创建, verf_status=0

步骤 4: SubmitCertification(user_id, role_id, document_urls=["房产证.jpg"],
         real_name="张三", id_card_number="110101...", building=3, unit=2, room=1501)
  → ✓ code=0, certification 创建, role.verf_status=0→1

步骤 5: ReviewCertification(certification_id, reviewer_id=9001, result=2)
  → ✓ code=0, cert.status=2, role.verf_status=2(已通过), residence 创建(house_id="3-2-1501")
  → ✓ Redis auth:roles:{user_id} 已删除
  → ✓ user_base.real_name 回填为 "张三"

步骤 6: RefreshToken(refresh_token)
  → ✓ 新 AT 的 roles=[{"r":"owner","c":2001}]
  → ✓ Redis 角色缓存 MISS → gRPC 穿透 → 回填

步骤 7: Login(encrypted_phone, encrypted_password)
  → ✓ AT 包含 roles=[{"r":"owner","c":2001}]

步骤 8: CheckAccess(user_id, role_codes=["owner"], community_id=2001)
  → ✓ allowed=true, matched_role="owner", matched_community_id=2001

步骤 9: CheckAccess(user_id, role_codes=["owner"], community_id=9999)
  → ✓ allowed=false (范围校验)

步骤 10: GetResidences(membership_id)
  → ✓ 返回 residence，house_id="3-2-1501", building="3", unit="2", room="1501"
```

### 6.2 多角色场景

```
用户已有 owner 角色（已认证），再申请 committee 角色：

步骤 1: ApplyRole(role_code=committee)
  → ✓ code=0, verf_status=0

步骤 2: SubmitCertification(role_id, 业委会证明材料)
  → ✓ code=0

步骤 3: ReviewCertification(result=2)
  → ✓ role.verf_status=2

步骤 4: RefreshToken
  → ✓ AT roles=[{"r":"owner","c":2001},{"r":"committee","c":2001}]
```

### 6.3 驳回后重新提交

```
步骤 1: SubmitCertification → ReviewCertification(result=3, 驳回)
  → ✓ cert.status=3, role.verf_status=3
  → ✓ auth:roles:{uid} 已删除

步骤 2: SubmitCertification(重新提交)
  → ✓ code=0（verf_status=3 允许提交），新 certification 创建, role.verf_status=3→1

步骤 3: ReviewCertification(result=2, 通过)
  → ✓ role.verf_status=1→2
  → ✓ 首次驳回的 certification 记录仍在（status=3），新记录 status=2
```

### 6.4 商家角色全流程

```
步骤 1: 加入小区（不需要，商家不绑小区）
  → 跳过

步骤 2: ApplyRole(role_code=merchant)
  → ✓ membership_id=NULL, community_id=0

步骤 3: SubmitCertification(无 building/unit/room)
  → ✓ document_urls 不含房屋信息

步骤 4: ReviewCertification(result=2)
  → ✓ role.expires_at=NULL（永久）, 不创建 residence

步骤 5: RefreshToken
  → ✓ AT roles=[{"r":"merchant","c":0}]

步骤 6: CheckAccess(role_codes=["merchant"], community_id=0)
  → ✓ allowed=true
```

### 6.5 网格员跨多小区

```
步骤 1: JoinCommunity(2001) + ApplyRole(grid_worker) + SubmitCertification + Review(通过)
步骤 2: JoinCommunity(2002) + ApplyRole(grid_worker) + SubmitCertification + Review(通过)
步骤 3: JoinCommunity(2003) + ApplyRole(grid_worker) + SubmitCertification + Review(通过)

步骤 4: RefreshToken
  → ✓ AT roles=[
       {"r":"grid_worker","c":2001},
       {"r":"grid_worker","c":2002},
       {"r":"grid_worker","c":2003}
     ]

步骤 5: CheckAccess(role_codes=["grid_worker"], community_id=2001) → true
步骤 6: CheckAccess(role_codes=["grid_worker"], community_id=2002) → true
步骤 7: CheckAccess(role_codes=["grid_worker"], community_id=9999) → false
```

---

## 七、边界条件与异常场景

| 编号 | 场景 | 验证 |
|------|------|------|
| **E-01** | 注册时 user-service gRPC 超时 | auth-service 返回友好错误，不写 credential |
| **E-02** | 登录时 user-service gRPC 超时 | 返回 509503 |
| **E-03** | 审核认证时 user_id 不匹配 | certification.user_id ≠ role.user_id，安全校验 |
| **E-04** | 并发注册同一手机号 | 第二次 CreateUser 返回 100002 |
| **E-05** | 并发加入同一小区 | uk_user_community 唯一约束，第二次幂等返回 |
| **E-06** | 退出小区后 role 保留 | LeaveCommunity → membership.bind_status=0, role 记录保留（verf_status 不变） |
| **E-07** | 加入第 6 个小区 | `code=100006` |
| **E-08** | 审核人不存在 | reviewer_id=99999，当前实现不校验 reviewer 存在性——**待确认** |

---

## 八、测试执行计划

### 8.1 执行顺序

```
Phase 1: auth-service 单元测试（Mock user-service gRPC + miniredis）
  ├── 2.1 注册 (Register) — 12 个用例
  ├── 2.2 账密登录 (Login) — 14 个用例
  ├── 2.3 短信验证码登录 (LoginSms) — 6 个用例
  ├── 2.4 Token 刷新 (RefreshToken) — 10 个用例
  ├── 2.5 注销 (Logout) — 6 个用例
  └── 2.6 Token 验证 (ValidateToken) — 6 个用例

Phase 2: user-service 单元测试（Mock DB + miniredis）
  ├── 3.1 加入小区 (JoinCommunity) — 7 个用例
  ├── 3.2 申请角色 (ApplyRole) — 11 个用例
  ├── 3.3 提交认证 (SubmitCertification) — 10 个用例
  ├── 3.4 审核认证 (ReviewCertification) — 14 个用例
  ├── 3.5 角色查询 (GetUserRoles) — 7 个用例
  ├── 3.6 权限校验 (CheckAccess) — 7 个用例
  └── 3.7 房屋绑定 (BindResidence) — 3 个用例

Phase 3: Redis 专项测试
  ├── 4.1 角色缓存 Cache-Aside — 4 个用例
  ├── 4.2 缓存失效时机 — 4 个用例
  ├── 4.3 RT 持久化与旋转 — 4 个用例
  ├── 4.4 AT 黑名单 — 4 个用例
  └── 4.5 Redis 不可用降级 — 4 个用例

Phase 4: JWT 与角色数据范围专项
  ├── 5.1 AT Payload 结构 — 7 个用例
  ├── 5.2 roles 字段正确性 — 9 个用例
  ├── 5.3 各角色 community_id 范围 — 7 个用例
  └── 5.4 RT 刷新后角色更新 — 4 个用例

Phase 5: 端到端串联（miniredis + Mock DB）
  ├── 6.1 注册→业主认证通过 — 10 步验证
  ├── 6.2 多角色场景
  ├── 6.3 驳回后重新提交
  ├── 6.4 商家全流程
  └── 6.5 网格员跨多小区
```

### 8.2 用例统计

| 分类 | 用例数 |
|------|:---:|
| 二、auth-service | 54 |
| 三、user-service | 59 |
| 四、Redis 专项 | 20 |
| 五、JWT 与角色数据范围 | 27 |
| 六、端到端串联 | 5 个场景 |
| 七、边界条件与异常 | 8 |
| **总计** | **168 个用例 + 5 个 E2E 场景** |

---

## 九、关注点检查清单

### 9.1 Redis 是否按设计使用

- [ ] `sms:code:{phone}` — 注册/登录时校验，校验后立即 DEL（防重放）
- [ ] `auth:rt:{uid}:{device}` — 登录/注册时 SET（TTL=15天），刷新时 Lua 原子旋转（DEL旧+SET新），注销时 DEL
- [ ] `auth:rt:{uid}:devices` — 登录时 SADD，强踢时 SMEMBERS→批量 DEL
- [ ] `auth:at:blacklist:{jti}` — 注销时 SET（TTL=AT剩余时间），ValidateToken 时检查
- [ ] `auth:roles:{uid}` — auth-service 签发 JWT 时 GET（Cache-Aside），user-service 审核通过/驳回时 DEL

### 9.2 Redis 是否按设计失效

- [ ] `sms:code` 验证后立即删除（不依赖 TTL）
- [ ] `auth:rt:{uid}:{device}` Lua 原子替换，旧 RT 不可复用
- [ ] `auth:at:blacklist:{jti}` TTL 精确对齐 AT 剩余有效期，到期自动清理
- [ ] `auth:roles:{uid}` user-service 主动 DEL + TTL 300s 兜底
- [ ] DEL 操作失败时 TTL 自然过期兜底，不阻塞主流程

### 9.3 JWT 是否封装了角色和小区 ID

- [ ] AT 包含 `roles` 数组，每项含 `r`(role_code) 和 `c`(community_id)
- [ ] `merchant` 角色 `c=0`（全局）
- [ ] `roles` 仅包含 `verf_status=2`（已认证通过）的角色
- [ ] RT 不含 `roles`（仅用于刷新）
- [ ] AT 不含敏感信息（phone、password、real_name、id_card_number）
- [ ] jti 格式：`{user_id}-{unix_nano}`

### 9.4 各角色数据范围是否正确

- [ ] owner/tenant：限制本小区（community_id = 角色所属小区）
- [ ] grid_worker：限制管辖小区（可跨多小区）
- [ ] community_admin/property_admin/committee：限制本小区
- [ ] merchant：全局（community_id=0），不绑定小区
- [ ] CheckAccess 校验 verf_status=2（已认证）+ community_id 匹配
- [ ] 退出小区后 CheckAccess 返回 false（membership.bind_status=0）

---

## 十、待确认事项

以下是代码审查中发现的与设计文档不一致或需要确认的点：

| 编号 | 事项 | 设计文档描述 | 代码现状 | 建议 |
|------|------|------------|---------|------|
| **Q-01** | Login 未校验账号禁用状态 | auth-design.md §3.1 步骤5：校验 status != 2 | loginlogic.go 调用了 GetUserByPhone 获取用户信息，但未校验 `user.status` | 增加 status 校验逻辑 |
| **Q-02** | ValidateToken 当前实现细节未知 | auth-design.md §3.6：检查黑名单 | 需确认 validatetokenlogic.go 是否包含黑名单检查 | 阅读代码确认 |
| **Q-03** | RefreshToken 旋转成功但 GetUserRoles 失败时 RT 已丢失 | auth-design.md 未明确 | 当前 L62-120：Lua 旋转先执行，GetUserRoles 后执行。GetUserRoles 失败时旧 RT 已被删除、新 RT 未返回给客户端 | 考虑先拉角色再旋转，或增加补偿机制 |
| **Q-04** | roles 上限 50 条未实现 | auth-design.md §2.3：超 50 条时省略 c | 当前 rolecache.go 无条目数限制 | 增加上限检查 |
| **Q-05** | JoinCommunity 未校验 user 存在性 | user-design.md 无明确说明 | join_community_logic.go 直接 INSERT，不校验 user_base 是否有记录 | 可忽略（由调用方保证） |
| **Q-06** | ApplyRole 同一小区不同 role_code 的行为 | user-design.md：uk_member_role 阻止同 role_code 但允许不同 | 当前实现仅查 FindByMembershipAndRole(membership_id, role_code)，应允许 | 确认无误 |
| **Q-07** | ReviewCertification 未校验 reviewer_id 存在性 | 设计文档未明确 | review_certification_logic.go 直接使用 reviewer_id，不校验 | 可接受（由上层保证） |

---

## 附录 A：错误码速查

| 错误码 | 含义 | 涉及服务 |
|--------|------|---------|
| 0 | 成功 | 全部 |
| 50001 | 登录失败（手机号或密码错误） | auth |
| 50002 | Token 已过期 | auth |
| 50003 | RT 已失效（已注销或已被旋转淘汰） | auth |
| 50004 | 验证码错误或已过期 | auth |
| 50005 | 账号已被禁用 | auth |
| 50006 | Token 已被拉黑（已注销） | auth |
| 509503 | 获取用户信息失败（user-service 不可用） | auth |
| 509504 | 获取用户角色失败 | auth |
| 500400 | 参数校验失败 | auth |
| 100001 | 用户不存在 | user |
| 100002 | 手机号已注册 | user |
| 100003 | 该角色已提交认证申请，请勿重复提交 | user |
| 100004 | 信用分不足 | user |
| 100005 | 小区成员关系不存在或已退出 | user |
| 100006 | 最多加入 5 个小区 | user |
| 100007 | 认证申请不存在或状态不允许操作 | user |
| 100008 | 该角色已存在 | user |
| 100009 | 角色认证已过期，请重新认证 | user |
| 100010 | 权限不足（CheckAccess 拒绝） | user |
| 100400 | 参数校验失败 | user |

## 附录 B：测试工具与依赖

```go
// 推荐使用的 Go 测试库
github.com/alicebob/miniredis/v2  // 内存 Redis，真实 Redis 语义
github.com/stretchr/testify       // 断言库（assert/require/mock）
github.com/golang/mock            // Mock 生成（user-service gRPC client）
```

## 附录 C：服务启动顺序

```
docker compose up -d  # MySQL, etcd, Redis, APISIX, MinIO
→ user-service (8082)
→ auth-service (8083)
```
