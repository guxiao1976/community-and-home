# Change: harness-pipeline-fix — 管线执行层修复

## 背景
access-control 变更 Wave 1 用 harness 管线跑了 2 次(~130万token),产出 0 行实现。
启动器(解析/沙箱/QA脚本)已修,执行层仍坏。

## 问题清单(真实失败信号)
| # | 问题 | 证据 |
|---|------|------|
| 1 | Generator 不聚焦 tasks.md,未实现目标文件 | scope.go/assertpublishscopelogic.go 不存在 |
| 2 | 实现+测试滞留隔离 worktree,不落主树 | QA 报"测试在 wf_xxx worktree 未 merge" |
| 3 | QA 审错范围(审旧 RBAC 代码 54e1a60,非本次 diff) | QA 报告验证范围=54e1a60+工作树 |
| 4 | TDD 证据缺失(RED→GREEN 无摘录,测试未 commit) | QA TDD FAIL |
| 5 | 任务分类 chore/feature 误判(临时用 args.taskType 绕过) | 首次跑 taskType=chore |

## 修复目标
1. Generator 严格按 tasks.md 逐任务实现 + 提交
2. 测试与生产代码同 commit 落主树
3. QA 只审本次 diff 新增函数
4. TDD 证据强制捕获
5. 任务分类修正

## 验收
- 小任务回归:单服务小改动,管线 ≤2 轮收敛,产出实现+测试
- 接入 access-control 变更① Wave 1 重跑出真实现

## 阶段
- [x] 0 request
- [ ] 1 定向审计
- [ ] 2 修复
- [ ] 3 小任务回归
- [ ] 4 接入 Wave 1
