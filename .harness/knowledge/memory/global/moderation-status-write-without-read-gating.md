---
triggers: ["moderation_status", "审核状态", "回调", "审核可见性", "读路径过滤", "审核门禁", "已拒绝内容可见", "审核管线", "不生效"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-08-12
updated: 2026-08-12
---

# 审核状态写入但读路径未按状态过滤，审核流水线无强制执行效果

## 为什么会有这条经验

新增审核回调/写 moderation_status 状态字段时，必须同步在读路径（详情/列表的 FindOne/FindList）应用该状态过滤（未过审内容不返回），否则审核流水线有写无读、状态不生效，被拒绝/待审内容仍对范围内用户可见。

社区案例（community-hub-service，2026-08-12 access-data-permission 阶段④）：UpdateNoticeModerationStatus / UpdateLostFoundModerationStatus 写入 moderation_status，但 model/notice.go、model/lost_found_item.go 的 FindOne/FindList 均无 moderation_status 条件，审核回调后内容可见性无变化。

## 怎么做

若为分期落地，必须在 CHANGELOG 明确「审核可见性门禁待后续 Wave」，避免读者误以为审核已强制执行。写状态与读过滤必须成对评审。

## 怎么验证

- 新审核回调写状态后，读路径是否同步应用状态过滤
- CHANGELOG 是否明确「审核可见性门禁待后续 Wave」分期说明

## 关联经验

[[rpc-callback-must-check-response-base]]
