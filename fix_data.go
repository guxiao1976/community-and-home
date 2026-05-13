package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/masterdata_db?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Step 1: Find the parent county of street 284633
	var parentId sql.NullInt64
	err = db.QueryRowContext(ctx, "SELECT parent_id FROM md_administrative_division WHERE id = 284633").Scan(&parentId)
	if err != nil {
		log.Fatalf("Query street 284633 failed: %v", err)
	}

	if !parentId.Valid {
		log.Fatal("Street 284633 has no parent_id")
	}

	fmt.Printf("Street 284633's parent county ID: %d\n", parentId.Int64)

	// Step 2: Count affected records
	var count int64
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM md_residential_area WHERE county_id = 284633").Scan(&count)
	if err != nil {
		log.Fatalf("Count failed: %v", err)
	}

	fmt.Printf("Found %d records with county_id=284633\n", count)

	// Step 3: Update the records
	result, err := db.ExecContext(ctx, `
		UPDATE md_residential_area
		SET county_id = ?, street_id = 284633, updated_time = NOW()
		WHERE county_id = 284633
	`, parentId.Int64)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}

	affected, _ := result.RowsAffected()
	fmt.Printf("Updated %d records: county_id=%d -> %d, street_id=284633\n", affected, 284633, parentId.Int64)

	// Step 4: Verify the update
	var verifyCount int64
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM md_residential_area WHERE county_id = ? AND street_id = 284633", parentId.Int64).Scan(&verifyCount)
	if err != nil {
		log.Fatalf("Verify failed: %v", err)
	}

	fmt.Printf("Verification: %d records now have county_id=%d and street_id=284633\n", verifyCount, parentId.Int64)
}
