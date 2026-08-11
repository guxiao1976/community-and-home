## 项目背景

全平台人员档案库，管理用户基础信息、实名状态、信用分。

## 数据模型

- **MySQL (`db_user`)**:
    - `user_base`: `id`, `phone(EncryptedString)`, `nickname`, `cert_status(0未/1待/2通过/3驳)`, `credit_score(默认100)`

## 接口清单

- `CreateUser(CreateUserReq) returns (UserBaseResp)`
- `GetUserProfile(GetUserReq) returns (UserProfileResp)`
- `SubmitIdentity(SubmitIdentityReq) returns (BaseResp)`
- `UpdateIdentityStatus(UpdateIdentityStatusReq) returns (BaseResp)` (MQ消费)
- `DeductCredit(DeductCreditReq) returns (BaseResp)` (MQ消费)

## 核心逻辑流

1. **CreateUser**:
    - 事务内：插入 `user_base` (手机号自动 AES 加密) -> 调用 Auth-Service gRPC 创建凭证 (若失败则回滚)。
2. **SubmitIdentity**:
    - 校验 `cert_status` 必须为 0（幂等）。
    - **同步 gRPC 调用** Approval-Service 的 `CreateApplication`。
    - 更新 `cert_status = 1` (待审)。
3. **UpdateIdentityStatus (消费 topic_approval_result)**:
    - 幂等校验：检查当前 `cert_status` 必须为 1。
    - 根据消息内容更新 `cert_status = 2` 或 `3`。
4. **DeductCredit (消费 topic_risk_credit_deduct)**:
    - 幂等校验：基于去重表或状态校验。
    - 更新 `credit_score = credit_score - deduct_points` (DB 层面做防负数校验 `WHERE credit_score >= deduct_points`)。
