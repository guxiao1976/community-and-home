# Plan Review — notice-multicommunity-publish（业务有效性视角）

**审查维度**: 业务自洽 / 非功能 / 合规 / 安全
**审查版本**: P1.3（fallback:r2:rc1）— 磁盘最新内容独立审查（spec 已于 23:41-23:42 更新，本轮不沿用 v2 结论）
**审查时间**: 2026-08-15

## 摘要
- 🔴 MUST FIX: 1（CRITICAL，一票否决） / 🟡 SHOULD FIX: 4 / 🔵 INFO: 3
- VERDICT: **REVISION**

## 上轮（v2）SHOULD FIX 复验

| # | v2 问题 | 当前磁盘状态 | 结论 |
|---|---------|-------------|------|
| 1 | REQ-PP-1「0=永久」与 DB NULL=永久 冲突 | **已消解**：getuserroleslogic.go 将 DB NULL→RPC expires_at=0（`if r.ExpiresAt.Valid {...} else 0`），proto UserRoleInfo 注释「0=永久」；REQ-PP-1 明确经 GetUserRoles RPC（禁直读 rel_user_role），「expires_at in future or 0=永久」与 RPC 契约一致 | ✅ 无需修复 |
| 2 | 未知节点 080006 vs 展开空 080005 错误码族不一致 | REQ-NP-3「目标小区不存在」→080006；REQ-NP-4「展开后无小区」→080005。仍在，但 spec 已给出两语义的独立理由（未知节点安全拒绝 / 展开空=参数无效） | 🟡 SHOULD（前端错误契约需显式区分） |
| 3 | 附件绑定校验未点名 file-service RPC 机制 | REQ-NP-6 仍写「verify exists, confirmed, belongs to user」，未点名 GetFileUrl/GetFileInfo（FileInfo 含 user_id） | 🟡 SHOULD（沿用） |
| 4 | GetNotice community_id 缺省行为 | REQ-NR-2 已显式：缺失/空→080005，且给独立场景 | ✅ 已修复 |

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|-----------|------|---------|
| 1 | specs/attachment-security/spec.md REQ-AS-3（+ REQ-AS-1 禁止清单） | **magic-bytes 容器识别映射过宽，白名单/禁止清单被绕过（CRITICAL 安全漏洞）**。REQ-AS-3 写「.doc 按 OLE2/CFB 魔数 D0 CF 11 E0 A1 B1 1A E1 识别并映射 doc」「.docx 按 ZIP PK 头 + OOXML [Content_Types].xml 识别」。但：① OLE2/CFB 魔数**并非 Word 独有**，Windows Installer(.msi)、Excel(.xls)、PowerPoint(.ppt) 同为 OLE2 容器——`.msi` 在禁止清单中显式列为 exe/bat/.../msi 之一，**攻击者把 .msi 改名 .doc 上传，魔数判定通过→被当 doc 放行**，安全第二层形同虚设；.xls/.ppt 亦被当 doc 放行（不在白名单）。② docx 判定「ZIP + [Content_Types].xml」同样过宽：.xlsx/.pptx 也含 [Content_Types].xml，改名 .docx 放行。这与本 capability 的成立理由（封堵上传可执行文件/越权类型）直接冲突。 | 魔数判定必须落到**类型独有特征**：.doc 需在 OLE2 内检出 `WordDocument` 流（Word 专属）而非仅 CFB 头；.docx 需在 ZIP 内检出 `word/document.xml` 部件 + [Content_Types].xml（非仅 [Content_Types].xml，排除 xlsx/pptx）；其他 OLE2 子类型（msi/xls/ppt）与 OOXML 子类型（xlsx/pptx）→ 070004。补验收场景「.msi 改名 .doc 上传 → ConfirmUpload 拦截 070004」「.xlsx 改名 .docx → 拦截」与「真实 .doc 含 WordDocument 流 → 放行」。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|-----------|------|------|
| 1 | specs/notice-publish/spec.md REQ-NP-3 | `repeated int64 community_ids`（字段 8）未显式声明 `[jstype=JS_STRING]`（division_id 已声明，community_ids 未提）；社区 ID 为 Snowflake 大整数，TS 端无 jstype 会精度丢失，违反硬性约束 #3 | REQ-NP-3 对 community_ids 显式注明 `[jstype=JS_STRING]` |
| 2 | specs/notice-publish/spec.md REQ-NP-6 | 附件「存在+已确认+归属本人」校验仍未点名 file-service 调用机制（GetFileUrl(file_id)→FileInfo.user_id 可支撑），实现无唯一解释 | 明确经 GetFileUrl/GetFileInfo 取 user_id 校验归属，并定义重复绑定/已被其他实体引用的处置 |
| 3 | specs/notice-publish/spec.md REQ-NP-5 + proposal 影响范围 | DeleteNotice 由「现 CheckPublishScope 数据范围判定」（deletenoticelogic.go，辖区内有权限者可删）收窄为「仅发布者本人」，属**行为回归**；proposal 已登记 421 回收回归但未登记删除权收窄（community_admin/property_admin 失去辖区内删除他人通知的能力） | 在 proposal 变更说明中登记该行为收窄，与 REQ-PP-4 回归登记保持一致 |
| 4 | specs/notice-publish/spec.md REQ-NP-3 vs REQ-NP-4 | 未知目标节点→080006 与 division 展开空→080005 分属数据权限/参数无效两族，前端错误处理契约不明确（v2 沿用） | spec 写明 080005/080006 的区分原则供前端统一处理 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | REQ-NP-3/REQ-NP-4 展开契约建议显式绑定 GetResidentialAreasByDivision 的 `status=1`（approved only）参数，避免实现时漏传导致「未审核小区也进快照」（现 spec 仅以行为契约「仅含审核通过」表述，未点名 RPC status 参数） |
| 2 | GetPublishPermission（REQ-PP-1）「0=永久」经 GetUserRoles RPC 边界验证正确（DB NULL→0），建议在 spec 加一句「判定基于 RPC 输出 expires_at==0 OR > now」，防实现者误对 DB 写 SQL |
| 3 | REQ-AS-4「既有 entity_type 上传不回归」依赖「现网无 >10MB 或非白名单类型」事实，建议实现前逐 entity_type（avatar/verification/lostfound/contacts/notice）核对现有允许类型与大小（v2 沿用） |

## 问题跟踪表

| 编号 | 问题 | 状态 |
|------|------|------|
| 1 | OLE2/CFB + OOXML 魔数映射过宽（.msi/.xls/.ppt → doc、.xlsx/.pptx → docx 绕过） | 待修复（CRITICAL） |
| 2 | community_ids 缺 jstype=JS_STRING | 待修复（SHOULD） |
| 3 | 附件绑定校验 RPC 机制未点名 | 待修复（SHOULD） |
| 4 | DeleteNotice 行为收窄未登记回归 | 待修复（SHOULD） |
| 5 | 080005/080006 前端错误契约未显式区分 | 待修复（SHOULD） |
| 6-8 | 见 INFO | 建议采纳 |

---

VERDICT: **REVISION**（存在 1 项 CRITICAL 级 MUST FIX —— 附件白名单魔数判定被 .msi 等禁止类型绕过，一票否决）
---
