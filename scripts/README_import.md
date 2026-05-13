# 敏感词导入脚本使用说明

## 功能说明

将 `Sensitive-lexicon/Vocabulary/` 目录下的词库文件批量导入到 `masterdata_db.md_sensitive_word` 表中。

- **文件名 → category 字段**：例如 `政治类型.txt` → category = `政治类型`
- **每行一个敏感词**：自动去除空行和首尾空格
- **批量插入**：默认每批 1000 条，提高性能
- **去重处理**：使用 `ON DUPLICATE KEY UPDATE`，重复词只更新时间

## 使用方法

### 1. 预览模式（推荐先运行）

```bash
cd scripts
go run import_sensitive_words.go -dry-run
```

### 2. 正式导入

```bash
go run import_sensitive_words.go
```

### 3. 清空后重新导入

```bash
go run import_sensitive_words.go -clear
```

## 参数说明

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-dsn` | `root:123456@tcp(127.0.0.1:3306)/masterdata_db?charset=utf8mb4&parseTime=True&loc=Local` | MySQL 连接字符串 |
| `-dir` | `./Sensitive-lexicon/Vocabulary` | 词库目录路径 |
| `-created-by` | `1` | 创建人 ID |
| `-batch` | `1000` | 批量插入大小 |
| `-dry-run` | `false` | 仅预览，不实际导入 |
| `-clear` | `false` | 导入前清空表数据 |

## 示例

### 自定义数据库连接

```bash
go run import_sensitive_words.go \
  -dsn "user:pass@tcp(192.168.1.100:3306)/masterdata_db?charset=utf8mb4&parseTime=True&loc=Local"
```

### 指定其他目录

```bash
go run import_sensitive_words.go \
  -dir "/path/to/other/vocabulary"
```

### 完整参数示例

```bash
go run import_sensitive_words.go \
  -dsn "root:123456@tcp(127.0.0.1:3306)/masterdata_db?charset=utf8mb4&parseTime=True&loc=Local" \
  -dir "./Sensitive-lexicon/Vocabulary" \
  -created-by 1 \
  -batch 500 \
  -clear
```

## 输出示例

```
✓ 数据库连接成功
✓ [COVID-19词库] 导入 82 个敏感词
✓ [GFW补充词库] 导入 4521 个敏感词
✓ [其他词库] 导入 68 个敏感词
✓ [反动词库] 导入 295 个敏感词
✓ [广告类型] 导入 62 个敏感词
✓ [政治类型] 导入 150 个敏感词
✓ [暴恐词库] 导入 97 个敏感词
✓ [色情词库] 导入 380 个敏感词

========== 导入完成 ==========
文件数: 17
敏感词总数: 65338
```

## 注意事项

1. **首次导入建议先用 `-dry-run` 预览**
2. **word 字段有唯一索引**，重复词会自动跳过（更新时间）
3. **默认字段值**：
   - `severity` = 1 (低危)
   - `action` = 1 (警告)
   - `status` = 1 (启用)
   - `submission_status` = 1 (已通过)
   - `word_type` = 1 (黑名单)
4. **需要 go-sql-driver/mysql 依赖**：
   ```bash
   go get github.com/go-sql-driver/mysql
   ```

## 词库文件格式

```
敏感词1
敏感词2
敏感词3
```

- 每行一个敏感词
- 无序号前缀
- 自动忽略空行
