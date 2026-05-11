## Why

社区平台需要一个内容审核微服务，对用户发布的文本和图片进行合规检查。当前仅有 masterdata 服务中的敏感词 CRUD 管理（纯黑名单存储），缺少运行时审核引擎。本项目基于 sensitive-word-master 项目的技术分析，采用三层审核架构（AC 自动机 → 本地小模型 → 远端大模型），并改进其 DFA 算法为 AC 自动机、新增拼音谐音检测和拆字干扰检测能力。

## What Changes

- 在 `services/moderation/` 已有框架基础上，实现完整的文本+图片内容审核引擎
- **修改 masterdata 服务**: 扩展 `md_sensitive_word` 表结构（新增 word_type 黑白名单区分、pinyin_expanded 谐音扩展标记、补齐 submission 工作流字段和 delete_time 软删除）
- **新增 moderation_db**: 创建 `mod_audit_log` 审核日志表
- **文本审核**: AC 自动机匹配（单趟 O(n)）+ 文本归一化链（全角半角/大小写/繁简/数字/英文/重复）+ 拼音谐音扩展 + 拆字干扰检测 + 白名单优先机制
- **图片审核**: 感知哈希（pHash）违规图库比对 + 多模态模型预留接口
- **模型调用**: 预留 LLM 接口（Ollama 本地小模型 + 远端大模型），当前为空实现
- **对外接口**: `POST /api/moderation/text/check`、`POST /api/moderation/image/check`

## Capabilities

### New Capabilities
- `ac-engine`: AC 自动机敏感词匹配引擎，替代 DFA 实现单趟 O(n) 扫描，支持词库动态加载与拼音谐音扩展
- `text-normalize`: 文本归一化链，包含全角半角、大小写、繁简转换（内置映射表）、数字格式、英文特殊字符、重复字符压缩
- `pinyin-homophone`: 拼音转换和谐音变体扩展，基于 go-pinyin 实现，每词最多 20 个变体
- `split-word-detect`: 拆字/插入干扰检测，识别"敏x感x词"类变体并还原匹配
- `whitelist`: 白名单 AC 自动机，白名单匹配长度 >= 黑名单时跳过（解决"长白黑短"问题）
- `llm-client`: LLM 调用预留接口，定义 OllamaClient + RemoteLLMClient（空实现待填充）
- `image-hash`: 图片感知哈希（pHash）+ 违规图库比对（预留接口）
- `review-engine`: 文本/图片审核引擎编排，串联 AC 引擎 → 小模型 → 大模型三层流程
- `word-store`: 词库管理，从 masterdata_db 直读敏感词表，定时同步 + AC 自动机增量重建
- `moderation-api`: 文本审核和图片审核 HTTP 接口，审核日志记录

### Modified Capabilities
- `sensitive-word-schema`: 扩展 md_sensitive_word 表结构，新增 word_type（黑白名单）、pinyin_expanded、补齐 submission 工作流字段和 delete_time

## Impact

- **数据库**: 修改 masterdata_db 的 md_sensitive_word 表（新增列），新建 moderation_db 及 mod_audit_log 表
- **服务**: masterdata 服务需要重新生成 model 和更新 API 类型以适配表结构变更；moderation 服务直读 masterdata_db 的敏感词表
- **依赖**: 新增 Go 依赖 cloudflare/ahocorasick、mozillazg/go-pinyin、goimagehash、imaging
- **配置**: moderation-api.yaml 新增 ACEngine、Normalizer、SmallModel、LargeModel、ImageHash 配置块
- **基础设施**: 可能需要新建 moderation_db 数据库实例
