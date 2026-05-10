#!/usr/bin/env python3
"""
回填 md_residential_area 表的 city_id 字段
从 county_id 对应的行政区划 path 中提取第2段（城市id）作为 city_id
"""

import pymysql

DB_CONFIG = {
    "host": "localhost",
    "port": 3306,
    "user": "root",
    "password": "123456",
    "database": "masterdata_db",
    "charset": "utf8mb4",
}

BATCH_SIZE = 5000

def main():
    conn = pymysql.connect(**DB_CONFIG, autocommit=False)
    cur = conn.cursor()

    # Count records with null city_id
    cur.execute("SELECT COUNT(*) FROM md_residential_area WHERE city_id IS NULL")
    total = cur.fetchone()[0]
    print(f"待回填记录数: {total}")

    if total == 0:
        print("无需回填")
        conn.close()
        return

    # For level=3 county: path=/省id/市id/县id/ -> city_id = segment 2
    # For level=2 county: path=/省id/市id/ -> city_id = segment 2 (itself)
    # Use a single UPDATE with JOIN (no LIMIT needed for one-shot)
    update_sql = """
        UPDATE md_residential_area ra
        JOIN md_administrative_division d ON ra.county_id = d.id
        SET ra.city_id = CAST(SUBSTRING_INDEX(SUBSTRING_INDEX(d.path, '/', 3), '/', -1) AS UNSIGNED)
        WHERE ra.city_id IS NULL
    """

    print("执行回填...")
    cur.execute(update_sql)
    updated = cur.rowcount
    conn.commit()
    print(f"回填完成: 共更新 {updated} 条")

    # Verify
    cur.execute("SELECT COUNT(*) FROM md_residential_area WHERE city_id IS NULL")
    remaining = cur.fetchone()[0]
    print(f"\n回填完成: 共更新 {updated} 条, 剩余空值 {remaining} 条")

    # Sample check
    cur.execute("""
        SELECT ra.id, ra.name, ra.city_id, c.name AS city_name
        FROM md_residential_area ra
        JOIN md_administrative_division c ON ra.city_id = c.id
        WHERE ra.city_id IS NOT NULL
        LIMIT 5
    """)
    print("抽样验证:")
    for r in cur.fetchall():
        print(f"  id={r[0]}, name={r[1]}, city_id={r[2]}, city_name={r[3]}")

    conn.close()

if __name__ == "__main__":
    main()
