# 任务系统维护指南

## 日常

- **Loop 每次运行时**：运行 `harness-tasks.sh scan --auto-create` 扫描新问题
- **Loop 开始执行任务时**：运行 `harness-tasks.sh status --id <id> --status in_progress`
- **Loop 完成任务后**：运行 `harness-tasks.sh status --id <id> --status closed`
- **任务完成后**：检查是否需要归档到 `_archive/`（关闭超过 7 天的任务）

## 每周

- **审查 open 任务**：
  - P0 任务超过 3 天未关闭？→ 升级给人
  - P1 任务超过 7 天未关闭？→ 检查是否阻塞
  - 是否有任务应该降低优先级（不再紧急）？
- **更新任务完成标准**：需求更明确后补充 `完成标准` checklist
- **合并重复任务**：`grep -rl "相同关键词" .harness/tasks/task-*.md`
- **运行 `harness-tasks.sh index`**：确保 BACKLOG.md 与实际文件一致

## 每月

- **分析已完成任务**（`_archive/` 中的任务）：
  - 按 source 分类：qa 发现的 / review 发现的 / sensor 发现的 / human 安排的
  - 哪种来源的任务最多？→ 考虑加强对应的传感器或预防措施
  - 哪些任务反复出现？→ 考虑升级为 `.harness/rules/` 中的硬规则或 `harness-checks.sh` 检查项
- **清理**：删除 3 个月以上且已关闭的归档任务
- **调整优先级**：根据实际情况重新评估 P2/P3 任务的优先级

## 命令速查

```bash
# 查看所有 open 任务（按优先级）
bash .harness/scripts/harness-tasks.sh list --status open | sort -k2

# 查看某个服务的待办
bash .harness/scripts/harness-tasks.sh list --service moderation-service --status open

# 扫描新问题（不自动创建）
bash .harness/scripts/harness-tasks.sh scan

# 扫描并自动创建任务
bash .harness/scripts/harness-tasks.sh scan --auto-create

# 创建新任务
bash .harness/scripts/harness-tasks.sh create \
  --title "任务标题" \
  --service master-data-service \
  --priority P1 \
  --type feature \
  --detail "详细描述"

# 更新任务状态
bash .harness/scripts/harness-tasks.sh status --id task-2026-06-16-001 --status in_progress

# 统计数据
bash .harness/scripts/harness-tasks.sh stats

# 重建索引
bash .harness/scripts/harness-tasks.sh index

# 归档已完成任务
mv .harness/tasks/task-*.md .harness/tasks/_archive/  # 手动归档已关闭的
```

## 和 Memory 系统的协作

- 如果某个任务反复出现（相同类型的问题被多次创建 task），说明应该有一条记忆来预防它
- 在 `.harness/knowledge/memory/` 中创建对应记忆，`type: pitfall`
- 在任务文件的「关联」部分用 `[[memory-slug]]` 链接相关记忆

## 和 Changes 系统的协作

- 当任务属于某个 OpenSpec 变更的一部分时，在 `source_detail` 中记录 `.harness/changes/<name>/`
- 变更完成时，将所有关联任务标记为 `closed`
- 在 `summary.md` 的「例外 & 未解决问题」中记录未关闭的关联任务
