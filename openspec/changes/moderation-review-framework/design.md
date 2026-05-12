## Context

社区平台需要 UGC 内容审核能力。当前 masterdata 服务已有敏感词 CRUD 管理（`md_sensitive_word` 表 + 完整审批流 API），但缺少运行时审核引擎。

参考开源项目 sensitive-word-master（Java，DFA 算法），分析其优缺点后，决定：
- 用 AC 自动机替代 DFA（单趟 O(n) vs 多趟扫描）
- 新增拼音谐音检测、拆字干扰检测
- 采用三层审核架构（AC 引擎 → 本地小模型 → 远端大模型）

moderation 服务已有基础框架（go-zero，端口 8890，health check），需在此基础上构建完整审核引擎。

## Goals / Non-Goals

**Goals:**
- 构建可运行的文本审核引擎（AC 自动机 + 归一化 + 谐音 + 拆字 + 白名单）
- 预留 LLM 模型调用接口（小模型/大模型），当前空实现
- 预留图片审核接口（感知哈希 + 多模态模型）
- 文本审核端到端可用：AC 层能拦截已知敏感词
- 词库从 masterdata_db 直读，支持定时同步

**Non-Goals:**
- 不实现 LLM 模型调用逻辑（Ollama/远端 API 的实际调用留后续）
- 不实现图片审核的实际模型调用
- 不实现人工复审工作台（仅标记 need_review，前端待开发）
- 不修改 masterdata 服务的敏感词 CRUD API（仅扩展表结构）

## Decisions

### D1: AC 自动机选型 — cloudflare/ahocorasick

**选择**: `github.com/cloudflare/ahocorasick`
**理由**: Cloudflare WAF 生产级使用，API 简洁（`NewMatcher(dict)` → `Match(text)`），无需 CGO，单趟 O(n) 扫描。
**替代方案**: `anknown/ahocorasick`（性能更高但社区小）、手写 Trie（开发成本高）。

### D2: 拼音转换 — mozillazg/go-pinyin

**选择**: `github.com/mozillazg/go-pinyin`
**理由**: Go 生态最成熟的拼音库，支持多音字、声调，纯 Go 实现无 CGO。
**约束**: 谐音扩展仅对 severity=1（高风险）词启用，每词最多 20 变体，防止组合爆炸。

### D3: 繁简转换 — 内置映射表

**选择**: 内置 `map[rune]rune` 映射表，覆盖 3500+ 常用汉字。
**理由**: 无外部依赖，启动快。覆盖社区 UGC 95%+ 繁体字场景。敏感词库本身用简体存储，繁体输入归一化后即可匹配。

### D4: 词库直读 masterdata_db

**选择**: moderation 服务直连 masterdata_db 读取 `md_sensitive_word` 表。
**理由**: 避免跨服务 RPC 调用延迟；词库数据变更频率低（分钟级），直连足够。
**约束**: moderation 服务配置中 DataSource 指向 masterdata_db；审核日志写入独立的 moderation_db。

### D5: 白名单策略 — 独立 AC 自动机

**选择**: 黑名单和白名单各维护一棵 AC 自动机，匹配时比较长度。
**理由**: 与 sensitive-word 的白名单策略一致，"长白黑短"时白名单优先跳过。独立树避免黑名单词被白名单词干扰。

### D6: LLM 调用 — 预留接口，空实现

**选择**: 定义 `LLMClient` 接口（`CheckText` + `CheckImage`），实现 `OllamaClient` 和 `RemoteLLMClient` 空结构体，方法返回 `ErrNotImplemented`。
**理由**: 引擎编排层完整实现三层逻辑，但 LLM 层后续填充，当前 AC 层可独立工作。

### D7: severity 语义调整

**选择**: 将 severity 从 `1=Low,2=Medium,3=High` 改为 `1=High,2=Medium,3=Low`。
**理由**: 业界通用惯例是 1=最高级别（P0/P1），与 action（1=Warn,2=Block,3=Review）逻辑对齐更直观。

## Risks / Trade-offs

**[词库同步延迟]** → moderation 与 masterdata 共享同一数据库，但内存中的 AC 自动机有同步间隔（默认 5 分钟）。极端情况下新词不会立即生效。缓解：提供手动Reload的管理接口。

**[谐音扩展误杀]** → 自动生成的谐音变体可能匹配正常文本。缓解：仅对 severity=1 的高风险词启用谐音扩展；置信度阈值可配置。

**[masterdata 表结构变更]** → 新增列需要 masterdata 服务重新生成 model 并更新 API。缓解：新增列均有默认值，不影响已有数据；masterdata 的 CRUD 逻辑只需小幅适配。

**[AC 自动机内存占用]** → 10 万级词库 + 谐音变体约 20 万词，AC 自动机约占用 100-200MB 内存。缓解：Go 的 GC 可回收旧树；对于社区平台规模足够。
