# "我的" 页面 & 加入小区流程 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 改造"我的"页面为V9卡片式，加入小区流程从4步扩展为5步（增加楼号-单元-房号输入），新建退出小区子页面。

**Architecture:** Proto 先行（全局 Claude 修改 `JoinCommunityRequest` + `CommunityMembership`），然后 user-service 后端（DB→Model→RPC→REST）与前端（join-community/leave-community/my）并行。

**Tech Stack:** Go (go-zero sqlx), Protobuf (Buf v2), Uni-app (Vue 3 + TypeScript), MySQL

---

### Task 0: Proto 变更（全局 Claude）

**Files:**
- Modify: `api-proto/api/user/v1/user.proto:265-291`
- Modify: `api-proto/CHANGELOG.md`

> ⚠️ **依赖**: 此 Task 必须在所有其他 Task 之前完成。完成后通知 user-service 子 Claude 和前端子 Claude。

- [ ] **Step 1: 扩展 JoinCommunityRequest — 增加 building/unit/room 字段**

编辑 `api-proto/api/user/v1/user.proto`，找到 `JoinCommunityRequest`（约 265 行），改为：

```protobuf
message JoinCommunityRequest {
  int64 user_id = 1 [jstype = JS_STRING];
  int64 community_id = 2 [jstype = JS_STRING];
  int32 building = 3;   // 楼号, 1-150
  int32 unit = 4;       // 单元号, 1-5
  int32 room = 5;       // 房号, 3位数字
}
```

- [ ] **Step 2: 扩展 CommunityMembership — 增加 building/unit/room 字段**

编辑同一文件，找到 `CommunityMembership`（约 251 行），在 `updated_at = 8` 之后增加：

```protobuf
message CommunityMembership {
  int64 id = 1 [jstype = JS_STRING];
  int64 user_id = 2 [jstype = JS_STRING];
  int64 community_id = 3 [jstype = JS_STRING];
  int32 bind_status = 4;
  int64 join_time = 5;
  int64 leave_time = 6;
  int64 created_at = 7;
  int64 updated_at = 8;
  int32 building = 9;   // 楼号
  int32 unit = 10;      // 单元号
  int32 room = 11;      // 房号
}
```

- [ ] **Step 3: 生成代码 + Lint + Breaking Check**

```bash
cd api-proto && make generate
cd api-proto && make lint
cd api-proto && make breaking-check
```

预期: `make generate` 成功，`make lint` PASS，`make breaking-check` 报告 `JoinCommunityRequest` 和 `CommunityMembership` 新增字段（非破坏性变更，WARNING 级别可以接受）。

- [ ] **Step 4: 更新 CHANGELOG.md**

在 `api-proto/CHANGELOG.md` 顶部追加：

```markdown
## 2026-06-07 — JoinCommunityRequest/CommunityMembership 增加地址字段

### 做了什么
- `JoinCommunityRequest` 增加 `building`(3), `unit`(4), `room`(5) 字段
- `CommunityMembership` 增加 `building`(9), `unit`(10), `room`(11) 字段

### 为什么
"我的"页面加入小区流程需要用户输入楼号-单元-房号

### 影响
- Proto: user/v1
- 调用方: user-service（需同步更新）
- 数据库: user_community_membership 表需加 building/unit/room 列
```

- [ ] **Step 5: 提交**

```bash
git add api-proto/
git commit -m "feat(proto): add building/unit/room fields to JoinCommunityRequest and CommunityMembership

Extend JoinCommunityRequest with building(1-150), unit(1-5), room(3-digit)
fields for the join community flow. CommunityMembership also extended
to return address info.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 1: DB Migration + Model 变更（user-service 子 Claude）

**Files:**
- Create: `services/user-service/migration/003_add_address_fields.sql`
- Modify: `services/user-service/model/user_community_membership.go`（全文件）
- Modify: `services/user-service/model/vars.go`（新增错误常量和唯一性校验方法签名）

- [ ] **Step 1: 创建 Migration 文件**

创建 `services/user-service/migration/003_add_address_fields.sql`：

```sql
-- 003_add_address_fields.sql
-- 为 user_community_membership 表增加楼号/单元号/房号字段

ALTER TABLE user_community_membership
  ADD COLUMN building INT DEFAULT 0 COMMENT '楼号',
  ADD COLUMN unit INT DEFAULT 0 COMMENT '单元号',
  ADD COLUMN room INT DEFAULT 0 COMMENT '房号';

-- 唯一索引：同一小区同一地址不可重复加入（仅对有效绑定）
-- 注意：MySQL 不支持 WHERE 条件的唯一索引，实际唯一性由应用层保证
-- 此处创建普通索引加速查询
CREATE INDEX idx_community_address 
  ON user_community_membership(community_id, building, unit, room);
```

- [ ] **Step 2: 执行 Migration**

```bash
# 确认 MySQL 运行中
docker exec -i mysql mysql -u root -p"$MYSQL_ROOT_PASSWORD" user < services/user-service/migration/003_add_address_fields.sql
```

预期: `Query OK` / 无报错。

- [ ] **Step 3: 扩展 Model struct**

编辑 `services/user-service/model/user_community_membership.go`，在 `UserCommunityMembership` struct 的 `UpdatedTime` 之后增加 3 个字段：

```go
type UserCommunityMembership struct {
	Id          int64        `db:"id"`
	UserId      int64        `db:"user_id"`
	CommunityId int64        `db:"community_id"`
	BindStatus  int64        `db:"bind_status"`
	JoinTime    time.Time    `db:"join_time"`
	LeaveTime   sql.NullTime `db:"leave_time"`
	CreatedTime time.Time    `db:"created_time"`
	UpdatedTime time.Time    `db:"updated_time"`
	Building    int          `db:"building"`  // 🆕 楼号
	Unit        int          `db:"unit"`      // 🆕 单元号
	Room        int          `db:"room"`      // 🆕 房号
}
```

- [ ] **Step 4: 扩展 Model interface — 增加 FindByAddress 方法签名**

在 `UserCommunityMembershipModel` interface 中增加：

```go
type UserCommunityMembershipModel interface {
	Insert(ctx context.Context, data *UserCommunityMembership) (sql.Result, error)
	FindOne(ctx context.Context, id int64) (*UserCommunityMembership, error)
	FindByUserAndCommunity(ctx context.Context, userId, communityId int64) (*UserCommunityMembership, error)
	FindByUserId(ctx context.Context, userId int64) ([]*UserCommunityMembership, error)
	CountActiveByUserId(ctx context.Context, userId int64) (int64, error)
	UpdateBindStatus(ctx context.Context, id int64, bindStatus int64, leaveTime time.Time) error
	// 🆕 按地址查询活跃成员（用于唯一性校验）
	FindByAddress(ctx context.Context, communityId int64, building, unit, room int) (*UserCommunityMembership, error)
}
```

- [ ] **Step 5: 实现 FindByAddress 方法**

在 `defaultUserCommunityMembershipModel` 的方法集中增加：

```go
func (m *defaultUserCommunityMembershipModel) FindByAddress(ctx context.Context, communityId int64, building, unit, room int) (*UserCommunityMembership, error) {
	query := fmt.Sprintf(`SELECT id, user_id, community_id, bind_status, join_time, leave_time, created_time, updated_time, building, unit, room FROM %s WHERE community_id = ? AND building = ? AND unit = ? AND room = ? AND bind_status = ?`, m.table)
	var resp UserCommunityMembership
	err := m.conn.QueryRowCtx(ctx, &resp, query, communityId, building, unit, room, MembershipBindStatusActive)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}
```

- [ ] **Step 6: 更新所有现有 SQL 查询**

所有查询语句需增加 `building, unit, room` 列：

**Insert** (约 46 行):
```go
func (m *defaultUserCommunityMembershipModel) Insert(ctx context.Context, data *UserCommunityMembership) (sql.Result, error) {
	query := fmt.Sprintf(`INSERT INTO %s (id, user_id, community_id, bind_status, join_time, created_time, updated_time, building, unit, room) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, m.table)
	return m.conn.ExecCtx(ctx, query, data.Id, data.UserId, data.CommunityId, data.BindStatus, data.JoinTime, data.CreatedTime, data.UpdatedTime, data.Building, data.Unit, data.Room)
}
```

**FindOne** (约 51 行):
```go
func (m *defaultUserCommunityMembershipModel) FindOne(ctx context.Context, id int64) (*UserCommunityMembership, error) {
	query := fmt.Sprintf(`SELECT id, user_id, community_id, bind_status, join_time, leave_time, created_time, updated_time, building, unit, room FROM %s WHERE id = ?`, m.table)
	// ... rest unchanged
}
```

**FindByUserAndCommunity** (约 63 行):
```go
func (m *defaultUserCommunityMembershipModel) FindByUserAndCommunity(ctx context.Context, userId, communityId int64) (*UserCommunityMembership, error) {
	query := fmt.Sprintf(`SELECT id, user_id, community_id, bind_status, join_time, leave_time, created_time, updated_time, building, unit, room FROM %s WHERE user_id = ? AND community_id = ?`, m.table)
	// ... rest unchanged
}
```

**FindByUserId** (约 76 行):
```go
func (m *defaultUserCommunityMembershipModel) FindByUserId(ctx context.Context, userId int64) ([]*UserCommunityMembership, error) {
	query := fmt.Sprintf(`SELECT id, user_id, community_id, bind_status, join_time, leave_time, created_time, updated_time, building, unit, room FROM %s WHERE user_id = ? AND bind_status = ?`, m.table)
	// ... rest unchanged
}
```

- [ ] **Step 7: 在 vars.go 增加地址重复错误**

在 `services/user-service/model/vars.go` 增加：

```go
// ErrAddressAlreadyTaken 同小区同地址已有人加入
var ErrAddressAlreadyTaken = errors.New("该地址已有人加入")
```

- [ ] **Step 8: 编译验证**

```bash
cd services/user-service && go build ./...
```

预期: 编译通过，无报错。

- [ ] **Step 9: 提交**

```bash
git add services/user-service/model/ services/user-service/migration/
git commit -m "feat(user-service): add building/unit/room to membership model

- Migration 003 adds building, unit, room columns + address index
- Model struct extended with Building/Unit/Room fields
- New FindByAddress query for uniqueness check
- New ErrAddressAlreadyTaken sentinel

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 2: RPC Logic — 加入小区地址校验（user-service 子 Claude）

**Files:**
- Modify: `services/user-service/rpc/internal/logic/user/join_community_logic.go:88-97`（Insert 段）
- Modify: `services/user-service/rpc/internal/logic/user/helper.go:50-67`（toProtoMembership）

- [ ] **Step 1: join_community_logic.go — 加入地址唯一性校验**

编辑 `join_community_logic.go`，在步骤 2（检查是否已加入同小区）之后、步骤 3（插入记录）之前，增加地址唯一性校验：

```go
	// 2.5. 校验同小区同地址唯一性（building+unit+room）
	if in.Building > 0 && in.Room > 0 {
		addrExisting, err := l.svcCtx.UserCommunityMembershipModel.FindByAddress(
			l.ctx, in.CommunityId, int(in.Building), int(in.Unit), int(in.Room))
		if err != nil && err != model.ErrNotFound {
			l.Errorf("check address uniqueness error: %v", err)
			return nil, err
		}
		if addrExisting != nil {
			return &userv1.JoinCommunityResponse{
				Base: responsex.NewBaseRespWithError(10011, "该地址已有人加入"),
			}, nil
		}
	}
```

> **错误码说明**: `10011` 是新分配的地址重复错误码。Proto 头部注释已占用 10001-10010，10040 是参数校验。10011 位于可用范围内。

- [ ] **Step 2: join_community_logic.go — 写入 building/unit/room**

编辑 `join_community_logic.go`，在 Insert 之前构造 membership 时增加新字段：

```go
	membership := &model.UserCommunityMembership{
		Id:          snowflake.NextID(),
		UserId:      in.UserId,
		CommunityId: in.CommunityId,
		BindStatus:  model.MembershipBindStatusActive,
		JoinTime:    now,
		CreatedTime: now,
		UpdatedTime: now,
		Building:    int(in.Building),  // 🆕
		Unit:        int(in.Unit),      // 🆕
		Room:        int(in.Room),      // 🆕
	}
```

同时更新 re-activate 分支（约 70-87 行），也需要写入 building/unit/room：

```go
	if existing != nil {
		// 重新激活时也更新地址信息
		existing.Building = int(in.Building)
		existing.Unit = int(in.Unit)
		existing.Room = int(in.Room)
		err = l.svcCtx.UserCommunityMembershipModel.UpdateBindStatus(l.ctx, existing.Id, model.MembershipBindStatusActive, now)
		// ... 同时更新 address 字段
		_ = l.svcCtx.UserCommunityMembershipModel.UpdateAddress(l.ctx, existing.Id, int(in.Building), int(in.Unit), int(in.Room))
```

> 注意：需要在 Model 层增加 `UpdateAddress` 方法。参见 Step 4。

- [ ] **Step 3: helper.go — toProtoMembership 增加 building/unit/room**

编辑 `helper.go` 的 `toProtoMembership` 函数：

```go
func toProtoMembership(m *model.UserCommunityMembership) *userv1.CommunityMembership {
	if m == nil {
		return nil
	}
	cm := &userv1.CommunityMembership{
		Id:          m.Id,
		UserId:      m.UserId,
		CommunityId: m.CommunityId,
		BindStatus:  int32(m.BindStatus),
		JoinTime:    m.JoinTime.Unix(),
		CreatedAt:   m.CreatedTime.Unix(),
		UpdatedAt:   m.UpdatedTime.Unix(),
		Building:    int32(m.Building),  // 🆕
		Unit:        int32(m.Unit),      // 🆕
		Room:        int32(m.Room),      // 🆕
	}
	if m.LeaveTime.Valid {
		cm.LeaveTime = m.LeaveTime.Time.Unix()
	}
	return cm
}
```

- [ ] **Step 4: Model 层 — 增加 UpdateAddress 方法**

在 `user_community_membership.go` 的 interface 中增加：

```go
	// 🆕 更新地址信息（重新激活时使用）
	UpdateAddress(ctx context.Context, id int64, building, unit, room int) error
```

实现：

```go
func (m *defaultUserCommunityMembershipModel) UpdateAddress(ctx context.Context, id int64, building, unit, room int) error {
	query := fmt.Sprintf(`UPDATE %s SET building=?, unit=?, room=?, updated_time=? WHERE id=?`, m.table)
	_, err := m.conn.ExecCtx(ctx, query, building, unit, room, time.Now(), id)
	return err
}
```

- [ ] **Step 5: 编译 + 测试**

```bash
cd services/user-service && go build ./...
cd services/user-service && go test ./... -count=1
```

预期: 编译通过，已有测试通过。

- [ ] **Step 6: 提交**

```bash
git add services/user-service/rpc/ services/user-service/model/
git commit -m "feat(user-service): add address validation to JoinCommunity

- JoinCommunity validates building/unit/room uniqueness
- New error code 10011 for address conflict
- toProtoMembership includes building/unit/room
- UpdateAddress method for re-activation

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 3: REST Types 变更（user-service 子 Claude）

**Files:**
- Modify: `services/user-service/api/internal/types/types.go:88-117`

- [ ] **Step 1: 扩展 JoinCommunityReq**

编辑 `types.go`，将 `JoinCommunityReq` 改为：

```go
type JoinCommunityReq struct {
	CommunityId int64 `json:"community_id,string"`
	Building    int   `json:"building"` // 🆕 楼号
	Unit        int   `json:"unit"`     // 🆕 单元号
	Room        int   `json:"room"`     // 🆕 房号
}
```

- [ ] **Step 2: 扩展 CommunityMembership**

在 `CommunityMembership` struct 末尾（`UpdatedAt` 之后）增加：

```go
type CommunityMembership struct {
	Id          int64  `json:"id,string"`
	UserId      int64  `json:"user_id,string"`
	CommunityId int64  `json:"community_id,string"`
	BindStatus  int32  `json:"bind_status"`
	JoinTime    int64  `json:"join_time"`
	LeaveTime   int64  `json:"leave_time"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	Building    int    `json:"building"` // 🆕
	Unit        int    `json:"unit"`     // 🆕
	Room        int    `json:"room"`     // 🆕
}
```

- [ ] **Step 3: 编译验证**

```bash
cd services/user-service && go build ./...
```

预期: 编译通过。

- [ ] **Step 4: 提交**

```bash
git add services/user-service/api/
git commit -m "feat(user-service): add building/unit/room to REST types

Extend JoinCommunityReq and CommunityMembership with address fields.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 4: 前端 API 层扩展（前端子 Claude）

**Files:**
- Modify: `web/mobile/src/api/user.ts`

- [ ] **Step 1: 扩展 joinCommunity 参数类型**

编辑 `web/mobile/src/api/user.ts`，找到 `joinCommunity` 函数，改为：

```ts
export interface JoinCommunityParams {
  community_id: string
  building: number   // 🆕
  unit: number       // 🆕
  room: number       // 🆕
}

export function joinCommunity(params: JoinCommunityParams): Promise<JoinCommunityResponse> {
  return request.post('/api/users/communities/join', params)
}
```

- [ ] **Step 2: 扩展 CommunityMembership 类型**

在同一文件中扩展返回类型：

```ts
export interface CommunityMembership {
  id: string
  user_id: string
  community_id: string
  bind_status: number
  join_time: number
  leave_time: number
  created_at: number
  updated_at: number
  building: number   // 🆕
  unit: number       // 🆕
  room: number       // 🆕
}
```

- [ ] **Step 3: 提交**

```bash
git add web/mobile/src/api/user.ts
git commit -m "feat(mobile): add building/unit/room to joinCommunity API types

Extend JoinCommunityParams and CommunityMembership with address fields.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 5: 加入小区 — Step 5 输入地址（前端子 Claude）

**Files:**
- Modify: `web/mobile/src/pages/join-community/join-community.vue`

> ⚠️ 当前文件约 356 行。需在 Step 4（搜索选中小区）之后插入 Step 5。

- [ ] **Step 1: 增加 Step 5 的状态变量**

在 `<script setup>` 顶部附近（与 `step` 变量同级）增加：

```ts
const step = ref(1) // 现有，1-4

// 🆕 Step 5 地址输入
const step5building = ref('')
const step5unit = ref('')
const step5room = ref('')
const selectedCommunity = ref<ResidentialArea | null>(null) // Step 4 选中的小区
```

- [ ] **Step 2: 修改 Step 4 的选择逻辑**

找到 Step 4 中用户点击「加入」按钮的处理函数，改为进入 Step 5 而非直接提交：

```ts
// 旧: await joinCommunity(area.id)
// 新:
function selectCommunity(area: ResidentialArea) {
  selectedCommunity.value = area
  step.value = 5
}
```

- [ ] **Step 3: 增加 Step 5 模板**

在 Step 4 的 `</template>` 之后，增加：

```vue
<!-- Step 5: 输入楼号-单元-房号 -->
<template v-if="step === 5">
  <!-- 已选小区信息 -->
  <view class="selected-community">
    <view class="community-name">🏘️ {{ selectedCommunity?.name }}</view>
    <view class="community-addr">{{ selectedCommunity?.cityName }} {{ selectedCommunity?.countyName }}</view>
  </view>

  <!-- 格式示例 -->
  <view class="address-example">
    <text class="example-label">示例</text>
    <text class="example-text">5号楼 2单元 301房间：<text class="highlight">5-2-301</text></text>
  </view>

  <!-- 三个输入框 -->
  <view class="address-inputs">
    <view class="input-group">
      <text class="input-label">楼号</text>
      <input v-model="step5building" type="number" placeholder="如 5" class="addr-input" />
    </view>
    <text class="input-sep">-</text>
    <view class="input-group input-unit">
      <text class="input-label">单元号</text>
      <input v-model="step5unit" type="number" placeholder="2" class="addr-input" />
    </view>
    <text class="input-sep">-</text>
    <view class="input-group">
      <text class="input-label">房号</text>
      <input v-model="step5room" type="number" placeholder="301" class="addr-input" />
    </view>
  </view>

  <!-- 校验错误输入框红框样式通过 :class 绑定 -->
  <!-- buildingError/unitError/roomError 为 ref(false)，校验失败时设为 true -->

  <button class="submit-btn" @click="submitJoin">确认加入</button>
</template>
```

- [ ] **Step 4: 实现校验函数**

```ts
// 每个输入框的错误状态（用于红框样式）
const buildingError = ref(false)
const unitError = ref(false)
const roomError = ref(false)

function validate(): string | null {
  // 重置错误状态
  buildingError.value = false
  unitError.value = false
  roomError.value = false

  const b = parseInt(step5building.value)
  if (isNaN(b) || b < 1 || b > 150) {
    buildingError.value = true
    return '楼号必须为数字，且不大于150'
  }
  const u = parseInt(step5unit.value)
  if (isNaN(u) || u < 1 || u > 5) {
    unitError.value = true
    return '单元号必须为数字，且不大于5'
  }
  const r = step5room.value
  if (!/^\d{3}$/.test(r)) {
    roomError.value = true
    return '房号必须为3位数字'
  }
  return null
}
```

- [ ] **Step 5: 实现 submitJoin 函数**

```ts
async function submitJoin() {
  // 前置校验：最多 3 个小区
  if (myCommunities.value.length >= 3) {
    uni.showToast({ title: '最多加入 3 个小区', icon: 'none' })
    return
  }

  // 输入校验
  const err = validate()
  if (err) {
    uni.showToast({ title: err, icon: 'none' })
    return
  }

  try {
    await joinCommunity({
      community_id: selectedCommunity.value!.id,
      building: parseInt(step5building.value),
      unit: parseInt(step5unit.value),
      room: parseInt(step5room.value),
    })

    // 更新 store
    communityStore.addCommunity({
      communityId: selectedCommunity.value!.id,
      communityName: selectedCommunity.value!.name,
    })

    // 显示成功 → 回首页
    showSuccess.value = true
  } catch (e: any) {
    uni.showToast({ title: e.message || '加入失败', icon: 'none' })
  }
}
```

- [ ] **Step 6: 提交**

```bash
git add web/mobile/src/pages/join-community/join-community.vue
git commit -m "feat(mobile): add Step 5 building/unit/room input to join flow

- Step 4 selects community, Step 5 inputs building/unit/room
- Validation at submit: building 1-150, unit 1-5, room 3 digits
- Toast errors with red border on invalid fields

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 6: 新建退出小区子页面（前端子 Claude）

**Files:**
- Create: `web/mobile/src/pages/leave-community/leave-community.vue`
- Modify: `web/mobile/src/pages.json`

- [ ] **Step 1: 注册路由**

编辑 `web/mobile/src/pages.json`，在 `pages` 数组中增加：

```json
{
  "path": "pages/leave-community/leave-community",
  "style": {
    "navigationBarTitleText": "退出小区",
    "navigationStyle": "custom"
  }
}
```

- [ ] **Step 2: 创建 leave-community.vue**

创建 `web/mobile/src/pages/leave-community/leave-community.vue`：

```vue
<template>
  <view class="page">
    <!-- 导航栏 -->
    <view class="nav-bar">
      <view class="nav-back" @click="goBack">← 返回</view>
      <text class="nav-title">退出小区</text>
    </view>

    <!-- 安全提示 -->
    <view class="warning-banner">
      ⚠️ 退出后将不再接收该小区的通知和公告。退出后可重新申请加入。
    </view>

    <!-- 已加入小区列表 -->
    <view class="section">
      <text class="section-title">已加入的小区（{{ communities.length }}个）</text>

      <view v-for="c in communities" :key="c.communityId" class="community-card">
        <view class="card-left">
          <view class="card-name">
            🏘️ {{ c.communityName }}
            <text v-if="c.communityId === currentId" class="tag-current">当前</text>
          </view>
          <text class="card-address">{{ c.address || '' }}</text>
        </view>
        <view class="btn-leave" @click="confirmLeave(c)">退出</view>
      </view>

      <view v-if="communities.length === 0" class="empty">
        暂未加入任何小区
      </view>
    </view>

    <!-- 确认弹窗 -->
    <view v-if="showModal" class="modal-mask" @click="showModal = false">
      <view class="modal-box" @click.stop>
        <text class="modal-icon">⚠️</text>
        <text class="modal-title">确认退出？</text>
        <text class="modal-desc">退出「{{ targetCommunity?.communityName }}」后，将不再接收该小区的通知公告等信息。</text>
        <view class="modal-btns">
          <view class="btn-cancel" @click="showModal = false">取消</view>
          <view class="btn-confirm" @click="doLeave">确认退出</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useCommunityStore } from '@/stores/community'
import { leaveCommunity } from '@/api/user'

const communityStore = useCommunityStore()
const communities = computed(() => communityStore.communities)
const currentId = computed(() => communityStore.currentCommunityId)

const showModal = ref(false)
const targetCommunity = ref<CommunityInfo | null>(null)

function goBack() {
  uni.navigateBack()
}

function confirmLeave(c: CommunityInfo) {
  targetCommunity.value = c
  showModal.value = true
}

async function doLeave() {
  if (!targetCommunity.value) return
  try {
    await leaveCommunity(targetCommunity.value.communityId)
    communityStore.removeCommunity(targetCommunity.value.communityId)
    uni.showToast({ title: '已退出小区', icon: 'success' })
    showModal.value = false
  } catch (e: any) {
    uni.showToast({ title: e.message || '退出失败', icon: 'none' })
  }
}
</script>
```

- [ ] **Step 3: 确保与 communityStore 的接口匹配**

确认 `communityStore` 暴露了以下内容（`web/mobile/src/stores/community.ts` 已有）：
- `communities` — `CommunityInfo[]`
- `currentCommunityId` — `string`
- `removeCommunity(id: string)` — 移除小区

- [ ] **Step 4: 提交**

```bash
git add web/mobile/src/pages/leave-community/ web/mobile/src/pages.json
git commit -m "feat(mobile): add leave-community page

New sub-page for leaving communities: list joined communities,
confirmation dialog, soft-delete via leaveCommunity API.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 7: "我的" 页面 V9 卡片式改造（前端子 Claude）

**Files:**
- Modify: `web/mobile/src/pages/my/my.vue`

> ⚠️ 当前文件 527 行，旧版使用侧边栏布局。需要整体替换模板和样式。

- [ ] **Step 1: 替换模板 — V9 卡片式布局**

用以下内容替换 `my.vue` 的 `<template>` 部分：

```vue
<template>
  <view class="my-page">
    <!-- 头部渐变区 -->
    <view class="header">
      <view class="avatar-circle">
        <text class="avatar-emoji">👤</text>
      </view>
      <text class="user-label">当前用户</text>
      <text class="user-phone">{{ maskedPhone }}</text>
    </view>

    <!-- 小区管理 -->
    <view class="section">
      <view class="section-header">
        <text class="section-icon">🏘️</text>
        <text class="section-title">小区管理</text>
        <text class="community-count">已加入 {{ communityStore.communityCount }}/3</text>
      </view>

      <view class="card-row">
        <view class="card card-join" @click="goJoin">
          <view class="card-icon-box card-icon-gold">
            <text class="card-emoji">➕</text>
          </view>
          <text class="card-label">加入小区</text>
        </view>

        <view class="card card-leave" @click="goLeave">
          <view class="card-icon-box card-icon-red">
            <text class="card-emoji">🚪</text>
          </view>
          <text class="card-label">退出小区</text>
        </view>
      </view>
    </view>

    <!-- 设置 -->
    <view class="section">
      <view class="settings-box">
        <view class="settings-header">
          <text>⚙️</text>
          <text class="settings-title">设置</text>
        </view>
        <view class="setting-item" @click="goPage('personal')">
          <text>个人信息</text>
          <text class="arrow">→</text>
        </view>
        <view class="setting-item" @click="goPage('security')">
          <text>账号安全</text>
          <text class="arrow">→</text>
        </view>
        <view class="setting-item" @click="goPage('about')">
          <text>关于我们</text>
          <text class="arrow">→</text>
        </view>
      </view>
    </view>
  </view>
</template>
```

- [ ] **Step 2: 替换脚本 — 简化逻辑**

```vue
<script setup lang="ts">
import { computed } from 'vue'
import { useCommunityStore } from '@/stores/community'

const communityStore = useCommunityStore()

const maskedPhone = computed(() => {
  // 从 store 或本地存储获取手机号
  const phone = uni.getStorageSync('user_phone') || '185****1501'
  return phone
})

function goJoin() {
  uni.navigateTo({ url: '/pages/join-community/join-community' })
}

function goLeave() {
  uni.navigateTo({ url: '/pages/leave-community/leave-community' })
}

function goPage(name: string) {
  // 占位 — 后续实现具体设置页面
  uni.showToast({ title: `${name} 页面开发中`, icon: 'none' })
}
</script>
```

- [ ] **Step 3: 替换样式 — V9 视觉**

```vue
<style scoped>
.my-page {
  min-height: 100vh;
  background: #FFFFFF;
}

/* 头部 */
.header {
  background: linear-gradient(160deg, #D4B896 0%, #E8DCCF 30%, #FAF8F5 55%, #FFFFFF 80%);
  padding: 80rpx 0 50rpx;
  text-align: center;
}
.avatar-circle {
  width: 140rpx; height: 140rpx;
  border-radius: 50%;
  background: rgba(255,255,255,0.8);
  border: 3rpx solid rgba(184,149,106,0.25);
  margin: 0 auto 24rpx;
  display: flex; align-items: center; justify-content: center;
  box-shadow: 0 4rpx 20rpx rgba(0,0,0,0.06);
}
.avatar-emoji { font-size: 72rpx; }
.user-label { font-size: 24rpx; color: #A6988A; display: block; margin-bottom: 6rpx; }
.user-phone { font-size: 32rpx; font-weight: 600; color: #3D3226; }

/* 通用 */
.section { padding: 0 40rpx; margin-top: 32rpx; }
.section-header { display: flex; align-items: center; gap: 10rpx; margin-bottom: 28rpx; }
.section-icon { font-size: 36rpx; }
.section-title { font-size: 34rpx; font-weight: 600; color: #3D3226; }
.community-count { margin-left: auto; font-size: 22rpx; color: #B8956A;
  background: rgba(184,149,106,0.08); padding: 4rpx 14rpx; border-radius: 20rpx; }

/* 卡片 */
.card-row { display: flex; gap: 20rpx; }
.card {
  flex: 1; height: 260rpx; background: #FAF8F5; border-radius: 24rpx;
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  box-shadow: 0 4rpx 16rpx rgba(0,0,0,0.04);
}
.card-icon-box {
  width: 160rpx; height: 160rpx; border-radius: 32rpx;
  display: flex; align-items: center; justify-content: center; margin-bottom: 16rpx;
}
.card-icon-gold { background: linear-gradient(135deg, #B8956A, #D4B896); }
.card-icon-red { background: linear-gradient(135deg, #D4958A, #E0ADA5); }
.card-emoji { font-size: 84rpx; color: #fff; line-height: 1; }
.card-label { font-size: 28rpx; font-weight: 500; color: #3D3226; }

/* 设置 */
.settings-box {
  background: #FAF8F5; border-radius: 20rpx; padding: 28rpx 32rpx;
}
.settings-header { display: flex; align-items: center; gap: 8rpx; margin-bottom: 24rpx; }
.settings-title { font-size: 30rpx; font-weight: 600; color: #3D3226; }
.setting-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: 14rpx 0; border-bottom: 1rpx solid rgba(184,149,106,0.1);
  font-size: 28rpx; color: #3D3226;
}
.setting-item:last-child { border-bottom: none; }
.arrow { font-size: 24rpx; color: #CCC4BA; }
</style>
```

- [ ] **Step 4: 编译验证**

```bash
cd web/mobile && npx vue-tsc --noEmit 2>&1 | head -20
```

预期: 无新增类型错误（或仅有预先存在的警告）。

- [ ] **Step 5: 提交**

```bash
git add web/mobile/src/pages/my/my.vue
git commit -m "feat(mobile): redesign 'My' page to V9 card layout

Replace sidebar layout with gradient header + two action cards
(join/leave community) + settings menu. Cards navigate to existing
join-community page and new leave-community page.

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### 实现顺序与依赖

```
Task 0 (Proto) ──必须最先──▶ Task 1 (DB+Model)
                                 │
                    ┌────────────┤
                    ▼            ▼
              Task 2 (RPC)   Task 4 (前端 API)
                    │            │
                    ▼            ▼
              Task 3 (REST)  Task 5 (join Step5)
                                 │
                                 ▼
                           Task 6 (leave pg)
                                 │
                                 ▼
                           Task 7 (my V9)
```

- Task 0: 全局 Claude 执行
- Task 1-3: user-service 子 Claude 串行执行
- Task 4-7: 前端子 Claude 串行执行（Task 4 需等 Task 0 完成后才能更新类型）
- Task 1-3 与 Task 4-7 **可并行**
