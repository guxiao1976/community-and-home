## ADDED Requirements

### Requirement: Clear existing division data
脚本启动时 SHALL 清空 `md_administrative_division` 表全部数据（`TRUNCATE TABLE`），为全量导入做准备。

#### Scenario: Table has existing data
- **WHEN** 脚本启动执行
- **THEN** 系统执行 `TRUNCATE TABLE md_administrative_division` 清空所有现有数据

#### Scenario: Table is empty
- **WHEN** 脚本启动执行且表为空
- **THEN** TRUNCATE 执行成功，不影响后续导入流程

### Requirement: Add urban_rural_code column
系统 SHALL 在 `md_administrative_division` 表中新增 `urban_rural_code VARCHAR(20) NULL` 列，位于 `sort_order` 之后。

#### Scenario: Column does not exist
- **WHEN** 脚本执行 ALTER TABLE
- **THEN** `urban_rural_code` 列成功添加

#### Scenario: Column already exists
- **WHEN** 脚本执行 ALTER TABLE 且列已存在
- **THEN** 脚本 SHALL 捕获异常并跳过，不中断执行

### Requirement: Parse Excel files for five-level divisions
脚本 SHALL 读取 `/mnt/d/wsl/data/行政区划/` 目录下所有 `*.xlsx` 文件（排除 `~$` 临时文件），解析每个文件中的五级行政区划数据。

#### Scenario: Valid Excel file
- **WHEN** 脚本读取一个省级Excel文件
- **THEN** 解析表头（Row 1），跳过统计行（Row 2），从 Row 3 开始提取五级数据

#### Scenario: Temporary Excel file
- **WHEN** 目录中存在 `~$` 前缀文件
- **THEN** 脚本自动跳过这些文件

#### Scenario: Excel columns mapping
- **WHEN** 解析数据行
- **THEN** 映射关系为：A=省级代码, B=省级名称, C=市级代码, D=市级名称, E=区县级代码, F=区县名称, G=乡镇级代码, H=乡镇名称, I=村级代码, J=城乡分类代码, K=村(居委)名称

### Requirement: Import data with correct field values
脚本 SHALL 将所有导入记录设置 `submission_status=2`（已审批）和 `status=1`（启用）。

#### Scenario: All imported records
- **WHEN** 任意级别的区划数据被插入
- **THEN** 该记录的 `submission_status=2` 且 `status=1`

### Requirement: Import urban_rural_code for level 5
脚本 SHALL 仅在导入第五级（村/社区）数据时，将Excel中"城乡分类代码"列的值写入 `urban_rural_code` 字段。

#### Scenario: Level 5 record with urban_rural_code
- **WHEN** 导入第五级区划数据且城乡分类代码非空
- **THEN** `urban_rural_code` 设为该值（如"111"、"112"、"220"等）

#### Scenario: Non-level 5 record
- **WHEN** 导入第一至第四级区划数据
- **THEN** `urban_rural_code` 为 NULL

### Requirement: Build parent-child hierarchy
脚本 SHALL 按层级顺序（level 1→2→3→4→5）插入数据，使用内存中的 code→id 映射构建 `parent_id` 和 `path`。

#### Scenario: Inserting level 1 (province)
- **WHEN** 插入省级数据
- **THEN** `parent_id=NULL`, `path=/省id/`

#### Scenario: Inserting level 2-5
- **WHEN** 插入非省级数据
- **THEN** `parent_id` 通过上级code查找已插入的id，`path` 在上级path后追加当前id

### Requirement: Validate import count per file
脚本 SHALL 在每个文件导入完成后，对比Excel实际数据行数与数据库中对应新增的记录数，记录并输出校验结果。

#### Scenario: Count matches
- **WHEN** 文件导入完成且Excel数据行数等于实际插入行数
- **THEN** 输出 `[OK]` 状态

#### Scenario: Count mismatch
- **WHEN** 文件导入完成且Excel数据行数不等于实际插入行数
- **THEN** 输出 `[MISMATCH]` 警告，记录文件名、预期数量、实际数量

#### Scenario: Duplicate code across files
- **WHEN** 同一个区划code在多个文件中出现
- **THEN** `INSERT IGNORE` 跳过重复，校验时记录为重复条目

### Requirement: Generate import summary report
脚本 SHALL 在所有文件导入完成后输出汇总报告，包含总文件数、成功文件数、失败文件数、总导入记录数、各文件明细。

#### Scenario: All files imported successfully
- **WHEN** 所有文件导入完成
- **THEN** 输出汇总报告，包含每个文件的导入状态和总数

#### Scenario: Some files have mismatches
- **WHEN** 部分文件数量校验不通过
- **THEN** 汇总报告中明确标注不通过的文件及差异详情
