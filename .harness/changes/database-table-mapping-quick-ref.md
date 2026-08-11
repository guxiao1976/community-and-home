# 数据表归属方案 - 快速参考

**生成日期**: 2026-06-24  
**项目**: community-and-home

---

## 📊 数据表归属一览表

### 1. user 数据库 (5张表)

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `user_base` | 用户基本信息 | user_base.go |
| `user_certification` | 用户认证信息 | user_certification.go |
| `user_community_membership` | 用户社区成员关系 | user_community_membership.go |
| `user_membership_role` | 用户成员角色 | user_membership_role.go |
| `user_residence` | 用户居住信息 | user_residence.go |

---

### 2. auth 数据库 (1张表)

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `auth_credentials` | 认证凭证（密码等） | authcredential.go |

**建议扩展**:
- `auth_tokens` - 访问令牌
- `refresh_tokens` - 刷新令牌
- `login_logs` - 登录日志

---

### 3. permission 数据库 (2张表)

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `permission` | 权限定义 | permission.go |
| `rel_*` | 各种关系表 | rel.go |

**建议扩展**:
- `roles` - 角色
- `role_permissions` - 角色权限关联
- `user_roles` - 用户角色关联

---

### 4. community_hub_db 数据库 (4张表)

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `community_contact` | 社区联系人 | community_contact.go |
| `lost_found_item` | 失物招领 | lost_found_item.go |
| `notice` | 公告通知 | notice.go |
| `notice_attachment` | 公告附件 | notice_attachment.go |

**建议扩展**:
- `posts` - 帖子
- `comments` - 评论
- `likes` - 点赞

---

### 5. moderation_db 数据库 (4张表)

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `mod_audit_log` | 审核日志 | mod_audit_log_model.go |
| `mod_pipeline_config` | 管道配置 | mod_pipeline_config_model.go |
| `md_pipeline_test_work_item` | 测试工作项 | - |
| `md_pipeline_test_work_record` | 测试工作记录 | - |

---

### 6. ai_model_db 数据库 (8张表)

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `am_api_key` | API 密钥 | amApiKeyModel.go |
| `am_alert_record` | 告警记录 | amalertrecordmodel.go |
| `am_call_log` | 调用日志 | amcalllogmodel.go |
| `am_cost_alert_config` | 成本告警配置 | amcostalertconfigmodel.go |
| `am_health_check` | 健康检查 | amhealthcheckmodel.go |
| `am_model_config` | 模型配置 | ammodelconfigmodel.go |
| `am_prompt_template` | 提示模板 | amprompttemplatemodel.go |
| `am_usage_statistics` | 使用统计 | amusagestatisticsmodel.go |

---

### 7. masterdata_db 数据库 (9张表)

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `md_administrative_division` | 行政区划 | mdAdministrativeDivisionModel.go |
| `md_audit_log` | 审核日志 | mdAuditLogModel.go |
| `md_configuration` | 配置 | mdConfigurationModel.go |
| `md_district_economic_data` | 区域经济数据 | mdDistrictEconomicDataModel.go |
| `md_division_statistics` | 区划统计 | mdDivisionStatisticsModel.go |
| `md_outbox_message` | 消息外箱 | mdOutboxMessageModel.go |
| `md_residential_area` | 小区数据 | mdResidentialAreaModel.go |
| `md_sensitive_word` | 敏感词 | mdSensitiveWordModel.go |
| `md_submission_record` | 提交记录 | mdSubmissionRecordModel.go |

---

### 8. file_db 数据库 (2张表)

| 表名 | 说明 | Model 文件 |
|------|------|-----------|
| `file` | 文件元数据 | file.go |
| `file_model` | 文件模型 | filemodel.go |

---

## 🎯 总计

| 数据库 | 表数量 | 服务 |
|--------|:---:|------|
| user | 5 | user-service |
| auth | 1+ | auth-service |
| permission | 2+ | permission-service |
| community_hub_db | 4+ | community-hub-service |
| moderation_db | 4 | moderation-service |
| ai_model_db | 8 | ai-model-service |
| masterdata_db | 9 | master-data-service |
| file_db | 2 | file-service |
| **总计** | **35+** | **8个服务** |

---

## 📋 快速查找

### 按功能查找

**用户相关**:
- 用户信息 → `user.user_base`
- 用户认证 → `user.user_certification`
- 认证凭证 → `auth.auth_credentials`

**权限相关**:
- 权限定义 → `permission.permission`
- 角色关联 → `permission.rel_*`

**社区相关**:
- 社区联系 → `community_hub_db.community_contact`
- 失物招领 → `community_hub_db.lost_found_item`
- 公告通知 → `community_hub_db.notice`

**审核相关**:
- 审核日志 → `moderation_db.mod_audit_log`
- 管道配置 → `moderation_db.mod_pipeline_config`

**AI 相关**:
- 模型配置 → `ai_model_db.am_model_config`
- 调用日志 → `ai_model_db.am_call_log`
- 提示模板 → `ai_model_db.am_prompt_template`

**主数据相关**:
- 行政区划 → `masterdata_db.md_administrative_division`
- 小区数据 → `masterdata_db.md_residential_area`
- 敏感词 → `masterdata_db.md_sensitive_word`

---

## ✅ 验证命令

```bash
# 查看所有数据库
docker exec -i mysql mysql -uroot -proot123456 -e "SHOW DATABASES;"

# 查看某个数据库的表
docker exec -i mysql mysql -uroot -proot123456 -e "SHOW TABLES FROM user;"

# 查看表结构
docker exec -i mysql mysql -uroot -proot123456 -e "DESC user.user_base;"
```

---

**文档版本**: 1.0  
**最后更新**: 2026-06-24
