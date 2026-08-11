# 「我的小区我的家」— 用户中心设计方案

## 一、定位

`user-service` 负责平台所有用户的**身份管理**、**小区归属**、**角色认证**。

### 1.1 做什么

| 职责 | 说明 |
|------|------|
| 用户注册与基本信息维护 | 手机号注册、昵称、头像、实名信息（认证后回填）、偏好设置 |
| 加入/退出小区 | 用户选择一个小区，建立成员关系（不表达身份，只表达"在这个小区"） |
| 角色申请与认证 | 用户申请成为业主/租户/网格员/社区管理员/物业管理员/业委会/商家，上传材料，平台审核，通过后获得角色 |
| 房屋绑定 | 居民（业主/租户）填写楼号、单元号、房号，建立房屋关联（仅认证通过后可操作） |
| 角色有效期管理 | 租户、网格员等有时限角色，到期自动标记过期 |
| 认证轨迹记录 | 所有认证申请的历史记录（申请→待审→通过/驳回→过期→重新认证） |
| 权限校验 | 提供 `CheckAccess` RPC，供 API Gateway 实时鉴权（用户+角色+小区范围） |

### 1.2 不做什么

| 不负责 | 归属 |
|--------|------|
| 登录/鉴权/Token 签发 | auth-service |
| 短信验证码 | auth-service |
| 设备指纹、换机验证 | auth-service |
| 登录日志（时间/IP/设备） | auth-service `auth_login_log` |
| 小区信息管理（小区名称、省市区、楼栋字典） | community-service |
| 房屋字典维护 | **不维护**，用户自助输入楼号/单元号/房号 |
| 角色-API 权限映射（网格员能调哪些接口） | API Gateway 配置文件 |
| 商家广告投放范围（小区/县区/地市） | ad-service `ad_merchant_scope` |
| 家庭管理（同住成员、生日共享） | home-service（未来） |
| 族谱 | home-service（未来） |
| 文件上传/存储 | file-service（本服务只存 MinIO URL 引用） |

### 1.3 核心设计决策

- **Membership 与 Role 分离**：加入小区 = 一条 membership 记录（一辈子一条）。角色 = 挂在 membership 下的多条 role 记录（每人每角色一条，独立认证）
- **房屋与身份分离**：房屋信息独立在 `user_residence` 表，仅 role_code=owner/tenant 时挂载
- **所有角色统一认证流程**：业主、租户、网格员、社区管理员、物业管理员、业委会、商家，全部走"申请→提交材料→审核→通过/驳回"同一流程
- **商家也是认证角色**：`role_code='merchant'`，`membership_id=NULL`，不绑定具体小区
- **角色有有效期**：租户（租约到期日）、网格员（1年）、管理员（任期）、业委会（任期）；业主/商家（永久）
- **不维护房屋字典**：用户输入楼号/单元号/房号，系统拼接为 `house_id`（如 `1-2-301`）
- **偏好用 JSON**：`preferences JSON` 列，可扩展
- **认证通过后才创建房屋记录**：ApplyRole 只创建角色（verf_status=0），SubmitCertification 暂存房屋信息，ReviewCertification(approved) 才创建 residence。房产证/租赁合同是房屋归属的证明，未认证不产生 residence 记录
- **实名信息回填**：认证审核通过时，real_name 和 id_card_number（AES 加密）通过 COALESCE 回填到 user_base（首次回填，已有不覆盖）
- **全表雪花 ID**：5 张表统一使用 snowflake 算法生成主键，分库分表无 ID 冲突
- **user_residence 冗余 user_id**：作为分片键，按 user_id 分片时无需 JOIN
- **按 user_id 分片**：所有表均含 user_id，同一用户所有数据落在一个分片

---

## 二、数据库设计（`user` 库，5 张表）

### 2.1 `user_base` — 用户基础表

```sql
CREATE TABLE user_base (
    id                  BIGINT NOT NULL COMMENT '用户ID（雪花算法生成）',
    phone               VARCHAR(255) NOT NULL COMMENT '手机号（AES加密存储）',
    nickname            VARCHAR(100) NULL COMMENT '昵称',
    avatar_url          VARCHAR(500) NULL COMMENT '头像URL',
    real_name           VARCHAR(50) NULL COMMENT '真实姓名（首次认证通过后回填）',
    id_card_number      VARCHAR(255) NULL COMMENT '身份证号（AES加密，首次认证通过后回填）',
    gender              TINYINT NULL COMMENT '性别：1-男 2-女',
    birth_date          DATE NULL COMMENT '出生日期',
    status              TINYINT NOT NULL DEFAULT 1 COMMENT '状态：1-正常 2-禁用',
    credit_score        INT NOT NULL DEFAULT 100 COMMENT '信用分（默认100，最低0）',
    preferences         JSON NULL COMMENT '用户偏好。例: {"default_community_id":123}',
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at         DATETIME NULL COMMENT '软删除时间',
    PRIMARY KEY (id),
    UNIQUE INDEX idx_phone (phone),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户基础表';
```

| 字段 | 说明 |
|------|------|
| `id` | 雪花算法生成，非自增 |
| `phone` | AES-256 加密存储，唯一索引 |
| `real_name` / `id_card_number` | 任何角色首次认证通过后回填，后续不再覆盖 |
| `status` | 1-正常 2-禁用（无 3-已删除，软删除走 deleted_at） |
| `preferences` | JSON 可扩展。当前仅含 `default_community_id`，后续可加语言、通知开关等 |
| `deleted_at` | 软删除。注销账号时设置，数据不物理删除 |

**无 `user_type`**：用户类型不在全局。角色是相对小区的，见 `user_membership_role`。
**无 `cert_status`**：认证状态在 role 级别，不在 user 级别。
**无 `scope_id`**：数据范围由 role 决定。

---

### 2.2 `user_community_membership` — 小区成员关系表

```sql
CREATE TABLE user_community_membership (
    id                  BIGINT NOT NULL COMMENT '雪花算法生成',
    user_id             BIGINT NOT NULL COMMENT '用户ID（FK → user_base.id）',
    community_id        BIGINT NOT NULL COMMENT '小区ID',
    bind_status         TINYINT NOT NULL DEFAULT 1 COMMENT '1-有效 0-已退出',
    join_time           DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
    leave_time          DATETIME NULL COMMENT '退出时间',
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_user_community (user_id, community_id),
    INDEX idx_community (community_id, bind_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户-小区成员关系（仅表达加入/退出）';
```

| 字段 | 说明 |
|------|------|
| `user_id + community_id` | 唯一索引。一个用户在一个小区只有一次"加入" |
| `bind_status` | 1-有效 0-已退出。退出时置 0，不删记录 |
| `join_time` / `leave_time` | 加入和退出的时间点 |

**此表不表达任何身份或角色**。只表达"这个用户是这个小区的一员"。身份由 `user_membership_role` 表达。

---

### 2.3 `user_membership_role` — 角色表

```sql
CREATE TABLE user_membership_role (
    id                  BIGINT NOT NULL COMMENT '雪花算法生成',
    user_id             BIGINT NOT NULL COMMENT '冗余：用户ID（FK → user_base.id，分片键）',
    membership_id       BIGINT NULL COMMENT '小区成员关系ID（FK → user_community_membership.id），商家为 NULL',
    community_id        BIGINT NOT NULL DEFAULT 0 COMMENT '冗余：小区ID，0=全局角色(商家)',
    role_code           VARCHAR(30) NOT NULL COMMENT '角色编码',
    verf_status         TINYINT NOT NULL DEFAULT 0 COMMENT '认证状态：0-未认证 1-待审 2-已通过 3-已驳回 4-已过期',
    verified_at         DATETIME NULL COMMENT '认证通过时间',
    expires_at          DATETIME NULL COMMENT '过期时间，NULL=永久有效',
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_member_role (membership_id, role_code),
    UNIQUE INDEX uk_user_community_role (user_id, community_id, role_code),
    INDEX idx_community_role (community_id, role_code, verf_status),
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色表';
```

| 字段 | 说明 |
|------|------|
| `membership_id` | 关联小区成员关系。商家为 NULL |
| `community_id` | 冗余字段，方便查询。`membership_id=NULL` 时为 0 |
| `role_code` | 角色编码（封闭枚举，见下表） |
| `verf_status` | 0-未认证 1-待审 2-已通过 3-已驳回 4-已过期 |
| `verified_at` | 认证通过时间 |
| `expires_at` | NULL=永久有效（业主、商家）。非 NULL=到期自动变更为 verf_status=4 |

**唯一约束**：
- `uk_member_role`：同一 membership 下同角色唯一
- `uk_user_community_role`：同一用户在同一社区（或全局）同一角色唯一

**角色编码**：

| role_code | 含义 | membership_id | community_id | expires_at | 备注 |
|-----------|------|:---:|:---:|:---:|------|
| `owner` | 业主 | 有值 | 小区ID | NULL（永久） | 认证通过后获得 |
| `tenant` | 租户 | 有值 | 小区ID | 租约到期日 | 认证可选 |
| `grid_worker` | 网格员 | 有值 | 小区ID | 1年 | 可多个小区（多条 role） |
| `community_admin` | 社区管理员 | 有值 | 小区ID | 任期结束日 | 同小区仅一个在职 |
| `property_admin` | 物业管理员 | 有值 | 小区ID | 合同到期日 | |
| `committee` | 业委会成员 | 有值 | 小区ID | 任期结束日 | |
| `merchant` | 商家 | NULL | 0 | NULL（永久） | 全局角色 |

---

### 2.4 `user_certification` — 认证记录表

```sql
CREATE TABLE user_certification (
    id                  BIGINT NOT NULL COMMENT '雪花算法生成',
    role_id             BIGINT NOT NULL COMMENT '角色ID（FK → user_membership_role.id）',
    user_id             BIGINT NOT NULL COMMENT '冗余：用户ID（FK → user_base.id）',
    document_urls       TEXT NULL COMMENT '证明材料URL列表（JSON数组，文件存 MinIO/file-service）',
    status              TINYINT NOT NULL DEFAULT 1 COMMENT '审核状态：1-待审核 2-已通过 3-已驳回',
    reviewer_id         BIGINT NULL COMMENT '审核人ID',
    review_time         DATETIME NULL COMMENT '审核时间',
    review_notes        VARCHAR(500) NULL COMMENT '审核备注',
    submit_time         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '提交时间',
    PRIMARY KEY (id),
    INDEX idx_role (role_id),
    INDEX idx_user (user_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='认证记录表（所有角色统一流程）';
```

| 字段 | 说明 |
|------|------|
| `role_id` | 关联到 `user_membership_role.id`，一个 role 可多次提交（驳回→重新提交→新记录） |
| `document_urls` | 认证元数据（JSON）。格式：`{"urls":[...], "real_name":"...", "id_card_number":"...（AES加密）", "building":"...", "unit":"...", "room":"..."}`。文件本身存在 file-service（MinIO） |
| `status` | 1-待审 2-已通过 3-已驳回。与 role.verf_status 联动更新 |

**所有 7 种角色（含商家）走同一张表、同一个流程**。

---

### 2.5 `user_residence` — 房屋明细表

```sql
CREATE TABLE user_residence (
    id                  BIGINT NOT NULL COMMENT '雪花算法生成',
    membership_id       BIGINT NOT NULL COMMENT '小区成员关系ID（FK → user_community_membership.id）',
    user_id             BIGINT NOT NULL COMMENT '冗余：用户ID（分片键，FK → user_base.id）',
    house_id            VARCHAR(50) NOT NULL COMMENT '房屋ID，系统拼接，如 1-2-301',
    building            VARCHAR(20) NOT NULL COMMENT '楼号（用户输入）',
    unit                VARCHAR(20) NOT NULL DEFAULT '' COMMENT '单元号（用户输入）',
    room                VARCHAR(20) NOT NULL COMMENT '房号（用户输入）',
    is_primary          TINYINT NOT NULL DEFAULT 0 COMMENT '多套房时标记主房产：1-是 0-否',
    start_date          DATE NULL COMMENT '入住/合同开始日期',
    end_date            DATE NULL COMMENT '搬离/合同结束日期',
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE INDEX uk_member_house (membership_id, house_id),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='居民房屋明细表';
```

| 字段 | 说明 |
|------|------|
| `membership_id` | 关联小区成员关系（而非 role） |
| `user_id` | 冗余字段，方便按用户分片，避免 JOIN membership 表 |
| `house_id` | 系统拼接：`{building}-{unit}-{room}`。unit 为空时：`{building}-{room}` |
| `is_primary` | 同一 membership 下可有多套房，标记哪套是主房产 |
| `start_date` / `end_date` | 租户记录合同起止；业主记录入住时间 |

**仅当用户拥有 `owner` 或 `tenant` 角色且认证通过后，才在此表有记录**。网格员、社区管理员等角色无房屋记录。

---

### 2.6 ER 总览

```
user_base (1)
  │
  ├──< user_community_membership (N)       "用户在小区"
  │     uk: (user_id, community_id)
  │     bind_status: 1-有效 / 0-已退出
  │     │
  │     ├──< user_membership_role (N)      "用户是什么身份"
  │     │     uk: (membership_id, role_code)
  │     │     verf_status: 0→1→2/3→4
  │     │     expires_at: NULL or 具体日期
  │     │     │
  │     │     └──< user_certification (N)  "认证怎么过的"
  │     │           role_id FK
  │     │           status: 1→2/3
  │     │
  │     └──< user_residence (N)            "用户住哪（认证通过后创建）"
  │           uk: (membership_id, house_id)
  │           user_id 冗余（分片键 + 直接查询）
  │           (仅 owner/tenant，认证通过后)
  │
  └── user_membership_role (membership_id=NULL, community_id=0)
        role_code='merchant'                "商家（不绑小区）"

注：所有表的 id 均为雪花算法生成，非 AUTO_INCREMENT。
```

---

## 三、业务流程

### 3.1 用户注册

```
客户端 → auth-service（发送短信验证码）
    → auth-service 调用 user-service.GetUserByPhone(phone)
        → 不存在，调用 CreateUser(phone, nickname)
            → INSERT user_base
                id = 雪花ID
                phone = AES(phone)
                status = 1
                credit_score = 100
    → auth-service 生成 JWT，返回 Token

注册完成: user_base 有 1 条记录
         其他 4 张表均无记录（用户尚未选择小区）
```

### 3.2 加入小区

```
用户 → JoinCommunity(user_id, community_id)

1. 校验：SELECT COUNT(*) FROM membership
   WHERE user_id=? AND bind_status=1
   → ≥ 5 个？返回错误（最多加入 5 个小区）

2. INSERT membership (user_id, community_id, bind_status=1)
   → uk_user_community 保证不重复

3. IF 首次加入小区（preferences 中无 default_community_id）:
   UPDATE user_base SET preferences = 
     JSON_SET(COALESCE(preferences,'{}'), '$.default_community_id', community_id)

加入完成: membership ×1, role ×0, residence ×0
         （用户还没选身份）
```

### 3.3 申请角色（以业主为例）

```
用户 → ApplyRole(user_id, community_id, role_code='owner', building, unit, room)

1. 查 membership:
   SELECT * FROM membership WHERE user_id=? AND community_id=? AND bind_status=1

2. 查是否已有该角色:
   SELECT * FROM role WHERE membership_id=? AND role_code='owner'
   → 已存在？返回错误

3. INSERT role (id=雪花ID, user_id, membership_id, community_id, 
                role_code='owner', verf_status=0)
   -- 注意：此时不创建 residence！房屋记录延后到认证通过时创建
   -- building/unit/room 在 SubmitCertification 时传入并暂存

此时角色为"未认证"（verf_status=0），还不能以"认证业主"身份行事。
申请通过后，building/unit/room 仅用于前端回显，不产生数据库记录。
```

### 3.4 提交认证材料

```
用户 → SubmitCertification(user_id, role_id, document_urls, real_name, 
                            id_card_number, building, unit, room)

1. 查 role:
   SELECT * FROM role WHERE id=? AND user_id=?
   → verf_status IN (0,3,4) 允许提交
   → verf_status IN (1,2) 返回错误（待审中或已通过）

2. 加密 id_card_number = AES(明文)

3. 构建认证元数据（JSON 存入 document_urls 字段）:
   {
     "urls": ["https://oss.example.com/deed.jpg"],
     "real_name": "张三",
     "id_card_number": "AES加密的身份证号",
     "building": "3",           // 仅 owner/tenant
     "unit": "2",               // 仅 owner/tenant
     "room": "1501"             // 仅 owner/tenant
   }

4. INSERT certification (id=雪花ID, role_id, user_id, document_urls=JSON, status=1)
   UPDATE role SET verf_status=1

返回 certification_id

关键设计：
  - real_name 和 id_card_number 暂存在 JSON 中，审核通过时回填 user_base（COALESCE）
  - building/unit/room 暂存在 JSON 中，审核通过时创建 residence
```

### 3.5 审核认证

```
管理员 → ReviewCertification(certification_id, reviewer_id, result, review_notes, expires_at)

result: 2-通过 / 3-驳回

1. 查 certification:
   SELECT * FROM certification WHERE id=? AND status=1(待审)
   → 不存在或状态不对，返回错误

2. 解析 certification.document_urls JSON → 提取 real_name, id_card_number, building, unit, room

3. 更新:
   a. UPDATE certification
      SET status=result, reviewer_id, review_notes, review_time=now()
   b. UPDATE role SET verf_status=result
   c. IF result=2(通过):
        SET role.verified_at = now()
        SET role.expires_at = 根据角色类型计算:
          owner/merchant → NULL（永久）
          tenant   → 审核人指定的租约到期日，或默认 +1年
          grid_worker → now() + 1年
          community_admin → 审核人指定的任期结束日，或默认 +2年
          property_admin → 审核人指定的合同到期日，或默认 +1年
          committee → 审核人指定的任期结束日，或默认 +2年
        -- 回填实名信息（首次回填，已有不覆盖）
        UPDATE user_base
        SET real_name=COALESCE(real_name, ?),
            id_card_number=COALESCE(id_card_number, ?)
        WHERE id=cert.user_id
        -- owner/tenant：创建 residence（房屋归属由房产证/合同证明）
        IF role_code IN ('owner','tenant'):
          INSERT residence (id=雪花ID, membership_id, user_id, 
                            house_id, building, unit, room, is_primary=1)

   d. IF result=3(驳回):
        UPDATE role SET verf_status=3
        用户可重新提交 → INSERT 新 certification + UPDATE role verf_status=1

关键设计：
  - residence 在认证通过后才创建，因为房产证/租赁合同才是房屋归属的证明
  - real_name/id_card_number 通过 COALESCE 回填，首次写入后不再覆盖
  - owner/tenant/merchant 永久有效；其他角色有时限，到期自动标记 verf_status=4
```

### 3.6 角色过期

```
定时任务（每天凌晨执行）:

UPDATE user_membership_role
SET verf_status = 4
WHERE verf_status = 2
  AND expires_at IS NOT NULL
  AND expires_at < NOW();

用户端:
  verf_status=4 → 提示"认证已过期，请重新认证"
  重新提交 → INSERT certification(新记录)
           → UPDATE role SET verf_status=1
           → 审核通过 → verf_status=2 + 新的 expires_at
```

### 3.7 退出小区

```
用户 → LeaveCommunity(user_id, community_id)

1. 查 membership:
   SELECT * FROM membership WHERE user_id=? AND community_id=? AND bind_status=1

2. 开启事务:
   a. UPDATE membership SET bind_status=0, leave_time=now()
   b. 该 membership 下所有 role 的 verf_status 不动（保留历史）
      -- 角色实际上是"自然失效"，因为 membership.bind_status=0
      -- 查询时通过 JOIN membership ON bind_status=1 过滤
   c. IF preferences.default_community_id = 此小区:
        查找用户其他有效的 membership.community_id 更新 preferences
        若无，置 NULL

注意：退出小区 ≠ 删除 role 记录。role 保留历史，便于审计。
```

### 3.8 角色生命周期总览

```
申请角色
  │
  ▼
verf_status=0 (未认证)
  │
  │ 提交材料
  ▼
verf_status=1 (待审核)
  │
  ├── 审核通过 → verf_status=2 (已通过) + verified_at + expires_at
  │    │
  │    ├── 被撤销 → 管理员直接 UPDATE verf_status=3
  │    │
  │    └── 到期 → verf_status=4 (已过期) → 重新提交 → 回到 1
  │
  └── 审核驳回 → verf_status=3 (已驳回)
       │
       └── 重新提交材料 → INSERT 新 certification(新记录)
                          UPDATE role SET verf_status=1
```

---

## 四、Go 数据模型

### 4.1 UserBase

```go
type UserBase struct {
    Id            int64          `db:"id"`
    Phone         string         `db:"phone"`
    Nickname      sql.NullString `db:"nickname"`
    AvatarUrl     sql.NullString `db:"avatar_url"`
    RealName      sql.NullString `db:"real_name"`
    IdCardNumber  sql.NullString `db:"id_card_number"`
    Gender        sql.NullInt64  `db:"gender"`
    BirthDate     sql.NullTime   `db:"birth_date"`
    Status        int64          `db:"status"`
    CreditScore   int64          `db:"credit_score"`
    Preferences   sql.NullString `db:"preferences"`
    CreatedTime   time.Time      `db:"created_at"`
    UpdatedTime   time.Time      `db:"updated_at"`
    DeleteTime    sql.NullTime   `db:"deleted_at"`
}
```

### 4.2 UserCommunityMembership

```go
type UserCommunityMembership struct {
    Id          int64        `db:"id"`
    UserId      int64        `db:"user_id"`
    CommunityId int64        `db:"community_id"`
    BindStatus  int64        `db:"bind_status"`
    JoinTime    time.Time    `db:"join_time"`
    LeaveTime   sql.NullTime `db:"leave_time"`
    CreatedTime time.Time    `db:"created_at"`
    UpdatedTime time.Time    `db:"updated_at"`
}
```

### 4.3 UserMembershipRole

```go
type UserMembershipRole struct {
    Id           int64         `db:"id"`
    UserId       int64         `db:"user_id"`
    MembershipId sql.NullInt64 `db:"membership_id"`  // merchant 为 NULL
    CommunityId  int64         `db:"community_id"`   // merchant 为 0
    RoleCode     string        `db:"role_code"`
    VerfStatus   int64         `db:"verf_status"`
    VerifiedAt   sql.NullTime  `db:"verified_at"`
    ExpiresAt    sql.NullTime  `db:"expires_at"`
    CreatedTime  time.Time     `db:"created_at"`
    UpdatedTime  time.Time     `db:"updated_at"`
}
```

### 4.4 UserCertification

```go
type UserCertification struct {
    Id           int64          `db:"id"`
    RoleId       int64          `db:"role_id"`
    UserId       int64          `db:"user_id"`
    DocumentUrls sql.NullString `db:"document_urls"`
    Status       int64          `db:"status"`
    ReviewerId   sql.NullInt64  `db:"reviewer_id"`
    ReviewTime   sql.NullTime   `db:"review_time"`
    ReviewNotes  sql.NullString `db:"review_notes"`
    SubmitTime   time.Time      `db:"submit_time"`
}
```

### 4.5 UserResidence

```go
type UserResidence struct {
    Id           int64         `db:"id"`
    MembershipId int64         `db:"membership_id"`
    UserId       int64         `db:"user_id"`  // 冗余：分片键
    HouseId      string        `db:"house_id"`
    Building     string        `db:"building"`
    Unit         string        `db:"unit"`
    Room         string        `db:"room"`
    IsPrimary    int64         `db:"is_primary"`
    StartDate    sql.NullTime  `db:"start_date"`
    EndDate      sql.NullTime  `db:"end_date"`
    CreatedTime  time.Time     `db:"created_at"`
    UpdatedTime  time.Time     `db:"updated_at"`
}
```

---

## 五、核心查询示例

```sql
-- 查用户在某小区的完整身份
SELECT 
    m.id AS membership_id,
    m.bind_status,
    GROUP_CONCAT(r.role_code) AS roles,
    GROUP_CONCAT(CASE WHEN r.verf_status=2 THEN r.role_code END) AS verified_roles,
    res.house_id
FROM user_community_membership m
LEFT JOIN user_membership_role r ON r.membership_id = m.id
LEFT JOIN user_residence res ON res.membership_id = m.id
WHERE m.user_id = ? AND m.community_id = ? AND m.bind_status = 1
GROUP BY m.id, res.house_id;

-- 查某小区的所有成员
SELECT u.nickname, u.avatar_url, 
       m.id AS membership_id,
       r.role_code, r.verf_status,
       res.house_id, res.is_primary
FROM user_community_membership m
JOIN user_base u ON u.id = m.user_id
LEFT JOIN user_membership_role r ON r.membership_id = m.id
LEFT JOIN user_residence res ON res.membership_id = m.id
WHERE m.community_id = ? AND m.bind_status = 1;

-- 查即将过期的角色（定时任务）
SELECT id, user_id, role_code, expires_at
FROM user_membership_role
WHERE verf_status = 2 
  AND expires_at IS NOT NULL 
  AND expires_at < DATE_ADD(NOW(), INTERVAL 7 DAY);

-- 到期角色批量标记
UPDATE user_membership_role
SET verf_status = 4
WHERE verf_status = 2 
  AND expires_at IS NOT NULL 
  AND expires_at < NOW();

-- 权限校验（API Gateway 实时调用）
SELECT role_code, community_id FROM user_membership_role
WHERE user_id = ? 
  AND verf_status = 2 
  AND community_id = ?
  AND role_code IN ('owner', 'tenant');

-- 按用户查房屋（利用 user_id 索引，无需 JOIN）
SELECT * FROM user_residence WHERE user_id = ?;

-- 按 membership 查房屋
SELECT * FROM user_residence WHERE membership_id = ?;
```

---

## 六、错误码

| 错误码 | 含义 |
|--------|------|
| 0 | 成功 |
| 10001 | 用户不存在 |
| 10002 | 手机号已注册 |
| 10003 | 该角色已提交认证申请，请勿重复提交 |
| 10004 | 信用分不足 |
| 10005 | 小区成员关系不存在或已退出（含：非认证业主/租户不可绑定房屋） |
| 10006 | 最多加入 5 个小区 |
| 10007 | 认证申请不存在或状态不允许操作 |
| 10008 | 该角色已存在 |
| 10009 | 角色认证已过期，请重新认证 |
| 10010 | 权限不足（CheckAccess 拒绝） |
| 10040 | 参数校验失败 |

---

## 七、与其他服务的交互

```
auth-service
  ├── CreateUser(phone, nickname) → user_id    首次登录自动注册
  ├── GetUserByPhone(phone) → user             登录时查用户（只校验 Base，不读 User）
  ├── UpdateUser(id, status=3)                 Saga 补偿：软删除用户
  └── GetUserRoles(user_id, verf_status=2)     签发 JWT 时获取已认证角色

API Gateway
  ├── CheckAccess(user_id, role_codes, community_id) → allowed
  │     敏感操作实时鉴权（审核认证、发通知等）
  └── GetUserRoles(user_id, community_id) → [roles]
        用于前端展示当前用户的身份

community-service
  ├── GetUsersByIds([id1, id2, ...]) → [users]
  │     小区成员列表批量补全用户信息
  └── GetUserMemberships(user_id) → [memberships]
        查询用户的小区归属

ad-service
  └── ad_merchant_scope                         商家广告范围（本服务不管）

聊天服务（未来）
  ├── 维护 chat_group_member 表（按 community_id 分片）
  ├── JoinCommunity 事件 → 同步添加群成员
  └── GetUsersByIds([...]) → 批量补全头像/昵称

注：permission-service 已合入 user-service，不再独立存在。
```

---

## 八、迁移记录

| 版本 | 迁移 | 内容 |
|------|------|------|
| v2.0 | `migration/001_refactor_to_v2.sql` | 旧 7 表 → 新 5 表，user_base 结构调整 |
| v2.1 | `migration/002_snowflake_ids.sql` | 全表去 AUTO_INCREMENT 改雪花 ID，user_residence 加 user_id |

---

## 九、权限设计

### 9.1 模型

```
角色 (role_code)  +  范围 (community_id)  +  认证状态 (verf_status=2)
      │                        │                        │
      ▼                        ▼                        ▼
  能做什么                  在哪做                   是否已验证
```

### 9.2 各角色权限矩阵

| 操作 | owner | tenant | grid_worker | community_admin | property_admin | committee | merchant |
|------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| 发帖/聊天 | ✅ 本小区 | ✅ 本小区 | - | - | - | - | - |
| 发通知 | - | - | ✅ 管辖小区 | ✅ 管辖小区 | ✅ 管辖小区 | ✅ 管辖小区 | - |
| 管理网格员 | - | - | - | ✅ 本小区 | - | - | - |
| 审核认证 | - | - | - | ✅ | ✅ | - | - |
| 发广告 | - | - | - | - | - | - | ✅ 全局 |
| 绑定房屋 | ✅ | ✅ | - | - | - | - | - |

### 9.3 两种鉴权路径

**路径一：JWT Claims（常规操作）**

```
Auth-service 签发 JWT:
  调用 GetUserRoles(user_id, verf_status=2) → 只取已认证角色
  编码进 JWT: {"roles":[{"r":"owner","c":1001}]}

API Gateway 校验:
  路由配置: POST /api/communities/:cid/posts → allow_roles=[owner,tenant]
  从 JWT 解析 roles → 匹配 role_code + community_id → 放行/403
```

**路径二：CheckAccess 实时校验（敏感操作）**

```
API Gateway 调用:
  CheckAccess(user_id, role_codes=["community_admin","property_admin"], community_id=1001)
  → allowed=true, matched_role="community_admin"
  → allowed=false → 403
```

### 9.4 CheckAccess RPC

```protobuf
rpc CheckAccess(CheckAccessRequest) returns (CheckAccessResponse);

message CheckAccessRequest {
  int64 user_id = 1;
  repeated string role_codes = 2;  // 允许的角色，任一命中即通过
  int64 community_id = 3;          // 范围，0=全局（仅 merchant）
}

message CheckAccessResponse {
  bool allowed = 1;
  string matched_role = 2;         // 命中的角色编码
  int64 matched_community_id = 3;  // 命中的小区 ID
}
```

内部实现：`SELECT role_code FROM user_membership_role WHERE user_id=? AND verf_status=2 AND community_id=? AND role_code IN (?)`

---

## 十、ID 策略与分库分表

### 10.1 全表雪花 ID

5 张表统一使用 snowflake 算法生成主键：

| 表 | 生成位置 | 说明 |
|----|---------|------|
| `user_base` | `create_user_logic.go` | 64 位雪花 ID |
| `user_community_membership` | `join_community_logic.go` | 同 |
| `user_membership_role` | `apply_role_logic.go` | 同 |
| `user_certification` | `submit_certification_logic.go` | 同 |
| `user_residence` | `review_certification_logic.go` + `bind_residence_logic.go` | 同 |

不采用 AUTO_INCREMENT，分库分表时无 ID 冲突。

### 10.2 分片策略

**分片键：`user_id`**

所有表均含 `user_id` 列，同一用户的所有数据落在同一分片：

```
user_base                    ← 按 id (=user_id) 分片
user_community_membership    ← 按 user_id 分片
user_membership_role         ← 按 user_id 分片
user_certification           ← 按 user_id 分片
user_residence               ← 按 user_id 分片（user_id 冗余列）
```

**优势**：
- 90%+ 查询是 user_id 单键查询 → 单分片命中
- 登录、鉴权、个人信息、角色列表 — 全部单分片
- 社区维度的聚合查询由 community-service 负责（按 community_id 分片）

**社区维度查询的处理**：

```
"小区 X 的成员列表"
  → community-service 查 community_members → 得到 [user_id...]
  → user-service.GetUsersByIds([...])       → 并发精确命中各分片
  → 无全分片 scatter
```

---

## 十一、Redis 缓存设计

### 11.1 缓存策略：Cache-Aside

```
GetUserRoles(verf_status=2)
  ┌─ Redis GET auth:roles:{user_id}
  │   ├─ HIT  → 直接返回（<1ms）
  │   └─ MISS → gRPC 查询 DB
  │           → Redis SETEX auth:roles:{user_id} 300 <value>
  │           → 返回
```

### 11.2 缓存范围

| 接口 | 缓存 | 原因 |
|------|:---:|------|
| `GetUserRoles(verf_status=2)` | ✅ | auth-service 签发 JWT 高频调用 |
| `GetUserRoles(无 verf_status)` | ❌ | 低频（管理后台），直接查 DB |
| `GetUser` | ❌ | 低频，个人资料变更频繁 |
| `GetUserByPhone` | ❌ | 低频，仅登录时调用 |

### 11.3 失效策略

```
触发时机                 谁操作              动作
─────────────────────────────────────────────────────────
审核通过/驳回    ReviewCertification    DEL auth:roles:{user_id}
自然过期         Redis                  300s 后自动删除
定时任务标记过期  Cron（未来）           DEL auth:roles:{user_id}
```

代码位置：
- 缓存逻辑：`rpc/internal/logic/user/cache.go`
- 读取：`get_user_roles_logic.go`
- 失效：`review_certification_logic.go`

### 11.4 Redis Key 规范

| Key | 类型 | TTL | 值 |
|-----|------|:---:|---|
| `auth:roles:{user_id}` | String | 300s | JSON 序列化的 `GetUserRolesResponse` |

### 11.5 容错

Redis 不可用时自动降级为纯 DB 查询，不影响核心链路：

```go
if rds == nil {
    return nil  // cache miss → fall through to DB
}
```
