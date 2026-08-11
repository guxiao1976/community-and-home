## 基本信息
- 名称：审批中心
- 功能：低频长流程工作流。接收业务请求，流转审批，**严禁同步回调业务方**。

## 数据模型

- **MySQL (`db_approval`)**:
    - `bpm_application`: `id`, `biz_type(identity/room)`, `biz_id`, `applicant_id`, `status(pending/approved/rejected)`
    - `bpm_task_record`: `id`, `application_id`, `approver_id`, `action(approve/reject)`, `comment`

## 接口清单

- `CreateApplication(CreateAppReq) returns (CreateAppResp)`
- `GetMyPendingTasks(GetPendingTasksReq) returns (TaskListResp)`
- `ApproveTask(ApproveTaskReq) returns (BaseResp)`

## 核心逻辑流

1. **CreateApplication**: 创建 `bpm_application` 记录，返回 Application ID。
2. **ApproveTask**:
    - 更新 `bpm_application` 状态为 approved/rejected。
    - 插入 `bpm_task_record` 审批记录。
    - **Outbox 模式强制**：在同一事务中写入 `outbox_messages`，Event Type 为 `topic_approval_result`，Payload 包含 `biz_type`, `biz_id`, `status`。**绝对禁止在审批逻辑中 gRPC 调用 User/Master 服务修改状态**。
