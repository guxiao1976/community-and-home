package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type ConversionStats struct {
	Total        int
	Success      int
	Skipped      int
	Errors       int
	ErrorSamples []string
	StartTime    time.Time
}

type VillageRecord struct {
	ID       int64
	Code     string
	Name     string
	ParentID sql.NullInt64
	Level    int64
}

type HierarchicalIDs struct {
	ProvinceID     int64
	CityID         int64
	CountyID       int64
	StreetID       sql.NullInt64
	CommunityDivID sql.NullInt64
}

type ResidentialAreaInsert struct {
	Code           string
	Name           string
	CountyID       int64
	StreetID       sql.NullInt64
	CommunityDivID sql.NullInt64
}

func main() {
	// Fix #1: Use environment variable with fallback for credentials
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:123456@tcp(localhost:3306)/masterdata_db?parseTime=true&loc=Local"
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Fix #3: Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Starting village committee conversion...")

	stats := &ConversionStats{
		ErrorSamples: make([]string, 0, 10),
		StartTime:    time.Now(),
	}

	// Fix #2: Add context timeout with cancel
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := convertVillages(ctx, db, stats); err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	duration := time.Since(stats.StartTime)

	log.Printf("\n=== Conversion Summary ===")
	log.Printf("Total processed: %d", stats.Total)
	log.Printf("Successfully converted: %d", stats.Success)
	log.Printf("Skipped (duplicates): %d", stats.Skipped)
	log.Printf("Errors: %d", stats.Errors)
	log.Printf("Duration: %s", duration.Round(time.Second))
	if duration.Seconds() > 0 {
		log.Printf("Average rate: %.0f records/second", float64(stats.Total)/duration.Seconds())
	} else {
		log.Printf("Average rate: N/A (completed too quickly)")
	}
	if len(stats.ErrorSamples) > 0 {
		log.Printf("Sample errors:")
		for _, errMsg := range stats.ErrorSamples {
			log.Printf("  - %s", errMsg)
		}
	}
}

func convertVillages(ctx context.Context, db *sql.DB, stats *ConversionStats) error {
	const batchSize = 1000
	const insertBatchSize = 100 // Insert 100 records at a time
	offset := 0

	for {
		villages, err := extractVillageCommittees(ctx, db, offset, batchSize)
		if err != nil {
			return fmt.Errorf("extract villages at offset %d: %w", offset, err)
		}

		if len(villages) == 0 {
			break
		}

		log.Printf("Processing batch: offset=%d, count=%d", offset, len(villages))
		stats.Total += len(villages)

		pendingInserts := make([]ResidentialAreaInsert, 0, insertBatchSize)

		// Process each village
		for _, village := range villages {
			ids, err := parseHierarchicalIDs(ctx, db, village.ID)
			if err != nil {
				errMsg := fmt.Sprintf("Village %d: %v", village.ID, err)
				log.Printf("Failed to parse hierarchy: %s", errMsg)
				stats.Errors++
				if len(stats.ErrorSamples) < 10 {
					stats.ErrorSamples = append(stats.ErrorSamples, errMsg)
				}
				continue
			}

			transformedName := transformVillageName(village.Name)

			code, err := generateResidentialAreaCode(ctx, db, village.Code)
			if err != nil {
				errMsg := fmt.Sprintf("Village %d: failed to generate code: %v", village.ID, err)
				log.Printf("%s", errMsg)
				stats.Errors++
				if len(stats.ErrorSamples) < 10 {
					stats.ErrorSamples = append(stats.ErrorSamples, errMsg)
				}
				continue
			}

			// Check for duplicates
			isDuplicate, err := checkDuplicateResidentialArea(ctx, db, transformedName, ids.CountyID)
			if err != nil {
				errMsg := fmt.Sprintf("Village %d: failed to check duplicate: %v", village.ID, err)
				log.Printf("%s", errMsg)
				stats.Errors++
				if len(stats.ErrorSamples) < 10 {
					stats.ErrorSamples = append(stats.ErrorSamples, errMsg)
				}
				continue
			}

			if isDuplicate {
				stats.Skipped++
				continue
			}

			// Add to pending batch
			pendingInserts = append(pendingInserts, ResidentialAreaInsert{
				Code:           code,
				Name:           transformedName,
				CountyID:       ids.CountyID,
				StreetID:       ids.StreetID,
				CommunityDivID: ids.CommunityDivID,
			})

			// Insert when batch is full
			if len(pendingInserts) >= insertBatchSize {
				if err := insertResidentialAreaBatch(ctx, db, pendingInserts); err != nil {
					errMsg := fmt.Sprintf("Batch insert failed: %v", err)
					log.Printf("%s", errMsg)
					stats.Errors += len(pendingInserts)
					if len(stats.ErrorSamples) < 10 {
						stats.ErrorSamples = append(stats.ErrorSamples, errMsg)
					}
				} else {
					stats.Success += len(pendingInserts)
				}
				pendingInserts = pendingInserts[:0] // Clear batch
			}
		}

		// Insert remaining records
		if len(pendingInserts) > 0 {
			if err := insertResidentialAreaBatch(ctx, db, pendingInserts); err != nil {
				errMsg := fmt.Sprintf("Batch insert failed: %v", err)
				log.Printf("%s", errMsg)
				stats.Errors += len(pendingInserts)
				if len(stats.ErrorSamples) < 10 {
					stats.ErrorSamples = append(stats.ErrorSamples, errMsg)
				}
			} else {
				stats.Success += len(pendingInserts)
			}
		}

		// Progress logging
		elapsed := time.Since(stats.StartTime)
		if elapsed.Seconds() > 0 {
			rate := float64(stats.Total) / elapsed.Seconds()
			remaining := 611669 - stats.Total
			if rate > 0 && remaining > 0 {
				eta := time.Duration(float64(remaining)/rate) * time.Second
				log.Printf("Progress: %d/%d (%.1f%%) | Success: %d | Skipped: %d | Errors: %d | Rate: %.0f/s | ETA: %s",
					stats.Total, 611669, float64(stats.Total)/611669*100,
					stats.Success, stats.Skipped, stats.Errors,
					rate, eta.Round(time.Second))
			} else {
				log.Printf("Progress: %d/%d (%.1f%%) | Success: %d | Skipped: %d | Errors: %d",
					stats.Total, 611669, float64(stats.Total)/611669*100,
					stats.Success, stats.Skipped, stats.Errors)
			}
		}

		offset += batchSize
	}

	return nil
}

func extractVillageCommittees(ctx context.Context, db *sql.DB, offset, limit int) ([]VillageRecord, error) {
	query := `
		SELECT id, code, name, parent_id, level
		FROM md_administrative_division
		WHERE level = 5 AND delete_time IS NULL AND submission_status != 4
		ORDER BY id
		LIMIT ? OFFSET ?
	`

	rows, err := db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query villages failed: %w", err)
	}
	defer rows.Close()

	var villages []VillageRecord
	for rows.Next() {
		var v VillageRecord
		if err := rows.Scan(&v.ID, &v.Code, &v.Name, &v.ParentID, &v.Level); err != nil {
			return nil, fmt.Errorf("scan village failed: %w", err)
		}
		villages = append(villages, v)
	}

	return villages, rows.Err()
}

func parseHierarchicalIDs(ctx context.Context, db *sql.DB, villageID int64) (*HierarchicalIDs, error) {
	query := `
		WITH RECURSIVE hierarchy AS (
			SELECT id, parent_id, level, 1 as depth
			FROM md_administrative_division
			WHERE id = ?

			UNION ALL

			SELECT d.id, d.parent_id, d.level, h.depth + 1
			FROM md_administrative_division d
			INNER JOIN hierarchy h ON d.id = h.parent_id
			WHERE h.depth < 10
		)
		SELECT id, level FROM hierarchy ORDER BY level
	`

	rows, err := db.QueryContext(ctx, query, villageID)
	if err != nil {
		return nil, fmt.Errorf("query hierarchy failed: %w", err)
	}
	defer rows.Close()

	ids := &HierarchicalIDs{}
	for rows.Next() {
		var id, level int64
		if err := rows.Scan(&id, &level); err != nil {
			return nil, fmt.Errorf("scan hierarchy failed: %w", err)
		}

		switch level {
		case 1:
			ids.ProvinceID = id
		case 2:
			ids.CityID = id
		case 3:
			ids.CountyID = id
		case 4:
			ids.StreetID = sql.NullInt64{Int64: id, Valid: true}
		// Note: Level 5 (village committee) is the entity being converted,
		// not a parent community division, so we don't store it in CommunityDivID
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate hierarchy rows: %w", err)
	}

	// Validate required fields
	if ids.ProvinceID == 0 || ids.CityID == 0 || ids.CountyID == 0 {
		return nil, fmt.Errorf("incomplete hierarchy: province=%d, city=%d, county=%d",
			ids.ProvinceID, ids.CityID, ids.CountyID)
	}

	return ids, nil
}

func transformVillageName(originalName string) string {
	// Remove common suffixes for village committees
	suffixes := []string{"社区居委会", "村委会", "居委会"}

	name := originalName
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			name = strings.TrimSuffix(name, suffix)
			break
		}
	}

	// If name is empty after trimming, use original
	name = strings.TrimSpace(name)
	if name == "" {
		return originalName
	}

	return name
}

func checkDuplicateResidentialArea(ctx context.Context, db *sql.DB, name string, countyID int64) (bool, error) {
	query := `
		SELECT COUNT(*) FROM md_residential_area
		WHERE name = ? AND county_id = ? AND delete_time IS NULL
	`

	var count int
	err := db.QueryRowContext(ctx, query, name, countyID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check duplicate failed: %w", err)
	}

	return count > 0, nil
}

func insertResidentialAreaBatch(ctx context.Context, db *sql.DB, records []ResidentialAreaInsert) error {
	if len(records) == 0 {
		return nil
	}

	// Build multi-value INSERT
	query := `INSERT INTO md_residential_area (
		code, name, county_id, street_id, community_div_id,
		community_type, submission_status, address, submitter_id
	) VALUES `

	values := []interface{}{}
	for i, record := range records {
		if i > 0 {
			query += ", "
		}
		query += "(?, ?, ?, ?, ?, ?, ?, ?, ?)"
		values = append(values,
			record.Code,
			record.Name,
			record.CountyID,
			record.StreetID,
			record.CommunityDivID,
			2, // community_type: village
			1, // submission_status: approved
			record.Name, // address placeholder
			1, // submitter_id: system
		)
	}

	_, err := db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("batch insert failed: %w", err)
	}

	return nil
}

// generateResidentialAreaCode generates a unique code for a residential area.
// Format: villageCode (12 digits) + sequential suffix (3 digits, e.g., 001, 002, ...)
// Note: This function is not concurrency-safe. For single-threaded migration use only.
// For concurrent use, implement database-level locking or rely on unique constraint retry logic.
func generateResidentialAreaCode(ctx context.Context, db *sql.DB, villageCode string) (string, error) {
	// Validate village code length (should be 12 digits)
	if len(villageCode) != 12 {
		return "", fmt.Errorf("invalid village code length: expected 12, got %d", len(villageCode))
	}

	// Query for existing codes with this prefix
	query := `
		SELECT code FROM md_residential_area
		WHERE code LIKE ? AND delete_time IS NULL
		ORDER BY code DESC
		LIMIT 1
	`

	var lastCode sql.NullString
	err := db.QueryRowContext(ctx, query, villageCode+"%").Scan(&lastCode)
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("query last code failed: %w", err)
	}

	// If no existing codes, start with 001
	if !lastCode.Valid || lastCode.String == "" {
		return villageCode + "001", nil
	}

	// Extract the suffix and increment
	if len(lastCode.String) <= len(villageCode) {
		log.Printf("Warning: Village code %s has malformed existing code %s, starting fresh with 001", villageCode, lastCode.String)
		return villageCode + "001", nil
	}

	suffix := lastCode.String[len(villageCode):]
	var seqNum int
	_, err = fmt.Sscanf(suffix, "%d", &seqNum)
	if err != nil {
		// If suffix is not a number, start fresh
		log.Printf("Warning: Village code %s has non-numeric suffix in %s, starting fresh with 001", villageCode, lastCode.String)
		return villageCode + "001", nil
	}

	// Increment and format with leading zeros
	seqNum++
	return fmt.Sprintf("%s%03d", villageCode, seqNum), nil
}
