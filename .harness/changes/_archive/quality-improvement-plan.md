# Harness 质量改进方案 — 符合驾驭工程理念

**创建时间**: 2026-06-23
**状态**: 待执行
**优先级**: P0
**预计周期**: 12 周

---

## 一、改进理念：Harness 核心原则

### 1.1 什么是 Harness？

```
Harness (驾驭) = 让 AI 团队像飞机驾驶舱一样精准控制开发过程

四大支柱：
├─ Mechanization   (机械化) — 能自动化的绝不手工
├─ Accountability  (可追溯) — 每个决策都有证据链
├─ Composition     (可组合) — Agent/Skill 按需编排
└─ Memory-Driven   (记忆驱动) — 错误只犯一次
```

### 1.2 当前问题的本质

**不是缺少工具，而是缺少"驾驭"**：

```
❌ 现状：有完整的仪表盘（Pipeline），但飞行员不看仪表
✅ 目标：建立强制检查清单（Checklist），不通过不起飞
```

**核心数据**：
- 测试覆盖率：7.5% vs 行业标准 70-80%
- API 层测试：0%（handler/logic 完全无测试）
- QA/Review 覆盖：25%（仅 2-3 个服务）
- TDD 执行率：存疑（缺少 RED 证据）

---

## 二、三阶段渐进式落地

### 阶段 0：修复当前系统（1周）— 让现有机制可用
### 阶段 1：建立最小闭环（4周）— 强制执行核心流程  
### 阶段 2：全面推广（12周）— 覆盖所有服务和场景

---

## 三、阶段 0：修复当前系统（P0，1周内）

### 任务清单

- [x] **Task 0.1**: 启用变更追踪机制
  - 状态：✅ 已完成
  - 产出：`.harness/changes/` 目录、INDEX.md、TEMPLATE/
  
- [ ] **Task 0.2**: 修复 harness-checks.sh 脚本
  - 问题：`check_frontend: command not found`
  - 问题：硬编码密钥检查失败
  - 详见：`.harness/tasks/P0-fix-harness-checks.md`
  
- [ ] **Task 0.3**: 建立门禁检查脚本
  - 创建：`.harness/scripts/harness-gate-check.sh`
  - 功能：阻断式验证阶段产出物完整性
  
- [ ] **Task 0.4**: TDD 证据强制验证
  - 修改：harness-pipeline.js QA 阶段
  - 要求：RED 列必须包含具体 FAIL 输出摘录

---

## 四、阶段 1：建立最小闭环（4周）

### 目标：强制执行核心质量流程

#### Week 1: API 层测试框架搭建

**目标**: 为 3 个核心服务建立 API 测试模板

```go
// 测试框架选择：httptest + gomock + testify
// 位置：services/<name>/api/internal/logic/<module>/*_test.go

func TestCreateUserLogic(t *testing.T) {
    tests := []struct {
        name       string
        req        *types.CreateUserReq
        mockSetup  func(*mock.MockUserModel)
        want       *model.User
        wantErr    bool
    }{
        {
            name: "success - 正常创建",
            req:  &types.CreateUserReq{Phone: "13800138000", ...},
            mockSetup: func(m *mock.MockUserModel) {
                m.EXPECT().Insert(gomock.Any(), gomock.Any()).
                    Return(int64(1), nil)
            },
            want: &model.User{Id: 1, ...},
            wantErr: false,
        },
        {
            name: "error - 手机号已存在",
            req:  &types.CreateUserReq{Phone: "13800138000"},
            mockSetup: func(m *mock.MockUserModel) {
                m.EXPECT().Insert(gomock.Any(), gomock.Any()).
                    Return(int64(0), sqlx.ErrNotFound)
            },
            wantErr: true,
        },
        // ... 边界/异常用例
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockUserModel := mock.NewMockUserModel(gomock.NewController(t))
            tt.mockSetup(mockUserModel)
            
            l := NewCreateUserLogic(context.Background(), &svc.ServiceContext{
                UserModel: mockUserModel,
            })
            
            got, err := l.CreateUser(tt.req)
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want.Id, got.Id)
            }
        })
    }
}
```

**交付物**:
- [ ] `services/user-service/api/internal/logic/user/createuserlogic_test.go`
- [ ] `services/auth-service/api/internal/logic/auth/loginlogic_test.go`
- [ ] `services/moderation-service/api/internal/logic/text_review/reviewtextlogic_test.go`
- [ ] 测试覆盖率：API logic 层 30%+

#### Week 2: TDD 纪律强制执行

**目标**: Pipeline 强制验证 TDD 证据

```javascript
// harness-pipeline.js 修改

## TDD 证据检查（强制）
| 新增函数 | RED 确认（必须有具体 FAIL 输出） | GREEN 确认 | 状态 |
|---------|:---:|:---:|:---:|
| FuncName | ✅ "FAIL: undefined: FuncName" | ✅ PASS | PASS |
| FuncName | ❌ "看到失败"（无实际输出） | ✅ PASS | FAIL |

- RED 列缺少具体 FAIL 输出摘录 → 判定 QA FAIL
- Generator 必须输出 TDD 证据到 console
```

**交付物**:
- [ ] 修改 `harness-pipeline.js` qaPrompt()
- [ ] 修改 QA 报告模板（TDD 证据必填）
- [ ] 更新 `.harness/skills/qa/SKILL.md`

#### Week 3: 门禁检查脚本

**目标**: 创建阻断式门禁检查

```bash
#!/usr/bin/env bash
# .harness/scripts/harness-gate-check.sh

# Phase 5 门禁（编码阶段 → 集成验证）
harness-gate-check.sh --phase 5 --change <name>

检查项：
1. 每个服务的 _qa.md 存在
2. 每个服务的 _review*.md 存在
3. QA PASS
4. Review 2/3 或 1/1 PASS
5. CHANGELOG 更新
6. TDD 证据完整（有 RED 输出摘录）

EXIT 1 → 阻塞进入归档，必须修复

# Phase 6 门禁（集成验证 → 交付）
harness-gate-check.sh --phase 6 --change <name>

检查项：
1. impl/ 目录存在
2. summary.md 完整
3. 包含关键章节
4. go build ./... 通过
5. 无 CRITICAL 未解决

EXIT 1 → 阻塞交付
```

**交付物**:
- [ ] `.harness/scripts/harness-gate-check.sh`
- [ ] 集成到 owner-agent.md 流程
- [ ] 添加到 `.harness/rules/项目编码规范.md`

#### Week 4: 前端测试体系

**目标**: 为前端核心组件添加 Vitest 测试

```typescript
// web/pc/src/components/__tests__/PermissionTree.spec.ts

import { mount } from '@vue/test-utils'
import { describe, it, expect, vi } from 'vitest'
import PermissionTree from '../business/PermissionTree.vue'

describe('PermissionTree', () => {
  it('renders tree structure correctly', () => {
    const wrapper = mount(PermissionTree, {
      props: {
        permissions: [
          { id: '1', name: '用户管理', children: [...] }
        ]
      }
    })
    
    expect(wrapper.find('.permission-tree').exists()).toBe(true)
    expect(wrapper.findAll('.tree-node').length).toBe(3)
  })
  
  it('emits selection when checkbox clicked', async () => {
    const wrapper = mount(PermissionTree, { props: {...} })
    
    await wrapper.find('input[type="checkbox"]').trigger('click')
    
    expect(wrapper.emitted('update:selected')).toBeTruthy()
    expect(wrapper.emitted('update:selected')[0]).toEqual([['1']])
  })
  
  // ... 边界/异常用例
})
```

**交付物**:
- [ ] `web/pc/src/components/__tests__/PermissionTree.spec.ts`
- [ ] `web/pc/src/stores/__tests__/division.spec.ts`
- [ ] `web/pc/src/api/__tests__/aimodel.spec.ts` (mock axios)
- [ ] 前端测试覆盖率：20%+

---

## 五、阶段 2：全面推广（8周）

### Week 5-6: API 层测试全覆盖

**目标**: 8 个微服务 API 层覆盖率 60%+

```bash
# 每个服务的核心接口测试
services/user-service/api/internal/logic/
  ├── user/createuserlogic_test.go          ✅
  ├── user/updateuserlogic_test.go          ✅
  ├── user/deleteuserlogic_test.go          ✅
  ├── user/getuserlogic_test.go             ✅
  └── community/listcommunitieslogic_test.go ✅

# 优先级：核心业务 > 辅助功能 > 管理接口
```

**策略**:
- 每天完成 2-3 个 Logic 的测试
- 使用 gomock 生成 Model mock
- 覆盖正常/边界/错误路径
- 测试命名：`Test<Logic>_<Scenario>_<ExpectedResult>`

**交付物**:
- [ ] 8 个服务的 API logic 层测试覆盖率 60%+
- [ ] 测试总数：200+ 测试函数
- [ ] 测试代码：20,000+ 行

### Week 7-8: 集成测试与冒烟测试

**目标**: 建立服务级集成测试

```go
// tests/integration/user_service_test.go

func TestUserService_CreateAndLogin(t *testing.T) {
    // 启动测试容器（MySQL + Redis）
    pool := setupTestContainers(t)
    defer pool.Purge()
    
    // 启动服务
    svc := startUserService(t, pool)
    defer svc.Stop()
    
    // E2E 测试：创建用户 → 登录 → 获取信息
    t.Run("create user", func(t *testing.T) {
        resp, err := http.Post(svc.URL+"/api/users", ...)
        assert.NoError(t, err)
        assert.Equal(t, 200, resp.StatusCode)
    })
    
    t.Run("login", func(t *testing.T) {
        resp, err := http.Post(svc.URL+"/api/auth/login", ...)
        assert.NoError(t, err)
        
        var result struct{ Token string }
        json.NewDecoder(resp.Body).Decode(&result)
        assert.NotEmpty(t, result.Token)
    })
}
```

**交付物**:
- [ ] `tests/integration/` 目录
- [ ] 核心业务流集成测试：5-10 个
- [ ] Docker Compose 测试环境
- [ ] API 冒烟测试自动化

### Week 9-10: QA/Review 全服务覆盖

**目标**: 8 个微服务全部纳入 Pipeline

```bash
# 当前状态：
✅ moderation-service (QA + 3×Review)
✅ ai-model-service (QA + Review)
⏳ master-data-service (部分 Review)

# 需补充（按优先级）：
□ user-service        (P0 - 基础服务)
□ auth-service        (P0 - 基础服务)
□ permission-service  (P1 - 基础服务)
□ community-hub-service (P1 - 核心业务)
□ file-service        (P2 - 支撑服务)
□ monitoring-service  (P3 - 运维服务)
```

**策略**:
- 每周完成 2 个服务的 QA + Review
- 使用 Workflow 并行执行
- 记录 confidence 评分
- 建立趋势监控

**交付物**:
- [ ] 8 个服务完整 QA 报告
- [ ] 8 个服务 3×Review 报告
- [ ] confidence 评分统计

### Week 11-12: 记忆系统优化与文档完善

**目标**: 优化记忆系统，建立知识闭环

```bash
# 1. Memory Suggestions 处理流程自动化
.harness/scripts/memory-apply-suggestions.sh

功能：
- 读取 Pipeline 返回的 memorySuggestions
- 检查 slug 是否已存在
- 创建 status: draft 的记忆文件
- 通知人工审核

# 2. 记忆应用率统计
.harness/scripts/memory-stats.sh

输出：
- 总记忆数
- 应用次数 (apply_count)
- 最近应用时间
- 未应用记忆（候选删除）

# 3. 记忆质量评估
.harness/scripts/memory-quality-check.sh

检查：
- 虚假匹配（M2）
- 遗漏检测（M3）
- 关联链接有效性
```

**交付物**:
- [ ] Memory Suggestions 自动处理脚本
- [ ] 记忆统计脚本
- [ ] 记忆质量检查脚本
- [ ] 更新 26 条记忆的 apply_count
- [ ] 补充 10+ 新记忆（来自 Review 建议）

---

## 六、关键成功指标（KPI）

### 6.1 测试覆盖率

| 层级 | 当前 | 阶段 1 目标 | 阶段 2 目标 | 行业标准 |
|------|:---:|:---:|:---:|:---:|
| **单元测试** | 7.5% | 30% | 60% | 70-80% |
| **API Logic** | 0% | 30% | 60% | 70% |
| **核心引擎** | 83-94% | 保持 | 保持 | 90%+ |
| **前端组件** | <5% | 20% | 50% | 60% |

### 6.2 流程覆盖率

| 维度 | 当前 | 阶段 1 目标 | 阶段 2 目标 |
|------|:---:|:---:|:---:|
| **QA 覆盖** | 25% (2/8服务) | 50% (4/8) | 100% (8/8) |
| **Review 覆盖** | 25% (2/8服务) | 50% (4/8) | 100% (8/8) |
| **TDD 执行率** | 存疑 | 80% | 95% |
| **门禁通过率** | N/A | 100% | 100% |

### 6.3 质量指标

| 指标 | 当前 | 阶段 2 目标 |
|------|:---:|:---:|
| **Bug 逃逸率** | 未统计 | <5% |
| **线上故障** | 未统计 | 0 |
| **回退次数** | 未统计 | <10% |
| **置信度平均值** | N/A | >0.75 |

---

## 七、执行保障机制

### 7.1 强制门禁（阻断式）

```bash
# 提交前门禁
bash .harness/skills/qa/scripts/harness-checks.sh --service <name>
EXIT 1 → 不允许提交

# 阶段 5 门禁
bash .harness/scripts/harness-gate-check.sh --phase 5 --change <name>
EXIT 1 → 不允许进入集成验证

# 阶段 6 门禁
bash .harness/scripts/harness-gate-check.sh --phase 6 --change <name>
EXIT 1 → 不允许交付
```

### 7.2 自动化监控

```bash
# 每日自动运行
.harness/scripts/quality-dashboard.sh

输出：
- 测试覆盖率趋势
- QA/Review 完成度
- 记忆应用率
- 门禁通过率

# 发送到监控服务
POST /api/monitoring/quality-metrics
```

### 7.3 定期回顾

```
Week 2:  阶段 0 完成度检查
Week 5:  阶段 1 中期评审
Week 8:  阶段 1 总结 + 阶段 2 启动
Week 12: 阶段 2 中期评审
Week 16: 全面总结 + 持续改进计划
```

---

## 八、风险与应对

### 风险 1: 工期压力导致测试被忽略

**应对**:
- 门禁强制执行，无测试不合入
- TDD 证据不足 → QA FAIL
- 将测试时间纳入估算

### 风险 2: API 层测试成本高

**应对**:
- 建立测试模板和 mock 库
- 复用性：handler → logic mock 可共享
- 优先核心接口，辅助功能降低优先级

### 风险 3: 开发者抵触

**应对**:
- 展示核心引擎的高质量（83-94% 覆盖率）
- 说明测试带来的信心和重构能力
- Pipeline 自动化减少手工负担

### 风险 4: Pipeline 执行时间长

**应对**:
- 并行执行（多服务、多视角）
- 缓存机制（未变更的服务跳过）
- 增量检查（只测变更文件）

---

## 九、预期收益

### 9.1 短期收益（4周）

- ✅ API 层有测试覆盖，破坏性变更可检测
- ✅ TDD 纪律建立，代码质量提升
- ✅ 门禁阻断低质量代码合入
- ✅ 前端核心组件有测试保护

### 9.2 中期收益（12周）

- ✅ 测试覆盖率达到行业标准（60%）
- ✅ QA/Review 覆盖全部服务
- ✅ 集成测试建立，端到端质量保证
- ✅ 记忆系统完善，错误不重复

### 9.3 长期收益（持续）

- ✅ 重构信心：有测试保护，敢于优化
- ✅ Bug 率下降：线上故障趋近于 0
- ✅ 交付速度：自动化检查节省人工
- ✅ 知识积累：26+ 记忆沉淀，新人快速上手

---

## 十、总结

### 当前状态

**理论完善，实践不足** — 有业界一流的质量框架，但执行率仅 7.5%

### 改进目标

**从理念到实践** — 让 Harness Pipeline 真正驾驭开发过程

### 核心策略

**渐进式落地 + 强制门禁** — 不求一步到位，但确保不倒退

### 成功标准

```
阶段 1 (4周):  测试覆盖率 30%, 门禁建立
阶段 2 (12周): 测试覆盖率 60%, 全服务覆盖
持续改进:      测试覆盖率 70%+, Bug 率 <1%
```

### 一句话愿景

**"让 Harness Pipeline 从展示工具变成强制纪律，从可选项变成不可或缺。"**

---

**下一步行动**: 执行阶段 0 的 4 个 P0 任务，1 周内完成。
