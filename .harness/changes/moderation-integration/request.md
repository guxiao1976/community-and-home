# Request: 内容审核全链路集成

**用户原话**: 
1、所有发布内容的地方，都调用合规审核管线来审核
2、审核管线（AC+模型）审核结果有问题的需要推送给人工审核
3、所有内容审核都要记录审核类型（机器/人工）、审核结果
4、需要开发人工审核界面：按板块、审核状态筛选，点击详情，通过/不通过
5、所有app发布的内容中，记录审核结果（机器通过/不通过/人审通过/不通过）

**路径**: OpenSpec
**理由**: 跨 4 服务 (moderation-service, community-hub-service, user-service, web/pc) + Proto 变更 (api-proto/) + 架构决策 (异步审核流程/Redis消息队列)
**涉及服务**: moderation-service, community-hub-service, user-service, web/pc, api-proto
**创建时间**: 2026-06-17 14:00
