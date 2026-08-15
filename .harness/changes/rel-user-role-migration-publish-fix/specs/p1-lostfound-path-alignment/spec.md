# P1 移动端寻失列表路径对齐 Specification

## Purpose

消除移动端寻失列表请求路径与 community-hub-service 实际注册路由不一致导致的 HTTP 404 与「列表页永远空白且无报错」，使已加入小区的用户能正常加载寻失数据；后端路由保持不动，修复方向为前端对齐后端。

## Requirements

### Requirement: REQ-P1-PATH-1 前端请求对齐后端注册路由

The mobile lost-and-found list retrieval SHALL request the route registered by community-hub-service (`GET /api/community/lostfound`, no hyphen), and SHALL NOT request the unregistered path `GET /api/community/lost-found`.

后端路由定位（统一引用，proposal 与本 spec 一致）：`services/community-hub-service/api/internal/handler/routes.go` 的 lostfound 段，`/api/community/lostfound` 四个路由注册于约 54-73 行（Create=POST 56-57 / **List=GET 60-62** / Get=GET 65-67 / Resolve=POST 70-72）。行号仅作定位锚点，验收以路由路径文本为准。

#### Scenario: 正确路径加载寻失列表（正向）

- **GIVEN** 用户已加入小区且后端 community-hub-service 已注册 `GET /api/community/lostfound`（`services/community-hub-service/api/internal/handler/routes.go` lostfound 段，List 注册于 60-62 行）
- **WHEN** `notice.vue` 触发 `fetchLostFound` 加载寻失列表
- **THEN** 请求命中 `GET /api/community/lostfound`，返回数据正常渲染到「寻失互助」区，不再 404

#### Scenario: 无旧路径残留（边界）

- **GIVEN** 本变更已完成
- **WHEN** 检索 `web/mobile` 全仓代码
- **THEN** 不存在对 `GET /api/community/lost-found`（带连字符）的调用残留，仅存正确路径 `/api/community/lostfound`

#### Scenario: 后端服务不可用（系统异常）

- **GIVEN** 后端 community-hub-service 未启动或网络中断
- **WHEN** 触发 `fetchLostFound`
- **THEN** 请求因网络/服务不可达失败（而非路径 404），失败反馈按 REQ-P1-ERR-1（toast + console 日志）兜底提示，不静默

#### Scenario: 后端路由未来变化（边界）

- **GIVEN** 后续若后端修改 `/api/community/lostfound` 路由
- **WHEN** 前端仍按该路径调用
- **THEN** 调用失败且被 REQ-P1-ERR-1 显式提示（路径错误不再被静默吞掉）；本变更不改动后端路由

### Requirement: REQ-P1-PATH-2 后端路由保持不动

The change SHALL NOT modify the community-hub-service route registration; alignment is achieved by changing the mobile client only.

#### Scenario: 后端无变更（正向）

- **GIVEN** 后端 `services/community-hub-service/api/internal/handler/routes.go` 已注册 `/api/community/lostfound`（含 Create/List/Get/Resolve）
- **WHEN** 本变更完成
- **THEN** community-hub-service 的 routes.go 无任何改动，发布链路（CreateLostFound/ResolveLostFound）不受影响

#### Scenario: 后端路由为修复前置契约（边界）

- **GIVEN** 本变更以「后端 `/api/community/lostfound` 已注册」为修复前置契约
- **WHEN** 未来后端删除或改名该路由
- **THEN** 前端列表加载失败，且由 REQ-P1-ERR-1 显式 toast + console 提示而非静默（路径错误具备可见性）；路由回归不在本变更范围，需另行评审
