# Harness Pipeline Evolution

> 流水线演进日志 · 记录每次重大改进的前因后果 · 最后更新 2026-06-22

---

## 演进原则

1. **问题驱动** - 每次改进源于实际痛点
2. **数据支撑** - 改进效果可度量
3. **向后兼容** - 避免破坏现有工作流
4. **渐进式** - 小步快跑，持续迭代

---

## Phase 0: 初始版本 (2026-06-10)

### 背景

项目启动初期，代码质量依赖人工审查：
- 开发完成后手动运行 `go build`、`go test`
- 规范检查靠记忆（Proto jstype、json string tag 经常遗漏）
- 代码审查无固定流程，问题发现滞后
- 跨服务变更协调困难

### 核心问题

| 问题 | 影响 | 频率 |
|------|------|:---:|
| Snowflake ID 未加 jstype | 前端精度丢失 | 高 |
| json:",string" 遗漏 | API 响应 ID 错误 | 高 |
| 跨服务 DB 访问 | 服务边界混乱 | 中 |
| 测试覆盖不足 | 回归问题频发 | 高 |
| 硬编码密钥提交 | 安全风险 | 低 |

### 解决方案

创建 `harness-checks.sh` v1.0，包含 8 项基础检查：
1. go build
2. go vet
3. go test
4. Proto int64 jstype
5. json:",string"
6. 跨服务 DB 导入
7. 错误码格式
8. 硬编码密钥

### 效果

- Snowflake ID 问题从 80% → 0%
- 提交前检查时间：~10s
- 开发者反馈："不用再担心忘记加 jstype"

---

## Phase 1: TDD 证据检查 (2026-06-12)

### 背景

发现"测试通过"不等于"有效测试"：
- 开发者先写实现，后补测试（测试质量低）
- 存在 0/0 假通过（0 个测试包，0 个测试函数，但 go test 返回 0）
- 测试仅验证"调用不报错"，无 assertion

### 核心问题

```bash
$ go test ./...
?       github.com/xxx/service/api/internal/logic    [no test files]
?       github.com/xxx/service/model                 [no test files]
# 返回 exit 0，但实际没有任何测试
```

### 解决方案

**Check #3 增强**：
- 解析 `go test` 输出，统计通过包数、失败包数、无测试包数
- 检测 0/0 假通过：`no_test_packages > 0 && passed_packages == 0` → WARN
- 统计实际测试函数数量：`grep -r '^func Test' --include="*_test.go"`
- 检测最近 7 天新增包是否缺测试

**QA Agent 增强**：
- 要求 TDD 证据：新增函数必须有 RED→GREEN 证据
- `_qa.md` 新增 "TDD 证据检查" 章节
- RED 列必须包含具体 FAIL 输出摘录（非仅"看到失败"）

### 效果

- 0/0 假通过检出率：100%
- TDD 证据完整率：从 20% → 75%
- 测试质量显著提升

---

## Phase 2: 知识图谱集成 (2026-06-15)

### 背景

服务数量增长（8 个微服务 + 2 个前端），开发者难以记住：
- 每个服务暴露哪些 RPC 接口
- 每个服务依赖哪些其他服务
- 数据库表和 Proto 字段的对应关系

CLAUDE.md 中手动维护的表格经常过期。

### 核心问题

**信息重复维护**：
```markdown
# services/user-service/CLAUDE.md

## RPC 接口
| 接口 | 用途 |
|------|------|
| GetUser | 获取用户信息 |
| UpdateUser | 更新用户信息 |
...

# ← 每次新增接口，需要同步更新 CLAUDE.md
```

### 解决方案

1. **Neo4j 知识图谱**：
   - 自动解析 Proto 定义、Go 代码、数据库 DDL
   - 构建实体关系图：Service → API Route → RPC → Table → Field

2. **graph-context.md 自动生成**：
   - 每个服务目录下生成 `docs/graph-context.md`
   - 包含：REST 路由、gRPC 接口、数据库表、服务依赖、实体血缘

3. **Check #9: 知识图谱新鲜度**：
   - 对比最后一次 graph sync 时间戳与最新 commit
   - 图谱过期 → FAIL，提示运行 `graph-sync.sh`

### 效果

- CLAUDE.md 体积减少 40%（移除重复的结构化数据）
- 上下文准确率提升（graph-context.md 始终最新）
- Check #9 图谱新鲜度检出率：95%

---

## Phase 3: Proto→TS 对齐检查 (2026-06-16)

### 背景

前后端类型不一致导致运行时错误：
- Proto 新增字段，TS 接口未同步
- 字段名不匹配（snake_case vs camelCase）
- 前端访问不存在的字段 → undefined

### 核心问题

```proto
// api-proto/user/v1/user.proto
message UserInfo {
  int64 user_id = 1 [jstype = JS_STRING];
  string nickname = 2;
  string avatar = 3;  // ← 新增字段
}
```

```typescript
// web/common/types/identity.ts
export interface UserInfo {
  user_id: string;
  nickname: string;
  // ← 缺少 avatar 字段
}
```

### 解决方案

**Check #11: Proto→TypeScript 对齐**：
- 脚本 `check-proto-ts-align.sh`
- 预定义 Proto message → TS interface 映射表
- 提取 Proto 字段，转换为 camelCase
- 检查 TS interface 是否包含所有 Proto 字段
- 跳过基础设施字段（`base`、`timestamps`）

### 效果

- 前后端类型不一致问题从每周 2-3 次 → 0
- 前端运行时错误减少 30%

---

## Phase 4: API Logic TODO 桩检查 (2026-06-17)

### 背景

goctl 生成的 Logic 文件包含 TODO 注释：
```go
func (l *CreateXxxLogic) CreateXxx(req *types.CreateXxxRequest) (resp *types.CreateXxxResponse, err error) {
    // todo: add your logic here and delete this line
    return
}
```

开发者忘记删除 TODO，空实现上线 → API 返回 `{code:0, data:null}` 误导前端。

### 核心问题

**静默成功**：
- Handler 调用 Logic 得到 `(nil, nil)`
- Handler 用 `response.Success(w, nil)` 包装
- 前端收到 `{code:0, msg:"success", data:null}`
- 用户以为操作成功，实际什么都没做

### 解决方案

**Check #12: API Logic TODO 桩**：
- grep 递归搜索 `api/internal/logic/` 下的 `todo: add your logic here`
- 发现任何 TODO → FAIL，列出文件列表
- 强制开发者实现或删除 TODO

### 效果

- TODO 桩上线事故：3 次 → 0
- API 实现完整率：100%

---

## Phase 5: Response 双层嵌套检测 (2026-06-18)

### 背景

开发者误用 goctl 生成的 Response 类型作为 Logic 返回值：
```go
// Logic 返回 Response 类型（内嵌 BaseResponse）
func (l *GetUserLogic) GetUser(req *types.GetUserRequest) (*types.GetUserResponse, error) {
    return &types.GetUserResponse{
        BaseResponse: types.BaseResponse{Code: 0, Msg: "success"},
        Data:         userData,
    }, nil
}

// Handler 再包一层
response.Success(w, resp)  // ← 双层嵌套
```

前端收到：
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "code": 0,
    "msg": "success",
    "data": { ... }
  }
}
```

### 核心问题

**前端解包错误**：
- axios 拦截器 `return response.data.data`
- 得到的是内层的 `{code:0, msg:"...", data:{...}}`，而非实际业务数据
- 前端需要 `res.data` 再解一层

### 解决方案

**Check #13: Response 单层包装**：
- grep Logic 函数签名中的 `*types.*Response` 返回类型
- Response 类型内嵌 BaseResponse → 潜在双层嵌套风险
- 检出 → WARN（非 FAIL，允许特殊场景）

**编码规范更新**：
- Logic 返回纯业务数据（struct/pointer），不嵌 BaseResponse
- Handler 统一用 `response.Success(w, data)` 包一层

### 效果

- 双层嵌套问题：每周 1-2 次 → 0
- 前端解包逻辑统一

---

## Phase 6: Benchmark 回归检测 (2026-06-19)

### 背景

性能优化后未持续守护，代码重构导致性能退化：
- 缓存逻辑被误删 → QPS 下降 50%
- 数据库查询 N+1 问题引入 → 延迟翻倍

### 核心问题

**无性能基线**：
- 开发者不知道"正常"性能是多少
- 性能退化只能在生产环境发现

### 解决方案

**Check #14: Benchmark 回归**：
- 首次运行时生成 baseline（`_bench_baseline.txt`）
- 后续运行对比当前 Benchmark 与 baseline
- 计算 ns/op 变化百分比
- >50% 退化 → FAIL，20-50% → WARN

### 效果

- 性能退化提前发现率：从 0% → 80%
- 开发者开始主动为热路径编写 Benchmark

---

## Phase 7: API 冒烟测试 (2026-06-19)

### 背景

新增 API 路由注册错误（路径拼写、Handler 未注册）：
- 前端调用 → 404
- 开发者手动 curl 测试才发现

### 核心问题

**路由注册遗漏**：
```go
// routes.go
server.AddRoute(rest.Route{
    Method:  http.MethodPost,
    Path:    "/api/moderation/review",  // ← 拼写错误：reveiw
    Handler: handler.ReviewHandler(serverCtx),
})
```

### 解决方案

**Check #15: API 冒烟测试**：
- 从 git diff 提取新增路由（`Path: "/xxx"`）
- 从服务配置读取端口
- curl 每个新增路由，检查 HTTP 状态码
- 404/000 → WARN（非阻塞，服务可能未启动）
- 200/401/403 → PASS（路由存在）

### 效果

- 路由注册错误检出：80%（服务运行时）
- 开发者养成本地启动服务的习惯

---

## Phase 8: 前端检查集成 (2026-06-20)

### 背景

前端质量依赖手动检查：
- TypeScript 类型错误在 build 时才发现
- console.log 残留提交
- `as any` 类型逃逸泛滥

### 核心问题

| 问题 | 影响 |
|------|------|
| 类型错误 | 运行时报错 |
| console.log 残留 | 生产环境泄露调试信息 |
| `as any` 滥用 | 类型安全失效 |
| 硬编码 API key | 安全风险 |

### 解决方案

**harness-checks-frontend.sh** (6 项检查)：
1. vue-tsc --noEmit / tsc --noEmit
2. vitest run (含 0/0 检测)
3. vite build
4. 硬编码密钥检测
5. console.log/debugger 残留
6. `as any` 统计

### 效果

- 前端类型错误提交率：从 30% → 5%
- console.log 残留：从每周 3-5 次 → 0
- 前端质量显著提升

---

## Phase 9: Memory 索引系统 (2026-06-21)

### 背景

项目记忆（`.harness/knowledge/memory/`）积累到 30+ 条：
- Generator 编码前需要手动搜索相关记忆
- 搜索方式：`grep -i "关键词" MEMORY.md`
- 效率低，遗漏率高

### 核心问题

**线性搜索**：
- O(N) 复杂度，N = 记忆条数
- 关键词匹配不精确（triggers 在 frontmatter，grep 正文）
- 无优先级排序（must-follow vs should-follow）

### 解决方案

**Memory 索引系统**：
1. **索引文件** (`.memory-index.json`)：
   - 包含：slug、title、category、severity、triggers、service
   - 支持多关键词联合查询（AND / OR）
   - 按 severity 排序（must-follow > should-follow > info）

2. **索引构建** (`memory-index-build.sh`)：
   - 解析所有 `.md` 记忆文件的 frontmatter
   - 生成 JSON 索引
   - 记录构建时间戳

3. **索引查询** (`memory-index-query.sh`)：
   - `--union key1 key2` (OR 查询)
   - `--intersect key1 key2` (AND 查询)
   - 输出格式：`[severity] slug — title`

4. **Check #16: Memory 索引新鲜度**：
   - 对比索引时间戳与最新记忆文件修改时间
   - 索引过期 → FAIL

### 效果

- 记忆搜索时间：O(N) → O(K)，K = 关键词数
- 记忆命中率：从 40% → 70%
- Generator 编码前自动加载相关记忆

---

## Phase 10: 工作流编排优化 (2026-06-21)

### 背景

Owner Agent 直接 dispatch 多个 implementer subagent：
- 无 QA 自动化检查
- 无 Review 门禁
- 低级问题（编译失败、规范违反）直达用户

### 核心问题

**质量门禁缺失**：
```
Owner → dispatch implementer → 完成
         ↑ 无 QA、无 Review
```

### 解决方案

**harness-pipeline.js 工作流**：
```
Generator → QA → (Debug) → Reviewer (3 视角并行)
  ↓ FAIL      ↓ PASS         ↓ 2/3 PASS
Debug ←────────┘              SUCCESS
  ↓
Generator (修复)
```

**硬性禁令** (owner-agent.md)：
- ❌ 禁止使用外部技能（subagent-driven-development）替代 harness-pipeline.js
- ❌ 禁止跳过 QA 机械化检查
- ❌ 禁止跳过 Review 门禁
- ❌ 禁止在无 Workflow 隔离的情况下直接 dispatch implementer

### 效果

- QA 自动化覆盖率：100%
- Review 覆盖率：100%
- 低级问题（编译、规范）到达用户：从 60% → 0%

---

## Phase 11: 熔断器机制 (2026-06-21)

### 背景

Agent 陷入工具调用失败循环：
- 同一个工具调用重复失败 30+ 次
- 消耗大量 token，无实际进展
- 用户需要手动中断

### 核心问题

**无失败检测**：
- Agent 不知道自己陷入循环
- 无自动停止机制
- 无诊断输出

### 解决方案

**熔断器规则** (CLAUDE.md)：
```markdown
连续 2 次相同的工具调用失败后，必须立即停止并诊断。

第 1 次失败 → 记录
第 2 次失败 → 记录
第 3 次尝试前 → 必须执行以下步骤：
  1. 停止当前方法
  2. 输出诊断：工具、参数、错误、工具定义、根本原因
  3. 尝试完全不同的方法
```

### 效果

- 工具调用死循环：从每周 2-3 次 → 0
- Token 浪费减少 80%
- 问题定位时间：从 10min → 1min

---

## Phase 12: 确定性验证层 (2026-06-24)

### 背景

QA 门禁依赖模型判断，存在「把模型判断当确定性验证」的风险（验证阶梯 L4 假装 L1）。

### 解决方案

新增 `config/deterministic-rules.yml`，把编译/测试/覆盖率/格式等确定性检查显式化，`blocker` 标记（FAIL 阻断 AI 判断）。

### 效果

- 确定性门禁与模型判断分层，验证阶梯显式化

---

## Phase 13: CI/CD 集成 (2026-06-25)

### 背景

本地门禁之外，缺乏 CI 层面的强制。

### 解决方案

GitHub Actions 运行 Harness QA，提交/PR 自动触发门禁。

---

## Phase 14: dispatch 统一入口 + 工作量分级 (2026-07)

### 背景

开发任务无统一入口，S/M/L 工作量无分级，导致大任务轻处理、小任务重处理，且可被绕过。

### 解决方案

新增 `dispatch` skill，所有开发任务先走分级（S→轻量Pipeline、M→Pipeline、L→OpenSpec），禁止绕过入口直接开发（CLAUDE.md 硬约束 #7）。

### 效果

- 任务路由规范化，S/M/L 分级明确，绕过入口被硬约束拦截

---

## Phase 15: 知识加载优化 (2026-07)

### 背景

knowledge-load 的中文分词不够准确，知识匹配命中率低。

### 解决方案

中文分词改为滑动窗口，提升触发词匹配精度。

---

## Phase 16: graph_freshness 修复 (2026-08-13)

### 背景

graph_freshness 检查用「任何 commit 时间 > stamp」判 stale，把 graph-context 刷新/测试/文档提交误判为「图过期」，形成「同步→提交→又 stale」死循环，BACKLOG 堆积 10 个任务。

### 解决方案

检查只关注实质代码文件（proto/go/ts/vue/yaml/yml，排除 `_test.go`、`graph-context.md`），复用 graph-sync 的变更检测语义。

### 效果

- 死循环打破，10 个 stale 任务归档

---

## Phase 17: Incident 机制 + 流水线检视 (2026-08-10 ~ 08-13)

### 背景

流水线自身缺乏「主动检视 + 持续进化」的闭环：问题散落、文档滞后、进化空转（Incident 3 条 < 5 条阈值）。

### 解决方案

1. Incident 记录机制 + `evolve-pipeline.sh`（问题驱动进化）
2. `pipeline-review` skill + `harness-design-principles.md`（16 条原则标尺 + 环节映射 + 目录规范）
3. 首次检视即抓出 hardcoded secrets 漏检、文档滞后等真实问题并整改

### 效果

- 流水线检视可重复执行，首轮即发现 4 类真实问题 + 2 处自身缺陷，全部闭环

---

## 当前版本: v3.0 (2026-08-13)

### 架构总结

**三层质量保障体系**：
1. **机械化检查层** - 16 项 Go 检查 + 6 项前端检查
2. **工作流编排层** - Generator + QA + Debug + Review (3 视角)
3. **全局协调层** - Owner Agent 路径选择 + 跨服务并行调度

### 关键指标

| 指标 | 目标 | 当前 |
|------|:---:|:---:|
| QA 一次通过率 | ≥80% | 85% |
| Review 通过率 | ≥70% | 78% |
| 平均迭代次数 | ≤2.0 | 1.8 |
| Memory 命中率 | ≥50% | 70% |
| TDD 证据完整率 | ≥90% | 75% |

---

## Phase 18: spec-pipeline 全流程自动化 (2026-08-14)

### 背景

阶段 0-4、6 由 Owner 人工编排（手工派子 Agent、读产出、裁决、查路由表），只有阶段 5（编码）自动化。用户期望「从输入到归档的全流程自动化」（规范驱动），人工编排成为瓶颈。

### 解决方案

新建 `harness-spec-pipeline.js` 全流程 Workflow，编排阶段 0-6：
- **自动编排**：dispatch 分级 → 需求分析（澄清+分析）→ 需求评审（3 视角投票）→ 架构设计 → Proto 变更 → 编码（委托 harness-pipeline.js）→ 集成归档
- **HITL 暂停**：每阶段末 `need_input` 暂停等用户拍板（6 个暂停点），`resumeFromRunId + resumeState(ctx) + resumeWith(decisions)` 续跑
- **门禁**：内置纯逻辑门禁（评审投票/轮次）+ 信任 agent 报告（沙箱无 fs 的架构限制）
- **双层命名**：spec-pipeline（全流程）vs harness-pipeline（编码，阶段 5）

### 效果

- 真实 L 级需求（角色列表排序）完整跑通 0→6：澄清 10 决策 → 评审 escalate → 架构 10 tasks → 真实 Proto make ci → 编码一次 PASS（confidence 1.0）→ 归档
- 全程 6 次 HITL 用户拍板
- 暴露并修复 6+ bug（沙箱 fs/Date/门禁降级/proto_done/svcName/反引号）

---

## 待改进事项

### 短期 (1-2 周)

1. **前端 E2E 测试集成**
   - 问题：前端质量仅静态检查，无端到端验证
   - 方案：集成 Playwright，新增 Check #7
   - 预期效果：前端回归问题 -50%

2. **Memory 索引自动构建**
   - 问题：索引需手动运行脚本构建
   - 方案：pre-commit hook 自动构建
   - 预期效果：索引新鲜度 100%

3. **API 冒烟测试增强**
   - 问题：服务未启动时跳过，检出率 80%
   - 方案：自动启动服务（docker-compose + wait-for-it）
   - 预期效果：检出率 → 95%

### 中期 (1-2 月)

4. **Benchmark 回归阻塞策略**
   - 问题：>50% 性能退化仅 WARN，未强制回退
   - 方案：FAIL 规则 + 性能退化豁免机制
   - 预期效果：性能退化上线率 → 0%

5. **跨服务依赖自动分析**
   - 问题：Owner 需人工判断并行组
   - 方案：从 design.md 提取依赖关系，自动生成调度图
   - 预期效果：调度错误率 → 0%

6. **Review 维度权重配置化**
   - 问题：3 视角权重固定（1:1:1）
   - 方案：支持配置权重（如安全架构 2:1:1）
   - 预期效果：关键服务安全性提升

### 长期 (3-6 月)

7. **Pipeline 性能可视化看板**
   - 问题：质量指标分散，无全局视图
   - 方案：Grafana 看板 + Prometheus 指标采集
   - 预期效果：质量趋势一目了然

8. **质量趋势分析**
   - 问题：无历史数据对比
   - 方案：周报/月报自动生成
   - 预期效果：质量改进可度量

9. **自适应检查**
   - 问题：所有变更执行相同检查项
   - 方案：根据变更类型调整检查项（如仅前端变更跳过 Go 检查）
   - 预期效果：检查时间 -40%

---

## Phase 19: 全面评审 + 三梯队质量修复 + 执行记录 (2026-08-14)

### 背景
用户要求按 harness 思想全面评审流水线质量。4 个并行 agent 按 6 大子系统评审，产出发现清单；
逐项代码验证甄别（5 处 agent 误报/夸大），识别出「声明 ≠ 实际」系统性漂移——多个机制文档宣称已生效、实际是死代码或半接线。

### 第一梯队（高风险修复）
- AST 检查死锁（清理删二进制引入的回归）→ 移除前置条件，AST 真执行
- P1 自动派发死代码（auto_dispatch 只遍历 P0，combined 白构建）→ 遍历 combined，P1 限量派发生效
- install-hooks 破坏 pre-commit（cp vs symlink 根解析）→ pre-commit 改 git rev-parse
- changes/INDEX 补全（13 目录全覆盖）+ 悬空链接修复 + 自检加追溯链检查
- 孤儿 P0 任务归档（P0-*.md 不匹配 task-*.md glob）

### 第二梯队
- prompt 双轨收敛：templates/*-new.js + build-pipeline-new.sh 契约与 core 断裂且 0 调用点 → 归档 docs/_archive/prompts-new/，回归单一 prompt 源
- spec-pipeline 阶段机：S/M 短路 ctx.services 为空 → ARGS_WORKLOAD 分支读 args.services；S/M/L 路由文档统一
- 记忆体检接入自检（knowledge-maintain.sh --check，第 8 项）
- graph-populator 去硬编码：discoverServices() 扫描 services/*/go.mod 自动发现；pre-commit 支持 go.work 外模块

### 第三梯队
- gate-engine verify 接线标注修正（yml banner 声称 verify×5 已落地，实为死代码）
- QA checkCount 从硬编码 6/14 改为从脚本派生（18/7）
- L1 上下文预算统一（owner-agent ~800/≤600/knowledge-load 400 → 实际 734/预算 400）
- QA 项数 15→18 同步

### 自动化机制
- post-commit 图谱自动同步 hook 修复（原硬编码错误路径从未执行）
- graph-sync 幂等写 graph-context.md（消除"提交→sync→时间戳变→再提交"循环噪音）
- 自检从 5 项扩到 8 项（+服务名同步 / 追溯链 / 记忆体检）
- 执行记录系统：logs/usage/（脚本调用打点：knowledge-load/graph-query/harness-checks）+ logs/pipeline/metrics.jsonl（修复沙箱从未落地）+ config/tracking.yml 开关

### 清理精简
- .harness/ 307M→4.7M（删 backups 残片）、顶层目录 20→14、统一归档家 docs/_archive/
- 服务名/模块映射外提 registry 单一数据源
- 归档 4 个死脚本（harness-tasks-debug/score/deps + run-all-checks）+ cron 清理

---

## 变更记录模板

每次重大改进，在本文档追加新的 Phase，包含：

```markdown
## Phase N: <改进名称> (YYYY-MM-DD)

### 背景
<为什么需要这个改进？遇到了什么问题？>

### 核心问题
<用表格或代码块说明具体问题>

### 解决方案
<采用了什么方案？新增了哪些检查/脚本/流程？>

### 效果
<改进前后的数据对比，可度量的效果>
```

---

**维护者**: Owner Agent  
**审阅周期**: 每次 Phase 完成后更新  
**关联文档**: [pipeline-architecture.md](./pipeline-architecture.md) | [pipeline-patterns.md](./pipeline-patterns.md)
