# Code Review — 社区枢纽服务（设计业务视角）

**审查时间**: 2026-08-16 22:00
**审查范围**: 工作树未提交改动 + 未跟踪文件（xss-sanitization：internal/sanitize 净化器 + Create/Update(submit) 写路径接入 + bluemonday 依赖）
**审查维度**: 设计一致性(#2)、代码质量(#4)、Migration(#8部分)

## 摘要
- 🔴 CRITICAL: 0 / 🟡 WARNING: 1 / 🔵 NOTE: 3

## 发现

### 🔴 CRITICAL
无。

### 🟡 WARNING
| # | 文件:行号 | 维度 | 问题 | 修复建议 |
|---|----------|------|------|---------|
| 1 | `rpc/internal/logic/notice/updatecontentpostlogic.go:157`（配合 :139、:146） | 设计一致性 #2 / 代码质量 #4 | **submit 发布分支 Kafka 推送使用事务前旧正文（未净化）**：`post` 在请求开头经 `FindOne` 载入（含存量 draft 未净化正文），submit 分支只把 `sanitizedText` 写入 DB（:142 `UpdateContentTx`），随后 :157 `KafkaProducer.Push(l.ctx, post, atts)` 仍携带 `post.Text` 原始值。`buildMessage`（producer.go:145）`Text: post.Text` 直接把未净化 HTML 原样放入 content-review 消息。REQ-XSS-6 本要关闭「净化前存量草稿经 submit 发布」缺口，DB 侧已封死，但**正是这条路径（存量 draft 带恶意 HTML 提交）会把原始攻击载荷经 Kafka 转发给 moderation-service 消费者**——事件 payload 与落库后状态不一致，重新打开一条传播通道。DB 读路径安全（主防线未破），故非 CRITICAL；但属净化目标场景内的真实漏洞。 | 在事务成功提交后、Push 之前，将净化后正文回写内存对象：`post.Text = sanitizedText`（提交失败/幂等无需改写时不回写）。并在测试中断言**推送 payload 的 text**（而非仅断言 Push 被调用）。 |
### 🔵 NOTE
| # | 文件:行号 | 建议 |
|---|----------|------|
| 1 | `rpc/internal/logic/notice/notice_helpers_test.go:248-257`（fakePusher）+ `updatecontentpostlogic_test.go` submit 用例 | **测试盲区正是 WARNING #1 漏网的原因**：`fakePusher.Push` 仅记录 `post.Id`，submit 用例只断言 `pusher.pushed == []int64{1}`，未断言消息内 text 是否净化。建议 fakePusher 捕获完整 `ContentReviewMessage`/`post.Text`，并对「存量 draft 恶意正文经 submit」补一条 payload 断言。 |
| 2 | `docs/design.md`（安全机制 / 业务流程·通知发布） | design.md 未同步 XSS 净化决策（D4 净化器归属、D7 净化后空串落库唯一化、D9 submit 前净化存量 draft）。另有一处既有陈旧：设计文档「通知发布」流程仍写「写入 notices 表」（早于 content_posts 重构遗留），建议随本变更一并校正。 |
| 3 | `internal/sanitize/sanitize_test.go` + CHANGELOG（REQ-XSS-8 交叉核对） | 白名单冻结仅抽样 5 行存量正文且全部为纯文本，未用后台编辑器实际 HTML 输出校验（如 `<table>`/`<h2 style>`）。已按 spec 程序满足并记录，但覆盖面偏窄，建议后续拿到编辑器真实产出补一轮比对，避免「存量合法富文本一经编辑重存即降级」。 |

## 设计一致性检查（#2）
- [x] 变更与 spec（xss-sanitization REQ-XSS-1/2/3/6）一致：3 个写入口（Create / Update 内容编辑 / Update submit）均已接入净化，与 spec 枚举一致（`createcontentpostlogic.go:137`、`updatecontentpostlogic.go:184`、`updatecontentpostlogic.go:139`）
- [x] 净化与非空校验顺序符合 D7：非空校验（080005）以原始正文先行，净化在落库前执行，净化后为空接受空串落库（不回判）——与 REQ-XSS-1/D7 一致
- [x] 幂等符合 REQ-XSS-3：submit 净化前后一致不二次改写，测试覆盖
- [x] 净化范围符合 REQ-XSS-5：仅正文 text，title 不净化（`{{ }}` 插值），wire 契约零变更
- [x] 数据模型正确：本次无 schema 变更，落库列 `text` 语义不变
- ⚠️ 事件 payload 与落库状态不一致：见 WARNING #1（submit 推送旧正文）

## 代码质量检查（#4）
- [x] 空指针防护：`post.PublisherId` nil 判断完备；`in.Text` presence 判断正确
- [x] 单例构建 `sync.Once` 并发安全；Sanitize 纯函数无共享可变状态
- [x] 幂等正则归一化经 Go html `escape()` 验证：文本节点 `"` 会转义为 `&#34;`，`normalizeAnchorRel` 的 `rel="..."` 字面匹配仅命中 bluemonday 生成的 rel 属性，不会改写纯文本内容（已核实，非误报）
- [x] 事务原子性：submit 分支净化写 + 置公开在同一事务；编辑分支 all-or-nothing
- ⚠️ 内存对象陈旧复用：submit 分支 Push 前未回写 `post.Text`（见 WARNING #1）

## Migration 安全性检查（#8 部分）
- [x] 本次变更无 DDL/Migration 变更（净化为纯应用层），不涉及回滚方案/锁表/现有数据影响
- [x] 存量已发布恶意正文不回填属 spec 明确接受（REQ-XSS-6/D5，out_of_scope），无 schema 演进风险
- [x] CHANGELOG 已更新（2026-08-16 公告正文存储型 XSS 净化条目）；依赖 bluemonday v1.0.27 锁定版本 + go.sum

---
VERDICT: PASS
---

（WARNING #1 建议在本轮或紧邻轮修复：一行回写 `post.Text = sanitizedText` + 一条 payload 断言。不阻塞提交，但请纳入技术债/回归跟踪。）
