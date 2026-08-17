# QA 机械化检查点设计文档

> 本文档是 QA 机械化检查点的**人读解释**。检查项的**权威清单**以脚本 `--list-checks` 机器生成输出为准，本文档不复制清单、只做「查什么 / 怎么查 / 为什么 / 抓不到什么」的四段式解释。
>
> 权威清单命令：
> ```bash
> .harness/skills/qa/scripts/harness-checks.sh --list-checks            # Go 服务
> .harness/skills/qa/scripts/harness-checks-frontend.sh --list-checks   # 前端服务
> ```
> 若本文档与 `--list-checks` 输出不一致，以 `--list-checks` 为准（脚本函数定义是唯一事实源）。

---

## 一、Go 服务检查（21 项）

### 分层视角

```
1-3   编译/测试层     "能不能跑、有没有坏味道、测试真不真"
4-5   ID 序列化层     "Snowflake ID 全链路无损"（本项目核心）
6     服务边界层      "微服务隔离"（架构）
7-8   规范安全层      "错误码格式 + 无硬编码密钥"
9-11  同步层          "文档/契约不漂移"
12-13 完整性层        "没有半成品、响应格式统一"
14-15 质量层(非阻塞)  "性能不倒退、路由不404"
16-18 治理层          "记忆/版本/测试有效性健康"
```

### 1️⃣ go build — 编译门禁
- **查什么**：`./...` 全部包能否编译。
- **怎么查**：`go build ./...`，exit≠0 即 FAIL。
- **为什么**：最低门槛，编不过一切白搭。
- **抓不到**：逻辑错误、运行时 panic、竞态、安全漏洞。**编译通过 ≠ 代码正确**。

### 2️⃣ go vet — 静态分析（可疑模式）
- **查什么**：常见坏味道——`printf` 格式串不匹配、`copylocks`（复制 Mutex）、`lostcancel`（context cancel 泄漏）、`unreachable`、`structtag`、`unusedresult`、`stringintconv`、`loopclosure` 等。
- **怎么查**：`go vet ./...`，基于 `go/analysis` 标准检查器。
- **为什么**：抓"能编译但明显写错"的代码。
- **抓不到**：业务逻辑、安全漏洞、依赖问题。

### 3️⃣ go test — 单元测试 + 0/0 假通过检测
- **查什么**：跑全部单测；**0/0 检测**（输出 "0 tests" 判 FAIL，防测试文件存在但没断言）；**新包检测**（新增包无测试函数判 FAIL）。
- **为什么**：单测是质量基础，0/0 假通过是常见坑。

### 3.5️⃣ gofmt — 变更文件格式
- **查什么**：变更的 Go 文件是否已 `gofmt` 格式化。
- **为什么**：对齐 pre-commit 提交门，统一格式。

### 4️⃣ Proto int64 jstype — ID 精度
- **查什么**：api-proto 里所有 `int64` 的 ID 字段（`id`/`*_id`）是否带 `[jstype = JS_STRING]`。
- **为什么**：**Snowflake ID 是 19 位整数，超过 JS `Number.MAX_SAFE_INTEGER`（2^53≈16位）**，不加 jstype 前端 JSON 解析会精度丢失。
- **抓不到**：字段号冲突、语义错误。

### 5️⃣ Go json:",string" — 后端 JSON 序列化
- **查什么**：Go 结构体 int64 ID 字段 JSON tag 是否带 `json:",string"`。
- **怎么查**：AST 检查器（go-ast-checker）+ 正则兜底。
- **为什么**：与 proto jstype 配套，后端响应给前端时 int64 ID 必须序列化为字符串。

### 6️⃣ 跨服务 DB import — 服务隔离
- **查什么**：一个服务是否 import 了**其他服务的 `model` 或 `internal` 包**。
- **为什么**：**微服务必须通过 gRPC 通信，禁止直连其他服务数据库/内部包**（否则破坏边界、绕过权限）。

### 7️⃣ 错误码格式 — 5 位错误码
- **查什么**：错误码是否是 **5 位数字**。
- **为什么**：5 位错误码体系（`10xxx` 用户域 / `08xxxx` 社区域）是全局约定，前端/网关靠它区分错误域。
- **抓不到**：错误码重复/冲突（机械化只查格式，语义靠 review）。

### 8️⃣ 硬编码密钥 — 安全
- **查什么**：源码里硬编码的 API key、密码、token、私钥。
- **为什么**：密钥必须走 `.env`（configx.MustLoad），硬编码=泄露风险。
- **抓不到**：变量名绕过的密钥、弱密钥。

### 9️⃣ 知识图谱新鲜度
- **查什么**：`docs/graph-context.md`（Neo4j 生成）是否过期。
- **为什么**：graph-context 是编码上下文来源，过期会误导子 Agent。

### 🔟 CLAUDE.md 结构数据
- **查什么**：CLAUDE.md 是否含必需结构化字段（角色/规则/命令）。
- **为什么**：保证服务文档可被子 Agent 正确消费。

### 1️⃣1️⃣ Proto→TS 对齐
- **查什么**：proto 生成的 TS 类型与前端使用是否一致。
- **为什么**：proto 改了但 TS 没重新生成 → 前端类型过时，运行时错。
- **现状**：常是 WARN（TS 滞后是存量）。

### 1️⃣2️⃣ API 逻辑 TODO stub — 未实现检测
- **查什么**：新 API 的 handler/logic 是否有 `TODO`/占位（返回空/not implemented）。
- **为什么**：防"路由注册了但逻辑没实现"的半成品上线。

### 1️⃣3️⃣ 响应单层包装
- **查什么**：API 响应是否统一 `responsex.Response` 单层 `{code, msg, data}`。
- **为什么**：前端拦截器依赖统一格式解包；双层包装会让前端拿错层。

### 1️⃣4️⃣ Benchmark 回归（非阻塞）
- **查什么**：关键函数 benchmark vs 基线，性能是否显著回退。
- **为什么**：性能优化是渐进的，回退只 WARN。

### 1️⃣5️⃣ API 冒烟（非阻塞）
- **查什么**：新/改 REST 端点 curl 一次，确认**非 404**。
- **为什么**：只查路由存在，不查业务正确。

### 1️⃣6️⃣ 记忆索引新鲜度
- **查什么**：`.harness/knowledge/memory/MEMORY.md` + `.memory-index.json` 是否与记忆文件同步。
- **为什么**：索引过期 → 子 Agent 找不到经验 → 重复踩坑。

### 1️⃣6.5️⃣ 设计一致性
- **查什么**：model 列 vs 标准迁移源是否一致。
- **为什么**：数据模型漂移会破坏迁移闭环。

### 1️⃣7️⃣ Git 卫生
- **查什么**：子模块登记（.gitmodules）、孤儿 worktree、脏指针。
- **为什么**：gitlink 漂移破坏环境一致性。

### 1️⃣8️⃣ 变异测试（gomu）
- **查什么**：对代码做变异（`>`→`<`），看单测能否杀掉变异体，**分数=测试有效性**。
- **为什么**：单测全绿 ≠ 测得好；理想分数 ≥80%。
- **现状**：常显示 `?%`（未解析分数）→ WARN。

### 1️⃣9️⃣ pipeline evals
- **查什么**：管线自身回归语料库（防管线改动回归）。
- **为什么**：管线也是代码，需要回归保护。

---

## 二、前端服务检查（8 项）

### 分层视角

```
1-3  编译/类型/测试    "能不能编、类型对不对、测试真不真"
4    安全              "无硬编码密钥"
5-6  代码卫生          "无调试残留、少类型逃逸"
7    契约对齐          "字段名与后端一致"（WARN）
8    单位规范          "只用 rem"（mobile 硬规则）
```

### 1️⃣ type-check（vue-tsc）
- **查什么**：TS 类型一致，含 **.vue 模板里的类型**。
- **怎么查**：`vue-tsc --noEmit`，失败回退 `tsc`。
- **为什么**：前端一半 bug 是类型不匹配；vue-tsc 比 tsc 多查模板。
- **抓不到**：逻辑错、`any` 逃逸。

### 2️⃣ unit-test（vitest）
- **查什么**：跑 `vitest run`，**0/0 假通过检测**。
- **为什么**：前端逻辑函数必须有测试，0/0 是自欺。
- **抓不到**：未覆盖分支（覆盖率工具未装）。

### 3️⃣ production build（vite build）
- **查什么**：能否产出生产包。
- **为什么**：**dev 能跑 ≠ 生产能编**（树摇/分割/资源处理）。

### 4️⃣ hardcoded-secrets
- **查什么**：源码里硬编码密钥/token/密码。
- **为什么**：前端密钥随代码分发到浏览器，硬编码=泄露。

### 5️⃣ debug-artifacts
- **查什么**：非测试代码的 `console.log/debug/dir/trace` + `debugger`。
- **怎么查**：正则扫 src，排除 spec/test/node_modules，**放行 `console.error`**。
- **为什么**：调试残留进生产包=日志噪音+可能泄露数据。

### 6️⃣ type-safety（禁 as any）
- **查什么**：非测试代码 `as any` 数量（目标 ≤10）。
- **为什么**：`as any` = 类型逃逸，掩盖类型错误。
- **注意**：WARN 非 FAIL（存量允许）。

### 7️⃣ api-field-align（snake/camel 对齐）
- **查什么**：前端读取后端字段名是否与 JSON tag 对齐（`created_at` vs `createdAt`）。
- **为什么**：**后端 protojson 输出 snake_case，前端用 camelCase 读就取不到值（undefined）**。项目反复踩的坑。
- **注意**：WARN（存量 34 处）。

### 8️⃣ unit-standard（rem only，mobile 专属）
- **查什么**：`web/mobile` 源码是否还有 `rpx`/`px`。
- **怎么查**：正则扫，排除注释/根字号 `font-size:16px`/`env()/var()` 的 0px 兜底，命中即 FAIL。
- **为什么**：项目定的是 rem 单位体系（根字号 16px，rpx÷32/px÷16），§13 规范。

---

## 三、已知覆盖缺口（待补，按优先级）

| 优先级 | 缺口 | 说明 |
|:---:|------|------|
| P1 | XSS 面扫描 | v-html/rich-text 未机械化扫描（notice-detail XSS 靠 review 抓）——Semgrep 规则待接（第二批） |
| P1 | 前端 eslint 静态分析 | 前端无 ESLint 配置，仅 check_type_safety 查 `as any`——第二批接入 |
| P2 | 错误日志规范 | review 在查「日志含 userId/communityId」，但无机械化检查 |

## 四、已闭环（2026-08-17）

- **Go 依赖漏洞审计（govulncheck）**：`harness-checks.sh` 新增 check #20，扫 `./...` 的已知漏洞（含 stdlib/传递依赖），有漏洞 FAIL（exit 3）、未装/网络异常 WARN。注意：只覆盖 Go，前端 npm 依赖不在其内。
- **Go 依赖漏洞修复**：上线即扫出 14 个可达漏洞（9 stdlib + grpc/x-net/x-text/otel），随后全模块（8 服务 + common + api-proto + ai-model api/rpc）升级 **Go 1.25.13** + grpc v1.82.1 / x-net v0.56.0 / x-text v0.39.0 / otel v1.44.0（含 otlptracehttp v1.44.0），govulncheck 复验全模块 **0 漏洞**。模块图中 `golang.org/x/crypto` GO-2026-5932 **无修复版（Fixed=N/A）**，非可达调用链，接受风险持续跟踪。
- **前端覆盖率量化**：mobile 接入 `@vitest/coverage-v8`，vitest.config 设覆盖率阈值（Stmts 58/Branch 50/Funcs 55/Lines 58）；门禁侧 `harness-checks-frontend.sh` 的 unit_test 对装有 coverage provider 的项目跑 `vitest run --coverage`，覆盖率 < 阈值 FAIL。**pc 同轮接入 `@vitest/coverage-v8`**（2026-08-17）：实测基线 Stmts 4.2/Branch 3.2/Funcs 2.7/Lines 4.4（存量单测极少），阈值设「地板」3/2/2/3 防回退，随测试增长应同步上调（对齐 mobile 做法）。
  - 踩坑记录：coverage 曾报「Something removed coverage/*.tmp」——根因是 `unit-standard-gate.spec.ts` spawnSync 调用 harness-checks-frontend.sh，内层 vitest 再跑 --coverage 与当前进程冲突；已用 `HARNESS_RECURSE=1` 守卫（递归时不带 coverage）修复。
- **前端依赖漏洞审计（trivy）**：`harness-checks-frontend.sh` 新增 check #9，`trivy fs --scanners vuln --severity HIGH,CRITICAL --exit-code 1` 直接读 package-lock.json，HIGH/CRITICAL 即 FAIL、trivy 未装 WARN、DB 下载异常 WARN。CI `frontend-ci.yml` 两 job 均接入 `aquasecurity/trivy-action`。**不依赖 npm audit**（npmmirror 下不可用）。
- **Go 竞态检测（-race）**：`harness-checks.sh` check #3 go test 加 `-race`。本地热缓存约 +7s/服务可接受；CI 冷缓存首次约 +75s/服务。依赖 CGO（需 gcc，GitHub runner 默认有）。
- **golangci-lint 激活**：pre-commit hook 原引用但未安装、且 grep 模式匹配不到 golangci 输出（装了也永远"通过"）。现装 v1.64.8 + `.golangci.yml`（排除 SA5008 误报 go-zero `json:"x,optional"`），hook 改为**按变更模块跑 + 退出码判断**，新违规拦截（exit 1）、存量不阻塞。
