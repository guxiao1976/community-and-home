# 行政区划数据丢失问题分析报告

## 问题描述
2026-07-12 20:23

**现象**: 行政区划页面显示为空，但用户反馈"之前是有数据的"

---

## 🔍 问题诊断

### 数据库检查结果

**1. 表结构存在 ✅**
```sql
mysql> SELECT COUNT(*) FROM md_administrative_division;
+----------+
| COUNT(*) |
+----------+
|        0 |
+----------+
```

**2. 表为空 ❌**
- 当前记录数：0
- 表结构：完整
- 索引：正常

**3. API 返回正常**
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [],
    "total": 0
  }
}
```

---

## 💡 可能的原因

### 原因1: 数据库重新初始化（最可能）

在今天的数据库初始化过程中：
```bash
# 执行了数据库初始化脚本
docker exec -i mysql mysql -uroot -proot123456 masterdata_db < \
  services/master-data-service/migration/001_initial_schema.sql
```

**`001_initial_schema.sql` 只创建表结构，不包含数据**：
```sql
CREATE TABLE IF NOT EXISTS md_administrative_division (
    id bigint NOT NULL AUTO_INCREMENT,
    parent_id bigint NULL,
    level tinyint NOT NULL,
    name varchar(100) NOT NULL,
    code varchar(20) NOT NULL,
    ...
) ENGINE=InnoDB;
```

### 原因2: 缺少数据导入脚本

检查 migration 目录：
```
services/master-data-service/migration/
├── 001_initial_schema.sql      ✅ 表结构
├── 002_create_outbox_messages.sql
├── 003_system_config_refactor.sql
└── (缺少数据导入脚本)          ❌ 初始数据
```

**结论**: 没有包含行政区划初始数据的SQL文件

### 原因3: 数据在初始化时被清空

如果之前有数据，可能的清空场景：
- 执行了 `DROP DATABASE` 然后重建
- 执行了 `TRUNCATE TABLE`
- 执行了带 `DROP TABLE` 的脚本

---

## 📋 数据恢复方案

### 方案1: 从备份恢复（如果有备份）

**检查是否有备份**：
```bash
# 查找可能的备份文件
find / -name "*masterdata*.sql" -mtime -7 2>/dev/null
find / -name "*division*.sql" -mtime -7 2>/dev/null

# Docker volume 备份
docker volume ls | grep mysql
```

**恢复步骤**：
```bash
# 如果找到备份文件
docker exec -i mysql mysql -uroot -proot123456 masterdata_db < backup.sql
```

### 方案2: 从其他环境导出（如果有生产/测试环境）

```bash
# 在有数据的环境导出
mysqldump -u root -p masterdata_db md_administrative_division > divisions_data.sql

# 在本地导入
docker exec -i mysql mysql -uroot -proot123456 masterdata_db < divisions_data.sql
```

### 方案3: 重新导入标准数据

**需要创建数据导入脚本**：

`services/master-data-service/migration/004_seed_divisions.sql`

```sql
-- 中国省级行政区划数据 (示例)
INSERT INTO md_administrative_division 
(parent_id, level, name, code, path, sort_order, status, submission_status) 
VALUES
(NULL, 1, '北京市', '110000', '/1/', 1, 1, 2),
(NULL, 1, '天津市', '120000', '/2/', 2, 1, 2),
(NULL, 1, '河北省', '130000', '/3/', 3, 1, 2),
(NULL, 1, '山西省', '140000', '/4/', 4, 1, 2),
(NULL, 1, '内蒙古自治区', '150000', '/5/', 5, 1, 2),
(NULL, 1, '辽宁省', '210000', '/6/', 6, 1, 2),
(NULL, 1, '吉林省', '220000', '/7/', 7, 1, 2),
(NULL, 1, '黑龙江省', '230000', '/8/', 8, 1, 2),
(NULL, 1, '上海市', '310000', '/9/', 9, 1, 2),
(NULL, 1, '江苏省', '320000', '/10/', 10, 1, 2),
(NULL, 1, '浙江省', '330000', '/11/', 11, 1, 2),
(NULL, 1, '安徽省', '340000', '/12/', 12, 1, 2),
(NULL, 1, '福建省', '350000', '/13/', 13, 1, 2),
(NULL, 1, '江西省', '360000', '/14/', 14, 1, 2),
(NULL, 1, '山东省', '370000', '/15/', 15, 1, 2),
(NULL, 1, '河南省', '410000', '/16/', 16, 1, 2),
(NULL, 1, '湖北省', '420000', '/17/', 17, 1, 2),
(NULL, 1, '湖南省', '430000', '/18/', 18, 1, 2),
(NULL, 1, '广东省', '440000', '/19/', 19, 1, 2),
(NULL, 1, '广西壮族自治区', '450000', '/20/', 20, 1, 2),
(NULL, 1, '海南省', '460000', '/21/', 21, 1, 2),
(NULL, 1, '重庆市', '500000', '/22/', 22, 1, 2),
(NULL, 1, '四川省', '510000', '/23/', 23, 1, 2),
(NULL, 1, '贵州省', '520000', '/24/', 24, 1, 2),
(NULL, 1, '云南省', '530000', '/25/', 25, 1, 2),
(NULL, 1, '西藏自治区', '540000', '/26/', 26, 1, 2),
(NULL, 1, '陕西省', '610000', '/27/', 27, 1, 2),
(NULL, 1, '甘肃省', '620000', '/28/', 28, 1, 2),
(NULL, 1, '青海省', '630000', '/29/', 29, 1, 2),
(NULL, 1, '宁夏回族自治区', '640000', '/30/', 30, 1, 2),
(NULL, 1, '新疆维吾尔自治区', '650000', '/31/', 31, 1, 2),
(NULL, 1, '台湾省', '710000', '/32/', 32, 1, 2),
(NULL, 1, '香港特别行政区', '810000', '/33/', 33, 1, 2),
(NULL, 1, '澳门特别行政区', '820000', '/34/', 34, 1, 2);
```

执行导入：
```bash
docker exec -i mysql mysql -uroot -proot123456 masterdata_db < \
  services/master-data-service/migration/004_seed_divisions.sql
```

### 方案4: 通过前端管理界面手动添加

如果数据量不大，可以通过管理界面手动录入。

---

## 🎯 建议的解决步骤

### 立即行动

1. **确认是否有备份**
```bash
# 检查 Docker volumes
docker volume inspect mysql_data

# 检查项目目录
ls -lah /home/jiaoxh/my-project/community-and-home/backup/
```

2. **如果有备份，立即恢复**
```bash
docker exec -i mysql mysql -uroot -proot123456 masterdata_db < backup_file.sql
```

3. **如果没有备份，导入基础数据**
- 创建 `004_seed_divisions.sql`
- 至少导入省级数据（34条）
- 后续可以补充市、区县数据

---

## 🛡️ 预防措施

### 1. 添加数据备份脚本

`scripts/backup-database.sh`
```bash
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/home/jiaoxh/backups"
mkdir -p $BACKUP_DIR

docker exec mysql mysqldump -uroot -proot123456 \
  masterdata_db > $BACKUP_DIR/masterdata_$DATE.sql

# 保留最近7天的备份
find $BACKUP_DIR -name "masterdata_*.sql" -mtime +7 -delete
```

### 2. 区分schema和data脚本

```
migration/
├── 001_initial_schema.sql       # 只有表结构
├── 004_seed_divisions.sql       # 行政区划数据
├── 005_seed_residential.sql     # 小区数据
└── 006_seed_sensitive_words.sql # 敏感词数据
```

### 3. 在migration中添加注释

```sql
-- Migration 001: Schema only, NO data
-- To populate initial data, run:
-- - 004_seed_divisions.sql
-- - 005_seed_residential.sql
```

---

## 📊 数据重要性评估

### 行政区划数据

**重要性**: 🔴 高

**影响范围**:
- 用户注册（选择地址）
- 小区管理（地理位置）
- 数据统计（区域分析）
- 权限管理（区域管理员）

**恢复优先级**: P0 - 立即处理

---

## 总结

### 问题确认
✅ 表结构存在  
❌ 数据为空  
⚠️ 可能在今天的数据库初始化中丢失  

### 解决方案
1. 优先从备份恢复
2. 如无备份，导入标准行政区划数据
3. 建立数据备份机制

### 下一步
1. 检查是否有数据备份
2. 创建行政区划数据导入脚本
3. 建立定期备份机制

---

**报告时间**: 2026-07-12 20:23  
**优先级**: P0  
**建议**: 立即恢复数据
