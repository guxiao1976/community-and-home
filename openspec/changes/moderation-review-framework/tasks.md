## 1. 数据库变更

- [x] 1.1 编写 ALTER TABLE SQL 修改 masterdata_db.md_sensitive_word 表：新增 word_type、pinyin_expanded、submission_status、submission_type、change_snapshot、submitter_id、submit_time、reviewer_id、review_time、review_notes、delete_time 列
- [x] 1.2 在 masterdata_db 中执行 ALTER TABLE 语句
- [x] 1.3 编写 DDL 创建 moderation_db 数据库及 mod_audit_log 表
- [x] 1.4 执行 DDL 创建 moderation_db 和 mod_audit_log 表

## 2. Masterdata 服务适配

- [x] 2.1 重新生成 masterdata 的 mdSensitiveWordModel_gen.go（goctl 生成或手动更新，匹配新增列）
- [x] 2.2 更新 masterdata.api 中 SensitiveWord 类型定义，新增 WordType、PinyinExpanded 字段
- [x] 2.3 更新 CreateSensitiveWordReq 加入 word_type 字段（默认1=黑名单）
- [x] 2.4 编译 masterdata 服务验证通过

## 3. Moderation 服务基础设施

- [x] 3.1 创建 internal 目录结构：engine/、ac/、normalize/、pinyin/、splitword/、whitelist/、llm/、imagehash/、wordstore/
- [x] 3.2 添加 Go 依赖：cloudflare/ahocorasick、mozillazg/go-pinyin、corona10/goimagehash、disintegration/imaging
- [x] 3.3 更新 api/etc/moderation-api.yaml 配置文件，新增 ACEngine、Normalizer、SmallModel、LargeModel、ImageHash 配置块
- [x] 3.4 更新 api/internal/config/config.go 配置结构体，对应新配置项
- [x] 3.5 新建 model/mod_audit_log_gen.go（审核日志表 generated model）
- [x] 3.6 新建 model/mod_audit_log_model.go（审核日志 custom model）

## 4. 文本归一化模块（internal/normalize/）

- [x] 4.1 实现 normalizer.go：Normalizer 结构体、链式组合 NormalizeChar/Normalize 方法、位置映射
- [x] 4.2 实现 width.go：全角半角转换（Unicode 范围 0xFF01-0xFF5E）
- [x] 4.3 实现 case.go：大写转小写
- [x] 4.4 实现 chinese.go：内置繁简映射表（350+汉字），繁体→简体转换
- [x] 4.5 实现 num.go：数字格式归一化（①→1、壹→1、全角１→1 等映射表）
- [x] 4.6 实现 english.go：英文特殊字符归一化（Ⓐ→A、ⓐ→a 等映射表）
- [x] 4.7 实现 repeat.go：连续重复字符压缩（3个及以上重复→单个）
- [x] 4.8 编写 normalize 单元测试：验证各策略输入输出、链式组合、位置映射

## 5. AC 自动机模块（internal/ac/）

- [x] 5.1 实现 ac.go：ACMachine 结构体（封装 ahocorasick.Machine）、读写锁并发保护
- [x] 5.2 实现 builder.go：Build(entries) 方法、Rebuild() 原子替换、拼音变体集成（调用 pinyin 模块）
- [x] 5.3 实现 matcher.go：MatchResult 结构体、Match(text) 单趟匹配方法
- [x] 5.4 编写 ac 单元测试：构建词库→匹配文本→验证结果、并发安全测试

## 6. 拼音谐音模块（internal/pinyin/）

- [x] 6.1 实现 pinyin.go：ToPinyin(hans) 方法，封装 mozillazg/go-pinyin，支持多音字
- [x] 6.2 实现 homophone.go：内置同音字映射表、ExpandHomophones(word, maxVariants) 谐音变体生成
- [x] 6.3 编写 pinyin 单元测试：汉字转拼音、多音字、谐音扩展、数量限制

## 7. 拆字检测模块（internal/splitword/）

- [x] 7.1 实现 detector.go：SplitDetector 结构体、配置化干扰字符集、Detect(text, acMachine) 方法
- [x] 7.2 实现干扰字符去除+片段拼接+AC 匹配逻辑
- [x] 7.3 编写 splitword 单元测试：单字符干扰、多字符干扰、正常文本不受影响

## 8. 白名单模块（internal/whitelist/）

- [x] 8.1 实现 whitelist.go：Whitelist 结构体（独立 AC 自动机）、LongestMatch(text) 方法
- [x] 8.2 编写 whitelist 单元测试：加载白名单、长白黑短跳过、仅有黑名单正常拦截

## 9. LLM 客户端预留（internal/llm/）

- [x] 9.1 实现 client.go：LLMClient 接口定义、CheckResult 结构体、ErrNotImplemented
- [x] 9.2 实现 ollama.go：OllamaClient 结构体（空实现，返回 ErrNotImplemented）
- [x] 9.3 实现 remote.go：RemoteLLMClient 结构体（空实现，返回 ErrNotImplemented）

## 10. 图片哈希预留（internal/imagehash/）

- [x] 10.1 实现 hash.go：ImageHasher 结构体、Hash/Distance/LoadViolationHashes 方法签名（基础实现）

## 11. 词库管理（internal/wordstore/）

- [x] 11.1 实现 store.go：WordStore 结构体、Load() 首次加载、Reload() 增量更新、版本号机制
- [x] 11.2 实现 sync.go：StartSync(ctx) 定时同步（默认 300 秒检查间隔）
- [x] 11.3 编写 wordstore 单元测试：加载逻辑、版本号对比

## 12. 审核引擎（internal/engine/）

- [x] 12.1 实现 engine.go：ModerationResult 结构体、MatchDetail 结构体
- [x] 12.2 实现 text_engine.go：TextEngine.Check() 文本审核三层编排（归一化→AC→白名单→拆字→模型降级）
- [x] 12.3 实现 image_engine.go：ImageEngine.Check() 图片审核编排（哈希→模型降级）
- [x] 12.4 编写 engine 单元测试：AC 直接拦截、AC 放行、模型降级放行+标记复审

## 13. API 接口层

- [x] 13.1 更新 api/moderation.api：定义 TextCheckReq/Resp、ImageCheck 接口类型（文本+图片）
- [x] 13.2 更新 api/internal/types/types.go：对应新 API 类型
- [x] 13.3 实现 text_review handler + logic：调用 TextEngine.Check()，记录审核日志
- [x] 13.4 实现 image_review handler + logic：接收文件上传，调用 ImageEngine.Check()，记录审核日志
- [x] 13.5 更新 handler/routes.go：注册新路由（POST /text/check、POST /image/check），JWT 鉴权

## 14. 服务上下文初始化

- [x] 14.1 更新 api/internal/svc/service_context.go：初始化 Normalizer、ACMachine、Whitelist、WordStore、TextEngine、ImageEngine
- [x] 14.2 在服务启动时调用 WordStore.Load() 加载词库
- [x] 14.3 启动 WordStore.StartSync() 定时同步 goroutine

## 15. 编译验证

- [x] 15.1 执行 go mod tidy + go build ./api/ 编译通过
- [ ] 15.2 验证 GET /api/moderation/health 返回 ok
- [ ] 15.3 验证 POST /api/moderation/text/check AC 层能拦截已知敏感词
- [ ] 15.4 验证 POST /api/moderation/text/check 正常文本通过
