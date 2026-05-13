package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	DSN         string
	VocabDir    string
	CreatedBy   int64
	BatchSize   int
	DryRun      bool
	ClearBefore bool
}

func main() {
	cfg := Config{}
	flag.StringVar(&cfg.DSN, "dsn", "root:123456@tcp(127.0.0.1:3306)/masterdata_db?charset=utf8mb4&parseTime=True&loc=Local", "MySQL DSN")
	flag.StringVar(&cfg.VocabDir, "dir", "../Sensitive-lexicon/Vocabulary", "词库目录路径")
	flag.Int64Var(&cfg.CreatedBy, "created-by", 1, "创建人ID")
	flag.IntVar(&cfg.BatchSize, "batch", 1000, "批量插入大小")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "仅预览，不实际导入")
	flag.BoolVar(&cfg.ClearBefore, "clear", false, "导入前清空表数据")
	flag.Parse()

	if err := run(cfg); err != nil {
		log.Fatalf("导入失败: %v", err)
	}
}

func run(cfg Config) error {
	// 连接数据库
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	log.Println("✓ 数据库连接成功")

	// 清空表（如果需要）
	if cfg.ClearBefore && !cfg.DryRun {
		if _, err := db.Exec("TRUNCATE TABLE md_sensitive_word"); err != nil {
			return fmt.Errorf("清空表失败: %w", err)
		}
		log.Println("✓ 已清空表数据")
	}

	// 读取词库目录
	files, err := os.ReadDir(cfg.VocabDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	totalWords := 0
	totalFiles := 0

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}

		category := strings.TrimSuffix(file.Name(), ".txt")
		filePath := filepath.Join(cfg.VocabDir, file.Name())

		count, err := importFile(db, filePath, category, cfg)
		if err != nil {
			log.Printf("✗ 导入文件 %s 失败: %v", file.Name(), err)
			continue
		}

		totalWords += count
		totalFiles++
		log.Printf("✓ [%s] 导入 %d 个敏感词", category, count)
	}

	log.Printf("\n========== 导入完成 ==========")
	log.Printf("文件数: %d", totalFiles)
	log.Printf("敏感词总数: %d", totalWords)
	if cfg.DryRun {
		log.Printf("(预览模式，未实际写入数据库)")
	}

	return nil
}

func importFile(db *sql.DB, filePath, category string, cfg Config) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// 读取所有敏感词
	words := make([]string, 0)
	skipped := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word == "" {
			continue
		}
		// 跳过超过100字符的词（数据库字段限制）
		if len(word) > 100 {
			skipped++
			continue
		}
		words = append(words, word)
	}

	if skipped > 0 {
		log.Printf("  ⚠ 跳过 %d 个超长敏感词（>100字符）", skipped)
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}

	if len(words) == 0 {
		return 0, nil
	}

	// 预览模式
	if cfg.DryRun {
		return len(words), nil
	}

	// 批量插入
	now := time.Now()
	inserted := 0

	for i := 0; i < len(words); i += cfg.BatchSize {
		end := i + cfg.BatchSize
		if end > len(words) {
			end = len(words)
		}

		batch := words[i:end]
		if err := insertBatch(db, batch, category, cfg.CreatedBy, now); err != nil {
			return inserted, fmt.Errorf("批量插入失败 (offset=%d): %w", i, err)
		}

		inserted += len(batch)
	}

	return inserted, nil
}

func insertBatch(db *sql.DB, words []string, category string, createdBy int64, now time.Time) error {
	if len(words) == 0 {
		return nil
	}

	// 构建批量插入 SQL
	valueStrings := make([]string, 0, len(words))
	valueArgs := make([]interface{}, 0, len(words)*8)

	for _, word := range words {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			word,      // word
			category,  // category
			1,         // severity (默认1-低危)
			1,         // action (默认1-警告)
			1,         // status (默认1-启用)
			createdBy, // created_by
			now,       // created_time
			now,       // updated_time
		)
	}

	query := fmt.Sprintf(
		"INSERT INTO md_sensitive_word (word, category, severity, action, status, created_by, created_time, updated_time) VALUES %s ON DUPLICATE KEY UPDATE updated_time = VALUES(updated_time), category = VALUES(category)",
		strings.Join(valueStrings, ","),
	)

	_, err := db.Exec(query, valueArgs...)
	return err
}
