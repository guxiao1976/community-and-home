## Why

需要将国家统计局2022年统计用区划代码和城乡划分代码数据（全国31个省级单位的五级行政区划）批量导入到 `md_administrative_division` 表中，以初始化行政区划基础数据，支撑业务系统的地址选择、区域管理等核心功能。

## What Changes

- **BREAKING**: 清空 `md_administrative_division` 表全部现有数据
- 新增 `urban_rural_code` 列到 `md_administrative_division` 表（VARCHAR(20)，存储城乡分类代码，仅第五级有值）
- 编写 Python 导入脚本，解析 `/mnt/d/wsl/data/行政区划/` 下 31 个省级 Excel 文件
- 每个文件包含五级数据：省级(110000)→市级(110100)→区县级(110101)→乡镇级(110101001)→村级(110101001001)
- 导入数据统一设置 `submission_status=2`（已审批）、`status=1`（启用）
- 第五级数据同时导入 `urban_rural_code`（城乡分类代码，如"111"表示主城区、"112"表示城乡结合区等）
- 每个文件导入后进行数量校验：Excel数据行数 vs 实际插入行数，不一致则记录到校验报告

## Capabilities

### New Capabilities
- `admin-division-import`: 行政区划数据批量导入工具，包含Excel解析、数据清洗、分层导入、数量校验和错误报告功能

### Modified Capabilities
（无现有spec需要修改）

## Impact

- **数据库**: `md_administrative_division` 表需要新增 `urban_rural_code` 列；导入前清空全表数据
- **数据源**: `/mnt/d/wsl/data/行政区划/` 目录下31个省级Excel文件（每个文件约7000+行五级数据，预计总计60万+条记录）
- **依赖**: Python 3.12 + openpyxl + pymysql（已安装）
- **数据库连接**: `root:123456@tcp(localhost:3306)/masterdata_db`
