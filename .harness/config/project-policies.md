# 项目策略归属（Project Policies）

> 本文件说明 Community-Home 项目特有策略的归属，对齐 `harness-design-principles.md` 原则 16「引擎与策略分离」。
> `harness-checks.sh` 头部标注为「项目特有策略」的检查，其策略定义、约束来源、迁移替换方式记录于此。

## 策略清单

| 策略 | 检查项（harness-checks.sh） | 约束来源 | 迁移到其他项目时 |
|------|---------------------------|---------|----------------|
| Snowflake ID 精度 | #4 `proto_jstype`、#5 `json_string` | `.harness/rules/项目编码规范.md §5` | 若新项目不用 Snowflake（改用 UUID/自增），可删除这两项检查 |
| 跨服务仅 gRPC | #6 `cross_service_import` | `.harness/rules/项目编码规范.md §1` | 若新项目非微服务或允许直连 DB，改写为对应的边界约束 |
| 5 位 errx 错误码 | #7 `error_codes` | `.harness/rules/项目编码规范.md`（错误码规范） | 若新项目用其他错误码体系，替换检测模式与常量来源 |

## 归属原则

- **通用引擎**：`go build/vet/test`、硬编码密钥、TODO 桩、API 冒烟、图谱新鲜度、Memory 索引等——不随项目业务变化，可跨项目复用。
- **项目策略**：上表 3 类——依赖本项目技术选型（Snowflake、gRPC-only、errx 5 位错误码），迁移时需替换。

## 维护约定

- 新增「项目特有」检查时，同步在 `harness-checks.sh` 头部注释 + 本文件登记。
- 策略约束变更时，先改 `.harness/rules/项目编码规范.md`，再同步本文件。
