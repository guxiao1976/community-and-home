# OpenSpec to Ralph Bridge

## 触发

"生成执行计划" / "openspec to ralph" / "导出任务到 Ralph" / "创建 fix_plan"

## 功能

将 OpenSpec 的 `tasks.md` 按服务拆分，为每个服务生成 Ralph 可执行的 `fix_plan.md`。

## 流程

### Step 1: 定位 OpenSpec change

从用户输入或上下文中确定 change 名称，读取：
```
openspec/changes/<change-name>/
  tasks.md       ← 任务清单
  design.md      ← 技术设计（了解服务归属）
```

### Step 2: 按服务拆分任务

解析 `tasks.md`，识别每个任务的服务归属：
- 任务前缀如 `[proto]` / `[global]` → 全局 Claude 执行
- `## <服务名>` 章节 → 归属该服务
- 任务描述中包含服务名 → 归属对应服务

### Step 3: 为每个服务生成 fix_plan.md

对每个受影响的服务，生成 `services/<name>/.ralph/fix_plan.md`：

```markdown
# Fix Plan: <change-name> — <service-name>

> 来源：openspec/changes/<change-name>/tasks.md
> 生成时间：YYYY-MM-DD HH:MM
> 关联设计：openspec/changes/<change-name>/design.md

## 前置阅读
1. 服务 CLAUDE.md
2. docs/design.md
3. CHANGELOG.md
4. .harness/knowledge/memory/MEMORY.md ← 加载经验

## 任务清单

- [ ] <task-id> <任务描述>
  - 关联 spec: `<spec-ref>`
  - 验收标准: <criteria>

- [ ] <task-id> <任务描述>
  ...
```

### Step 4: 全局任务保留

Proto 变更等全局任务保留在原地，提示用户由全局 Claude 执行：

```
⚠️ 以下任务需要全局 Claude 执行（子 Claude 不能修改 api-proto/）：
- [ ] 0.1 修改 api-proto/api/<svc>/v1/<file>.proto
- [ ] 0.2 cd api-proto && make generate && make lint && make breaking-check
```

## 输出

```
✅ 已为 N 个服务生成执行计划：
  - services/<svc1>/.ralph/fix_plan.md (M 个任务)
  - services/<svc2>/.ralph/fix_plan.md (K 个任务)
  - web/pc/.ralph/fix_plan.md (P 个任务)

⚠️ 1 个全局任务需要亲自执行（Proto 变更）

下一步：切换到各服务目录，运行 ralph start
```
