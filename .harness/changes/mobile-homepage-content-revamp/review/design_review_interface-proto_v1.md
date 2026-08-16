# Design Review — mobile-homepage-content-revamp（接口契约 + Proto 视角）

**审查模式**: 模式一.5 设计评审
**审查视角**: interface-proto（接口契约、Proto 破坏性、依赖顺序）+ data-model（数据模型、服务归属）
**审查对象**: `.harness/changes/mobile-homepage-content-revamp/design.md` + `tasks.md`
**对照基准**: `proposal.md` + `specs/{notice-time-window,notice-detail-preview,function-entries,contact-list-page,homepage-layout}/spec.md` + 实际代码（api-proto / community-hub-service / web/mobile）

## 摘要

- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 3

设计整体自洽：Proto 变更 additive 非破坏且字段号确认可用；REST 层 Base 错误上抛缺口（080005 静默吞错）被正确识别并独立成 Task 1.5；migration 004 DDL 与 001/模型逐列对齐；附件重生整单失败语义与既有 RPC 实现一致；前端类型扩展与 wire（file_id/file_type snake_case + string）对齐。无阻塞级问题。

## 已验证项（对照代码）

| 设计声明 | 验证结果 |
|---------|---------|
| `ListContentPostsRequest` 现占用 1-5 号字段，since_days=6 可用 | ✅ `community.proto:115-121`，additive 非破坏成立 |
| `int32 since_days` 非 ID 不需 jstype | ✅ CLAUDE.md(api-proto) 规则 5：jstype 仅约束 int64 ID |
| 080005 复用（CodeInvalidParam） | ✅ `scope.CodeInvalidParam = 80005`（scope.go:30），语义「参数无效」一致 |
| REST api logic 仅查 gRPC err 未查 resp.GetBase() | ✅ `api/internal/logic/notice/listcontentpostslogic.go:36-38`，Task 1.5 修复缺口真实存在 |
| getcontentpostlogic.go 的 ToError(Base) 既有模式 | ✅ `getcontentpostlogic.go:60-62`，Task 1.5 引用正确 |
| 附件重生失败→读整单失败（r2-6） | ✅ `getcontentpostlogic.go:100-103` toProtoAttachments 任一 GetFileUrl 失败 return (nil,err) |
| REST wire 已含 file_id/file_type（snake_case+string） | ✅ `types.go ContentPostAttachmentInfo` file_type/json:"file_type"、file_id/json:"file_id,string"；api helper.go 已映射 |
| migration 004 DDL 与 001/CommunityContactModel 逐列一致 | ✅ 001_initial.sql:39-49 与 model/community_contact.go 8 列 + idx_community 全对齐 |
| 首页跑马灯由同一 getNoticeList 派生（REQ-NTW-3 现状） | ✅ notice.vue:280-283 marqueeText 从 notices 派生 |
| 列表页现状「客户端 3 个月过滤 + 翻页」（Task 2.4 改造依据） | ✅ notice-browse.vue:112-114 threeMonthsAgo filter + currentIndex 翻页 |
| 首页现状「内嵌联络网格 + 2+1 广告分散」（Task 2.2/2.3 依据） | ✅ notice.vue:100-149 联络网格、120-150 广告×2、198-210 广告×1 |
| 任务粒度（模式一.5 刚性校验） | ✅ 单任务单服务；Proto(0.x)/Migration(1.1/1.2) 独立成任务；测试用例 ≤5；依赖顺序 全局→model→RPC→REST→mobile→运维验证 正确 |

## 发现

### 🔴 MUST FIX

无。

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|---------|------|------|
| S1 | design.md §索引设计 / Task 1.2 / ADR-3 | `idx_status_pinned_published (status, is_pinned, published_at)` 缺少 community_id：窗口查询实际驱动表为 content_post_scope（经既有 idx_scope_community(community_id, post_id)），content_posts 经主键 eq_ref 连接，优化器几乎必然不会选择该索引驱动（读全库 approved 行再回连 scope 过滤小区，代价远大于 scope 驱动）。REQ-NTW-6「不走 content_posts 全表扫描」由 JOIN 结构本身保证（PK eq_ref 无全表扫描），新索引大概率未被选用、filesort 仍在，成为热表（content_posts 频繁 kafka_push_status 更新）上的冗余索引。 | Task 3.2 EXPLAIN 验收须显式核对优化器实际选用的索引：若走 `idx_scope_community + filesort`（已满足 REQ-NTW-6 无全表扫描），应**删除 migration 005** 避免冗余索引维护成本；若确认选用新索引（stats 偏斜场景），再保留并记录。验收标准从「走任一索引」收紧为「核对实际执行计划并记录索引取舍」。 |
| S2 | tasks.md Task 0.1 字段注释 | 注释「非法值 ≤0 或 >365 → 080005」与同一条目「缺省 0=不过滤」矛盾：proto3 int32 无 presence，缺省==0，按注释实现「≤0 拒绝」会误伤全部 PC 既有调用方（不传 since_days→0→被拒→PC 列表故障）。RPC 侧 Task 1.4 实现 `<0 \|\| >365` 正确，但 Task 0.1 注释与 spec REQ-NTW-2 措辞（"values ≤0 ... rejected"）均与行为不一致。 | 将 Task 0.1 注释统一为「0=不过滤（缺省）；<0 或 >365 → 080005」，并同步修订 spec REQ-NTW-2 中 "values ≤0 or >365 SHALL be rejected" 措辞（0 为 additive 缺省哨兵，仅 <0 或 >365 拒绝），消除实现歧义。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| I1 | design.md §安全考虑 引用的 `file-service guard/magic.go` 路径缺 `internal/` 前缀（实际 `services/file-service/internal/guard/magic.go`）；且该文件 SniffType 只产出 png/jpg/gif/pdf/doc/docx，JPEG 一律嗅探为 `jpg`，wire **永不出现 "jpeg"**。白名单 `{png,jpg,jpeg,gif}` 为安全超集（无害，真实 JPEG 走 jpg），但「与 file-service 白名单对齐」表述不准——建议注释注明 wire 仅 png/jpg/gif 三种图片值（jpeg 为上传侧扩展名，嗅探归一为 jpg）。 |
| I2 | design §接口设计「非数字由 REST 网关解析层拒绝」与 spec r2-5「以同样参数无效错误呈现」措辞不完全一致：go-zero form 绑定 int32 失败返回的是绑定错误（HTTP 400，非 080005 语义码）。属网关边界可接受行为，但建议明确「网关绑定错误≠080005，服务端仅校验数值范围」，避免实现期误判。 |
| I3 | 任务粒度提示：Task 2.2/2.3/2.4/2.5 组件测试均标记「可选」，关键交互（附件分发分支、区块全序、列表窗口数据）全部下沉到 Task 3.2 端到端验收。若端到端验收被跳过/延误，前端交互回归无自动化兜底——建议至少保留 notice.spec.ts 中「4 入口顺序 + 点击行为 + since_days 传参」断言（Task 2.2 已列 RED 用例），其余可接受。 |

---

VERDICT: APPROVED
---
