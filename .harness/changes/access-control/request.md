# Change: access-control

## 用户原话(摘要)

跨 PC/移动端访问控制与数据权限改造,覆盖:
1. 角色分层: 注册用户(基角色,空数据范围) → 未认证业主/租户(加入即自动授权,可发布配额内,不可选举) → 认证业主(房产证AI认证,解锁选举);退出撤销
2. 端限制: 角色 platforms 配置驱动,移动端角色禁止 PC 登录(UX引导,非安全边界)
3. 数据权限统一模型: 祖先链命中规则 + scope 三态(global/限定/空)
4. 当前小区服务端持久化(user-service)
5. 板块发布配额(占位状态定义: 待审/展示占,驳回/解决/删除/下架释放)
6. 房屋注册上限 ≤6 + 同屋互见手机号(防冒充)
7. 商户广告投放(未来,模型兼容,本次不实现)

## 工作量分级

- 分级: L
- 命中信号: A=跨5服务+2前端 B=是(proto) C=是(common) D=远超5文件 E=远超200行 F=是(新API) G=是(数据模型/表结构) H=清晰
- 理由: 跨服务 + Proto + common + 新增公开 API + 架构决策
- 路由: OpenSpec → N×Pipeline
- 涉及服务: auth-service / permission-service / user-service / community-hub-service / master-data-service + web/pc + web/mobile

## 设计依据

- docs/specs/access-control-design.md(已提交 a2fcc3b)
- 修订 rbac-design.md §2.5(status 从"能否用"改为"能力分层",权限码加 min_verf_level)

## 阶段状态

- [x] 0 工具选择(dispatch,request.md)
- [ ] 1 需求分析(requirement-analyst)
- [ ] 2 需求评审(3 视角)
- [ ] 3 架构设计(architecture-designer → design.md + tasks.md)
- [ ] 4 Proto 变更(Owner 执行,make ci)
- [ ] 5 编码 N×Workflow(每服务 harness-pipeline.js)
- [ ] 6 集成归档
