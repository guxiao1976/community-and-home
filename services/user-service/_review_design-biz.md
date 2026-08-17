# Code Review — 用户服务（设计业务视角）

**审查时间**: 2026-08-16 23:36 (本地)
**审查维度**: 设计一致性(#2)、代码质量(#4)、Migration(#8部分)

## 摘要
- 🔴 CRITICAL: 0 / 🟡 WARNING: 1 / 🔵 NOTE: 3

## 变更范围确认

本次工作树 diff 为「接线类型」修复：`GetProfileLogic.GetProfile` 调 `UserRpc.GetUser` 时补传 `ViewerId: &userId`（原为 0），使本人查自身 profile 命中 RPC `GetUser` 的 `viewerId == in.Id` 分支，返回明文手机号。配套测试 4 用例 + CHANGELOG + go.mod（miniredis direct）+ graph-context。

核验：
- proto `viewer_id = 2 [optional int64]` → Go `*int64`，`&viewerID` 传指针正确。
- RPC `GetUser` masking 语义（`get_user_logic.go`）：`viewerId==0`→maskPhone 脱敏（默认安全）；`==in.Id`→明文+自身房屋；其他→同屋判定。本次变更与既有实现语义一致，未改动 masking 逻辑。
- 未登录 → `getUserIdFromJwt` 返回 0 → 提前 error（不触达 RPC）。RPC Base 10001 → `GetProfile` 显式检查 `resp.Base.GetCode()` 后 error。错误路径完整。
- 安全默认保持：`GET /api/users/:id` 通用端点未传 ViewerId → 仍脱敏，无新增信息泄露路径。ViewerId 来源是 JWT user_id（服务端可信），非客户端可控。

## 发现

### 🟡 WARNING

| # | 文件:行号 | 维度 | 问题 | 建议 |
|---|----------|------|------|------|
| 1 | `docs/design.md` §7（交互表） | 设计一致性 | GetUser 的 `viewer_id` 可见性契约（0=脱敏 / ==target=明文+房屋 / 其他=同屋判定，默认安全）是安全敏感的核心行为，Task 3.6 已实现、本次变更是其第一个消费方，但 design.md 交互表只列了 `GetUserByPhone`/`UpdateUser`，完全未记录 viewer_id 脱敏契约。维护者只能靠 `[[phone-encryption]]` 记忆或 CHANGELOG 才能发现该行为，design.md 作为单一事实源已漂移 | design.md §7 补 GetUser 端点说明：`GetUser(viewer_id)` 的脱敏/同屋/明文规则 + 列出「本人查看」应传 viewer_id==user_id；同时注明 `GET /api/users/profile` 为本人查自身路径 |

### 🔵 NOTE

| # | 文件:行号 | 建议 |
|---|----------|------|
| 1 | `api/internal/logic/user/user_logic.go:110-114`（`GetUserLogic.GetUser`） | 通用 `GET /api/users/:id` 端点调 RPC 后**未检查** `resp.Base.GetCode()`，用户不存在（RPC 返回 10001、`User=nil`）时以 HTTP 200 返回空 `UserInfo{}`；而同文件 `GetProfile`（本次变更）正确检查了 Base。同一文件内错误处理不一致（既有缺陷，非本次引入），建议对齐 |
| 2 | `api/internal/logic/user/user_logic.go:182-198` + `api/internal/types/types.go:33` | 自查看时 RPC 返回 `SameHouse`（自身房屋号），但 `types.GetUserResp` 无该字段，`toUserInfo` 丢弃。本次目标仅手机号，但若移动端【我的】页后续要展示房屋号需另接线（或给 `GetUserResp` 补 `SameHouse`） |
| 3 | `api/internal/logic/user/user_logic.go:104-115` vs `:182-198` | `GET /api/users/:id`（未传 ViewerId→脱敏）与 `GET /api/users/profile`（明文）对**本人**查看行为不一致。保守安全但 UX 不一致，若前端某处用 `:id` 查自仍显示脱敏手机号，建议前端统一走 profile 端点 |

### Migration / 数据一致性核查（#8 部分）
- 本次变更**无 DB Migration**，无表结构/既有数据影响，无需回滚方案/锁表评估。
- `go.mod`：`miniredis/v2` 由 indirect 提升为 direct（`get_user_roles_logic_test.go` 直接 import 的 `go mod tidy` 卫生），无依赖语义变化。
- design_consistency WARN（`deleted_at` 模型列未覆盖迁移源）为历史遗留，与本次 diff 无关。

### 设计一致性正面确认
- 变更符合 2026-08-13 CHANGELOG Task 3.6 确立的「默认脱敏、显式解封」语义，不破坏不变量。
- 手机号明文仅返回给手机号所有者本人（JWT user_id），符合 PII 最小暴露原则。

---
VERDICT: PASS — 设计业务视角无 CRITICAL；变更语义正确、错误路径完整、无 Migration 风险。
---
