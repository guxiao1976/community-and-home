---
name: edit-form-data-integrity
description: 编辑回显数据完整性——表单字段必须在全链路 8 层检查数据不丢失
metadata:
  type: project
---

# 编辑回显数据完整性

新增或修改表单字段时，必须验证 **Proto → types.go → Create Logic → Update Logic → Get Logic → TS 类型 → API 函数 → fetchData** 全链路。

## 典型案例：model_type 丢失

- **现象**：编辑模型时"部署方式"下拉框永远显示"云端模型"，与列表行不一致
- **根因**：前端表单有 `model_type` 字段，但 Proto `CreateModelConfigReq` 没有。前端 `...baseData` 发送了，后端静默丢弃
- **影响**：数据库 `model_type` 列存储的是 `type` 字段的值（"chat"），而非用户选择的 "cloud"/"local"
- **修复**：9 个文件逐层添加 `model_type` 字段（Proto ×4 消息、types.go ×3 结构体、API Logic ×4、RPC Logic ×3、TS ×3 处）

## 检查清单

1. Proto — 字段已添加到 Request / Response 消息
2. types.go — 字段已添加到 Create/Update/Info 结构体
3. Create Logic — 传递给 RPC 并入库
4. Update Logic — 传递给 RPC 并更新
5. Get/List Logic — 从 RPC 响应映射到 API 响应
6. TS 类型 — 已添加到 interface
7. TS API — create/update 签名包含
8. fetchData — Object.assign 包含

**少一层 = 数据丢失。**

## 关联

- 规范：[[项目编码规范 §12]]
- 规则文件：`.harness/rules/项目编码规范.md`
