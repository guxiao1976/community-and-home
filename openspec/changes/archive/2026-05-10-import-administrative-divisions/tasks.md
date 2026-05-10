## 1. Database Schema

- [x] 1.1 ALTER TABLE `md_administrative_division` ADD COLUMN `urban_rural_code VARCHAR(20) NULL` AFTER `sort_order`（脚本内自动执行，已存在则跳过）

## 2. Import Script

- [x] 2.1 编写 Python 导入脚本 `scripts/import_admin_divisions.py`，包含数据库连接、文件扫描（排除 `~$` 文件）、Excel 解析（跳过表头和统计行）、五级数据提取
- [x] 2.2 实现分层去重插入逻辑：按 level 1→2→3→4→5 顺序，使用 `INSERT IGNORE` + code→id 内存映射构建 parent_id 和 path
- [x] 2.3 实现字段赋值：`submission_status=2`, `status=1`，第五级写入 `urban_rural_code`
- [x] 2.4 实现每文件数量校验：对比 Excel 数据行数 vs 实际插入行数，输出 `[OK]` 或 `[MISMATCH]` 状态
- [x] 2.5 实现汇总报告：输出总文件数、成功/失败数、总导入记录数、各文件明细及差异

## 3. Execute and Verify

- [x] 3.1 执行导入脚本，清空表并导入全部31个省级文件数据
- [x] 3.2 验证汇总报告，确认所有文件数量校验通过（或明确记录差异原因）
- [x] 3.3 抽样验证数据库：检查五级层级关系正确性、urban_rural_code 仅第五级有值、submission_status=2 且 status=1
