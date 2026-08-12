# Change: access-data-permission — 数据权限核心

## 用户原话(摘要)

L 变更①:数据权限统一模型落地。覆盖:
1. scope 三态(global/限定/空) + 祖先链命中统一规则
2. 能力分层: sys_permission.min_verf_level(0=任意状态+scope / 2=需认证),修订 rbac-design.md §2.5
3. 注册用户基角色 registered_user(正式角色,空数据范围,注册自动分配)
4. 加入小区=自动授权(Join/Leave 联动 AssignRole/RevokeRole,owner/tenant + community scope)
5. 发布校验 AssertPublishScope + publisher_id 取自 JWT
6. master-data 提供"scope 节点→祖先链"解析
7. 商户广告为未来授权来源,模型兼容但不实现

## 工作量分级

- 分级: L
- 命中信号: A=跨4服务 B=是(proto) C=是(common) D>5 E>200 F=是(新API) G=是(数据模型)
- 路由: OpenSpec → N×Pipeline
- 涉及服务: permission-service / user-service / community-hub-service / master-data-service

## 设计依据(已定稿并提交)

- docs/specs/access-control-design.md §1.4 / §3 / §5(commit a2fcc3b)
- 修订 docs/specs/rbac-design.md §2.5 鉴权规则为能力分层

## 阶段状态

- [x] 0 工具选择(request.md)
- [ ] 1 需求分析(requirement-analyst)
- [ ] 2 需求评审(3 视角)
- [ ] 3 架构设计(architecture-designer)
- [ ] 4 Proto 变更(Owner 执行)
- [ ] 5 编码 N×Workflow
- [ ] 6 集成归档
