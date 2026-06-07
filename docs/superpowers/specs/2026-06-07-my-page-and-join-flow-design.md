# "我的" 页面 & 加入小区流程 — 设计 Spec

**日期**: 2026-06-07  
**状态**: 待实现  
**设计会话**: `.superpowers/brainstorm/68354-1780840775/`

---

## 1. 概述

本次设计涉及两个紧密关联的页面改造：

| 页面 | 现状 | 目标 |
|------|------|------|
| "我的" 页面 (`my.vue`) | 旧版侧边栏布局 | V9 卡片式：头像区 + 加入/退出卡片 + 设置 |
| 加入小区 (`join-community.vue`) | 4 步向导（省→市→区→搜索） | 5 步：省→市→区→搜索→**输入楼号-单元-房号** |

退出小区复用已有 API（`leaveCommunity`），仅改造 UI。

---

## 2. "我的" 页面（V9）

### 2.1 页面结构

```
┌──────────────────────────────┐
│         头像区（渐变）         │  ← 用户头像 + 手机号
│      👤 185****1501          │
├──────────────────────────────┤
│  🏘️ 小区管理        已加入2/3 │
│                              │
│  ┌────────┐  ┌────────┐     │
│  │  ➕    │  │  🚪    │     │  ← 两个卡片横排
│  │ 加入小区│  │ 退出小区│     │
│  └────────┘  └────────┘     │
├──────────────────────────────┤
│  ⚙️ 设置                     │
│    个人信息              →   │
│    账号安全              →   │
│    关于我们              →   │
└──────────────────────────────┘
```

### 2.2 交互行为

| 元素 | 行为 |
|------|------|
| 加入小区卡片 | `uni.navigateTo({ url: 'pages/join-community/join-community' })` |
| 退出小区卡片 | `uni.navigateTo({ url: 'pages/leave-community/leave-community' })`（新建页面） |
| 设置菜单项 | 跳转对应页面（个人信息/账号安全/关于我们） |

### 2.3 视觉规范

- 头部渐变: `linear-gradient(160deg, #D4B896 0%, #E8DCCF 30%, #FAF8F5 55%, #FFFFFF 80%)`
- 加入卡片: 金色系 `linear-gradient(135deg, #B8956A, #D4B896)`
- 退出卡片: 暖红色系 `linear-gradient(135deg, #D4958A, #E0ADA5)`
- 卡片尺寸: 等高 260rpx，圆角 24rpx
- 图标区域: 160rpx × 160rpx，圆角 32rpx，emoji 84rpx
- 设置区: 背景 `#FAF8F5`，圆角 20rpx
- 目标设备宽度: 390px（375rpx 内容区 + 40rpx padding）

---

## 3. 退出小区子页面（新建）

### 3.1 页面结构

```
┌──────────────────────────────┐
│ ← 返回    退出小区            │
├──────────────────────────────┤
│ ⚠️ 退出后将不再接收该小区的     │  ← 安全提示
│    通知和公告。可重新申请加入。  │
├──────────────────────────────┤
│ 已加入的小区（2个）            │
│                              │
│ ┌────────────────────────┐   │
│ │ 🏘️ 阳光花园小区  [当前]  │   │
│ │ 深圳市南山区 · 2025-03  │   │
│ │                  [退出] │   │  ← 每行一个退出按钮
│ └────────────────────────┘   │
│ ┌────────────────────────┐   │
│ │ 🏘️ 翠竹苑小区           │   │
│ │ 深圳市罗湖区 · 2025-06  │   │
│ │                  [退出] │   │
│ └────────────────────────┘   │
│                              │
│   ┌─── 确认弹窗（点击退出后）──┐  │
│   │       ⚠️                │  │
│   │    确认退出？            │  │
│   │ 退出「阳光花园小区」后    │  │
│   │ 将不再接收通知           │  │
│   │                         │  │
│   │  [取消]    [确认退出]    │  │
│   └─────────────────────────┘  │
└──────────────────────────────┘
```

### 3.2 交互行为

1. 页面加载：调用 `getUserMemberships()` 获取已加入小区列表
2. 点击「退出」→ 弹出确认弹窗
3. 点击「确认退出」→ 调用 `leaveCommunity(communityId)` → 成功后 `communityStore.removeCommunity(id)`
4. 退出的是当前小区 → store 自动切换到剩余第一个
5. 后端软删除：`bind_status` 设为 0，`leave_time` 记录时间

### 3.3 数据来源

- 小区列表: `communityStore.communities`（或直接调 `getUserMemberships()`）
- API: `POST /api/users/communities/leave`（已有，无需变更）

---

## 4. 加入小区流程（5 步）

### 4.1 流程步骤

| 步骤 | 内容 | 数据来源 | 状态 |
|------|------|---------|:---:|
| 1. 省 | 省份列表 | `getDivisions()` | 已有 |
| 2. 市 | 城市列表 | `getDivisions(parentId)` | 已有 |
| 3. 区 | 区县列表 | `getDivisions(parentId)` | 已有 |
| 4. 搜索 | 搜索小区 | `searchResidentialAreas({keyword, countyId})` | 已有 |
| 5. 输入地址 | 楼号-单元-房号 | 用户输入 + 提交时校验 | 🆕 |

### 4.2 步骤 5: 输入地址

```
┌──────────────────────────────┐
│ ← 返回    加入小区            │
├──────────────────────────────┤
│ 🏘️ 阳光花园小区               │  ← 已选小区信息
│    广东省深圳市南山区          │
├──────────────────────────────┤
│ ┌─────────────────────────┐  │
│ │ 示例                     │  │  ← 格式提示
│ │ 5号楼 2单元 301房间：     │  │
│ │          5-2-301         │  │
│ └─────────────────────────┘  │
├──────────────────────────────┤
│  楼号  -  单元号  -  房号    │  ← 三个输入框
│ [  5 ] - [  2  ] - [301 ]   │
├──────────────────────────────┤
│      [ 确认加入 ]            │  ← 提交按钮
└──────────────────────────────┘
```

### 4.3 校验规则（提交时执行）

界面**不显示内联错误提示**，保持干净。点击「确认加入」后执行：

| 字段 | 规则 | 不通过 → toast | 不通过 → UI |
|------|------|---------------|------------|
| 楼号 | 必填 · 纯数字 · 1 ≤ n ≤ 150 | "楼号必须为数字，且不大于150" | 输入框红框 |
| 单元号 | 必填 · 纯数字 · 1 ≤ n ≤ 5 | "单元号必须为数字，且不大于5" | 输入框红框 |
| 房号 | 必填 · 恰好 3 位数字 | "房号必须为3位数字" | 输入框红框 |
| 唯一性 | community_id+building+unit+room 不重复 | "该地址已有人加入" | toast 提示 |

校验通过 → 调用 `joinCommunity({ community_id, building, unit, room })`。

### 4.4 空状态 & 边界情况

- 用户未加入任何小区：主页两个卡片均显示（加入 + 退出），退出卡片点击后显示"暂未加入任何小区"
- 用户已加入 3 个小区（上限）：加入卡片点击后 toast "最多加入 3 个小区"
- 用户只剩下 1 个小区时退出：允许退出（store 自动清空 currentCommunityId）

---

## 5. Proto 变更

### 5.1 JoinCommunityRequest 扩展

```protobuf
// api-proto/api/user/v1/user.proto
message JoinCommunityRequest {
  int64 user_id = 1 [jstype = JS_STRING];
  int64 community_id = 2 [jstype = JS_STRING];
  int32 building = 3;   // 🆕 楼号, 1-150
  int32 unit = 4;       // 🆕 单元号, 1-5
  int32 room = 5;       // 🆕 房号, 3位数字
}
```

### 5.2 CommunityMembership 扩展（可选）

如需在成员列表/小区切换中展示地址：

```protobuf
message CommunityMembership {
  // ... existing fields ...
  int32 building = 9;   // 🆕
  int32 unit = 10;      // 🆕
  int32 room = 11;      // 🆕
}
```

### 5.3 操作流程

```
1. 修改 api-proto/api/user/v1/user.proto
2. cd api-proto && make generate
3. cd api-proto && make lint && make breaking-check
4. 更新 api-proto/CHANGELOG.md
5. 通知 user-service 重新构建
```

---

## 6. 数据库变更

### 6.1 Migration

```sql
ALTER TABLE user_community_membership
  ADD COLUMN building INT DEFAULT 0 COMMENT '楼号',
  ADD COLUMN unit INT DEFAULT 0 COMMENT '单元号',
  ADD COLUMN room INT DEFAULT 0 COMMENT '房号';

-- 唯一索引（防止同一地址重复加入）
CREATE UNIQUE INDEX idx_unique_address 
  ON user_community_membership(community_id, building, unit, room) 
  WHERE bind_status = 1;
```

### 6.2 Model 变更

`services/user-service/model/user_community_membership.go`:
```go
type UserCommunityMembership struct {
    // ... existing fields ...
    Building int `gorm:"column:building;default:0"`
    Unit     int `gorm:"column:unit;default:0"`
    Room     int `gorm:"column:room;default:0"`
}
```

---

## 7. 后端逻辑变更

### 7.1 join_community_logic.go

在 `JoinCommunity` 方法中增加：

```go
// 校验唯一性（同小区同地址不可重复加入）
existing, err := l.svcCtx.Model.FindByAddress(
    in.CommunityId, int(in.Building), int(in.Unit), int(in.Room))
if err == nil && existing != nil {
    return nil, errx.NewCodeError(10008, "该地址已有人加入")
}

// 写入 building/unit/room 到新记录
membership.Building = int(in.Building)
membership.Unit = int(in.Unit)
membership.Room = int(in.Room)
```

### 7.2 REST types 变更

`services/user-service/api/internal/types/types.go`:
```go
type JoinCommunityReq struct {
    CommunityId int64 `json:"community_id,string"`
    Building    int   `json:"building"`    // 🆕
    Unit        int   `json:"unit"`        // 🆕
    Room        int   `json:"room"`        // 🆕
}
```

---

## 8. 前端变更

### 8.1 改造 join-community.vue

在 Step 4（搜索选中小区）之后，增加 Step 5：

```vue
<!-- Step 5: 输入楼号-单元-房号 -->
<template v-if="step === 5">
  <!-- 已选小区信息 -->
  <view class="community-info">...</view>
  
  <!-- 格式示例 -->
  <view class="example">
    5号楼 2单元 301房间：<text class="highlight">5-2-301</text>
  </view>
  
  <!-- 三个输入框 -->
  <view class="address-inputs">
    <input v-model="building" type="number" placeholder="如 5" />
    <text class="sep">-</text>
    <input v-model="unit" type="number" placeholder="2" />
    <text class="sep">-</text>
    <input v-model="room" type="number" placeholder="301" />
  </view>
  
  <button @click="submitJoin">确认加入</button>
</template>
```

校验逻辑：
```ts
function validate(): string | null {
  const b = parseInt(building.value)
  if (isNaN(b) || b < 1 || b > 150) return '楼号必须为数字，且不大于150'
  const u = parseInt(unit.value)
  if (isNaN(u) || u < 1 || u > 5) return '单元号必须为数字，且不大于5'
  const r = room.value
  if (!/^\d{3}$/.test(r)) return '房号必须为3位数字'
  return null // pass
}

async function submitJoin() {
  const err = validate()
  if (err) {
    uni.showToast({ title: err, icon: 'none' })
    // 对应输入框加红框样式
    return
  }
  await joinCommunity({ community_id, building: +building, unit: +unit, room: +room })
  communityStore.addCommunity({ ... })
  // 显示成功 → 回首页
}
```

### 8.2 新建 leave-community.vue

从 `my.vue` 中提取退出小区逻辑，独立为新页面：

- 路由: `pages/leave-community/leave-community`
- 加载 `communityStore.communities`（或调 API）
- 点击「退出」→ 弹窗确认 → 调 `leaveCommunity(id)` → `removeCommunity(id)`

### 8.3 改造 my.vue

从旧版侧边栏改为 V9 卡片式：

- 头部渐变背景 + 头像 + 手机号
- 两个等大卡片横排（加入/退出），点击跳转
- 设置菜单（个人信息/账号安全/关于我们）
- 顶部显示 `已加入 N/3` 标签

### 8.4 pages.json 路由注册

```json
{
  "pages": [
    // ... existing ...
    {
      "path": "pages/leave-community/leave-community",
      "style": { "navigationBarTitleText": "退出小区" }
    }
  ]
}
```

---

## 9. 前端 API 层变更

`web/mobile/src/api/user.ts`:

```ts
// joinCommunity 参数扩展
export function joinCommunity(params: {
  community_id: string
  building: number
  unit: number
  room: number
}): Promise<JoinCommunityResponse> {
  return request.post('/api/users/communities/join', params)
}
```

---

## 10. 实现顺序

| 阶段 | 内容 | 负责 |
|------|------|------|
| 1. Proto | 修改 `JoinCommunityRequest` + 生成代码 | 全局 Claude |
| 2. DB | Migration 加 building/unit/room 列 + 唯一索引 | user-service 子 Claude |
| 3. Model | Go struct 加字段 | user-service 子 Claude |
| 4. RPC logic | join_community_logic 加校验 + 写入新字段 | user-service 子 Claude |
| 5. REST types | JoinCommunityReq 加字段 | user-service 子 Claude |
| 6. 前端 API | `joinCommunity()` 参数扩展 | 前端子 Claude |
| 7. 前端 join-community | Step 5 输入步骤 + 校验 | 前端子 Claude |
| 8. 前端 leave-community | 新建退出子页面 | 前端子 Claude |
| 9. 前端 my | V9 卡片式改造 | 前端子 Claude |

阶段 1 必须在其他阶段之前完成。阶段 2-5 可由 user-service 子 Claude 串行完成。阶段 6-9 可与 2-5 并行。

---

## 11. 未纳入本次的范围

- 退出小区后「当前」标签的小区是否禁止退出（本次不做限制）
- 房号已占用标记 API（Phase 2，本次不做）
- 加入小区后填写房号的"后补"模式（本次不做）
- 按楼栋配置总层数（本次不做，固定规则生成）

---

## 12. 设计文件索引

| 文件 | 内容 |
|------|------|
| `my-page-design.html` | V1 初始设计 |
| `my-page-v6.html` | V6 纵排大卡片 |
| `my-page-v7.html` | V7 大图标 |
| `my-page-v8.html` | V8 三屏布局 |
| `my-page-v9.html` | V9 最终定稿（复用+新建子页面） |
| `join-flow-v1.html` | 7 步流程 |
| `join-flow-v2.html` | 混合模式定稿 |
| `join-flow-v3.html` | 简化版（输入替代选择） |
| `join-flow-v4.html` | 最终版（提交校验+干净界面） |
| `final-review.html` | 总览页 |
