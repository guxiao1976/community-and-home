# 记忆系统维护指南

## 日常

- 每次 Agent 执行完毕后，检查是否有新记忆被创建（`git status .harness/memory/`）
- 新记忆 status: draft → 快速浏览，确认内容合理 → 改为 status: active
- 确认 `type` 字段正确：`pitfall`=踩过的坑 | `guideline`=编码/架构规范 | `process`=流程约束 | `decision`=技术决策 | `model`=数据模型

## 每周

- 审查 status: active 的记忆文件：
  - 是否有类似经验重复出现？ → 合并
  - 是否有经验已经不再适用？ → 改为 status: superseded
  - triggers 关键词是否准确？ → 补充或修正
  - type 分类是否仍然准确？ → 某些 pitfall 被内化为 guideline 后应改类型

## 每月

- 分析 superseded 记忆，总结模式（什么类型的经验容易过时？）
- 将高频触发的经验提升为 `.harness/rules/项目编码规范.md` 中的硬规则
- 将稳定的 `guideline` 类记忆考虑加入 `harness-checks.sh` 机械化检查
- 清理 3 个月以上的 superseded 记忆（git rm）

## 命令速查

```bash
# 查看所有记忆
ls -la .harness/memory/ services/*/.harness/memory/

# 查找特定主题的记忆
grep -rl "关键词" .harness/memory/ services/*/.harness/memory/

# 查看记忆状态统计
grep -r "status:" .harness/memory/ | sort | uniq -c

# 列出 superseded 记忆
grep -rl "status: superseded" .harness/memory/
```
