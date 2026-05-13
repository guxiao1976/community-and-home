package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/masterdata_db?charset=utf8mb4&parseTime=true")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 检查东营市数据的county_id分布
	fmt.Println("=== 东营市数据的county_id分布 ===")
	rows, err := db.QueryContext(context.Background(),
		"SELECT county_id, COUNT(*) as count FROM md_residential_area WHERE city_id = 283993 GROUP BY county_id ORDER BY count DESC")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var countyId sql.NullInt64
		var count int
		rows.Scan(&countyId, &count)
		if countyId.Valid {
			fmt.Printf("county_id: %d, count: %d\n", countyId.Int64, count)
		} else {
			fmt.Printf("county_id: NULL, count: %d\n", count)
		}
	}

	// 检查东营市的行政区划
	fmt.Println("\n=== 东营市的下级区县 (parent_id=283993) ===")
	rows2, err := db.QueryContext(context.Background(),
		"SELECT id, name, code FROM md_administrative_division WHERE parent_id = 283993 ORDER BY id")
	if err != nil {
		panic(err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var id int64
		var name, code string
		rows2.Scan(&id, &name, &code)
		fmt.Printf("id: %d, name: %s, code: %s\n", id, name, code)
	}

	// 检查东营市数据样例
	fmt.Println("\n=== 东营市数据样例 (前5条) ===")
	rows3, err := db.QueryContext(context.Background(),
		"SELECT id, name, county_id, city_id, community_type FROM md_residential_area WHERE city_id = 283993 LIMIT 5")
	if err != nil {
		panic(err)
	}
	defer rows3.Close()

	for rows3.Next() {
		var id int64
		var name string
		var countyId, cityId sql.NullInt64
		var communityType int
		rows3.Scan(&id, &name, &countyId, &cityId, &communityType)
		fmt.Printf("id: %d, name: %s, county_id: %v, city_id: %v, community_type: %d\n",
			id, name, countyId, cityId, communityType)
	}
}
