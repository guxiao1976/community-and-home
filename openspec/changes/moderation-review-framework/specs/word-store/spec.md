## ADDED Requirements

### Requirement: 词库加载
系统 SHALL 从 masterdata_db 的 `md_sensitive_word` 表加载 `status=1` 的记录，区分 word_type=1（黑名单）和 word_type=2（白名单），分别构建 AC 自动机。

#### Scenario: 首次加载
- **WHEN** 服务启动
- **THEN** 从数据库加载词库，构建黑名单 AC 自动机和白名单 AC 自动机

#### Scenario: 空词库
- **WHEN** 数据库中无 status=1 的记录
- **THEN** 构建空的 AC 自动机，服务正常运行（匹配始终返回空）

### Requirement: 定时同步
系统 SHALL 定期（默认 300 秒）检查词库版本，有变更时自动重建 AC 自动机。

#### Scenario: 检测到变更
- **WHEN** 定时检查发现数据库中 MAX(updated_time) 大于当前版本
- **THEN** 重新加载词库并重建 AC 自动机

#### Scenario: 无变更跳过
- **WHEN** 定时检查发现版本未变
- **THEN** 不执行重建，AC 自动机继续使用

### Requirement: 版本号机制
系统 SHALL 基于词库中 `MAX(updated_time)` 的 Unix 时间戳作为版本号，存储在内存中。

#### Scenario: 版本号更新
- **WHEN** 有新词被创建或更新
- **THEN** MAX(updated_time) 增加，版本号变化

### Requirement: 数据库连接
系统 SHALL 通过配置文件中的 DataSource 直连 masterdata_db，不通过 RPC 调用 masterdata 服务。

#### Scenario: 连接配置
- **WHEN** 配置中 DataSource 指向 masterdata_db
- **THEN** 系统直接连接该数据库读取敏感词表
