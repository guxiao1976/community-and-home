# 通知模块设计 — 多小区发布 + 通栏跑马灯 + 附件安全

> 日期：2026-08-15
> 范围：移动端首页通栏通知 + 通知发布/浏览/详情 + 多角色发布权限 + 附件安全 + 可复用组件
> 前置：已确认「小区/村层级统筹」记 backlog（task-016），不阻塞本模块

## 一、目标

把现有「单小区通知」升级为「多小区通知 + 完整发布/浏览/详情链路」，作为首页核心内容 + 可复用模块：

1. **首页通栏通知跑马灯**：切换小区下方，最近 15 天通知标题滚动，右侧「更多 ›」
2. **通知浏览页**：按发布时间倒序（现有 notice-browse）
3. **通知详情页**：标题/正文/发布时间/附件（现有 notice-detail）
4. **多角色发布**：网格员（多小区可选）/ 社区管理员（选社区）/ 物业/业委（本小区），业主只读
5. **发布入口**：【我的】页，有权限才显示
6. **附件安全**：白名单（图片/pdf/word）+ 大小上限 + 禁可执行/脚本/压缩包
7. **可复用组件**：通知模块组件化

## 二、数据模型

### 2.1 notices（改造：弃用 community_id，多小区走 notice_scope）

```sql
notices
├── id             BIGINT PK  -- 雪花
├── title          VARCHAR(200)
├── content        TEXT
├── role           VARCHAR(20)   -- 发布角色 community/committee/property/grid_officer
├── publisher      VARCHAR(100)  -- 发布单位/人名
├── publisher_id   BIGINT
├── is_pinned      TINYINT
├── published_at   DATETIME
├── moderation_status TINYINT    -- 审核状态（现有）
├── created_at / updated_at / deleted_at
```
- **`community_id` 弃用**（兼容期保留列但不再写入新数据；查询改走 notice_scope）
- 单源原则：多小区范围全部走 notice_scope，避免双写不一致

### 2.2 notice_scope（新增关联表）

```sql
notice_scope
├── id             BIGINT PK
├── notice_id      BIGINT NOT NULL  -- → notices.id
├── community_id   BIGINT NOT NULL  -- → md_residential_area.id（小区或村）
├── created_at
├── UNIQUE uk_notice_community (notice_id, community_id)
```
- 一条通知关联 N 个小区（网格员多小区发布）
- 社区管理员「选社区」→ **发布时展开为具体小区 id 快照**写入（范围确定，撤回/查询简单）
- 撤回 = 删 notices 行（notice_scope 级联或按 notice_id 删）

### 2.3 notice_attachments（现有 + 加 file_type）

```sql
notice_attachments
├── id, notice_id
├── file_name, file_url, file_size
├── file_type    VARCHAR(20)  -- 新增：白名单校验依据（png/jpg/pdf/docx 等）
└── created_at
```

## 三、权限模型（复用现有 RBAC 数据权限）

发布前后端校验：`AssertPublishScope(user, target_community_ids)`

| 角色 | scope 层级 | 发布范围 | 前端发布入口 |
|---|---|---|---|
| 网格员 grid_worker | community | 多小区（发布时多选，均在 scope 内） | 显示 |
| 社区管理员 community_admin | community_div | 选社区 → 展开为下所有小区 | 显示 |
| 物业 property_admin | community | 本小区 | 显示 |
| 业委 committee | community | 本小区 | 显示 |
| 业主 owner/tenant | community | 只读（不显示发布入口） | 隐藏 |

- **前端不判断权限**：后端返回「可发布标志」（如 `can_publish`），驱动【我的】页入口显隐
- 范围单位都是 `md_residential_area.id`（小区或村），`community_type` 不影响范围表达

## 四、接口契约

### 4.1 Proto 变更（api-proto/community/v1）

- `CreateNoticeRequest`：`community_id`（单）→ 改 `repeated int64 community_ids`（多小区）；保留 `attachment_ids`
- `ListNoticesRequest`：`community_id` 保持（读路径按当前小区 scope 过滤，经 notice_scope 关联）
- 新增 `GetPublishPermission` 或复用现有：返回用户可发布角色 + `can_publish` 标志
- 撤回：复用 `DeleteNotice`

### 4.2 附件上传（file-service）

- 白名单：`png/jpg/jpeg/gif/pdf/doc/docx`
- 禁止：`exe/bat/sh/cmd/com/msi/apk/js/vbs/ps1/py/pl/php` + **所有 zip/rar**（不接收）
- 单文件 ≤ 10MB
- file-service 上传时校验类型 + 大小（通用安全规则，作为可复用中间件/校验器）

## 五、前端（移动端）

```
首页 notice（改造）
├── 通栏跑马灯 NoticeMarquee（切换小区下，最近15天通知标题滚动，更多→浏览页）
├── 公告卡片（现有）
├── 联络卡片 / 寻失卡片（现有）
浏览页 notice-browse：发布时间倒序
详情页 notice-detail：标题/正文/时间/附件列表
【我的】我的页：can_publish → 显示"发布通知"入口
发布表单：标题/正文/附件上传(白名单)/范围选择
         （网格员多选小区 / 社区管理员选社区展开 / 物业业委固定本小区）
```

## 六、可复用组件

| 组件 | 职责 | 复用点 |
|---|---|---|
| `NoticeMarquee` | 通栏跑马灯 | 首页/任何聚合页 |
| `NoticePublisher` | 发布表单（权限+范围+附件） | 【我的】页/管理端 |
| `NoticeList` | 通知列表（倒序） | 浏览页/首页卡片 |
| `NoticeDetail` | 详情+附件展示 | 详情页 |
| 附件安全校验 | file-service 白名单+大小 | 任何文件上传（通用安全规则） |

## 七、审核流（复用现有 moderation）

- 发布 → moderation-service 审核 → `moderation_status=通过` 才可见
- 现有 `UpdateNoticeModerationStatus` 保留

## 八、边界

- 撤回 = 发布者本人，全局生效（一条通知多小区一次撤）
- 首页跑马灯「更多」进浏览页；单条点击进详情页
- 附件 10MB 上限、白名单外类型拒绝、zip/rar 不接收

## 九、验收标准

- [ ] 首页通栏跑马灯展示最近 15 天通知标题，更多→浏览页（倒序）
- [ ] 网格员发布可选多个小区（均在 scope 内）；社区管理员选社区展开为小区
- [ ] 物业/业委仅本小区；业主不显示发布入口
- [ ] 附件白名单外拒绝、>10MB 拒绝、zip/rar 拒绝、exe/sh 拒绝
- [ ] 撤回全局生效（多小区一次撤）
- [ ] 通知经 moderation 审核通过才可见
