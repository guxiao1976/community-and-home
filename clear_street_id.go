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

	// Count records with street_id=284633
	var count int64
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM md_residential_area WHERE street_id = 284633").Scan(&count)
	if err != nil {
		log.Fatalf("Count failed: %v", err)
	}

	fmt.Printf("Found %d records with street_id=284633\n", count)

	if count == 0 {
		fmt.Println("No records to update")
		return
	}

	// Clear the street_id field
	result, err := db.ExecContext(ctx, `
		UPDATE md_residential_area
		SET street_id = NULL, updated_time = NOW()
		WHERE street_id = 284633
	`)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}

	affected, _ := result.RowsAffected()
	fmt.Printf("Cleared street_id for %d records (street_id was incorrectly set during sync)\n", affected)

	// Verify the update
	var verifyCount int64
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM md_residential_area WHERE county_id = 284048 AND street_id IS NULL").Scan(&verifyCount)
	if err != nil {
		log.Fatalf("Verify failed: %v", err)
	}

	fmt.Printf("Verification: %d records now have county_id=284048 and street_id=NULL\n", verifyCount)
}
