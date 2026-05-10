#!/usr/bin/env python3
"""
行政区划五级数据批量导入脚本
从 /mnt/d/wsl/data/行政区划/ 目录读取省级Excel文件，导入到 md_administrative_division 表
"""

import os
import sys
import glob
import time
import openpyxl
import pymysql

DATA_DIR = "/mnt/d/wsl/data/行政区划/"
DB_CONFIG = {
    "host": "localhost",
    "port": 3306,
    "user": "root",
    "password": "123456",
    "database": "masterdata_db",
    "charset": "utf8mb4",
}

BATCH_SIZE = 2000


def get_connection():
    return pymysql.connect(**DB_CONFIG, autocommit=False)


def ensure_urban_rural_column(conn):
    """Add urban_rural_code column if not exists."""
    try:
        with conn.cursor() as cur:
            cur.execute(
                "ALTER TABLE md_administrative_division "
                "ADD COLUMN urban_rural_code VARCHAR(20) NULL AFTER sort_order"
            )
        conn.commit()
        print("[OK] Added urban_rural_code column")
    except pymysql.err.OperationalError as e:
        conn.rollback()
        if e.args[0] == 1060:  # Duplicate column
            print("[SKIP] urban_rural_code column already exists")
        else:
            raise
    except Exception:
        conn.rollback()
        raise


def truncate_table(conn):
    """Truncate md_administrative_division table, disabling FK checks temporarily."""
    with conn.cursor() as cur:
        cur.execute("SET FOREIGN_KEY_CHECKS = 0")
        cur.execute("TRUNCATE TABLE md_administrative_division")
        cur.execute("SET FOREIGN_KEY_CHECKS = 1")
    conn.commit()
    print("[OK] Table truncated")


def scan_excel_files(data_dir):
    """Scan for xlsx files, excluding temp files."""
    files = []
    for f in sorted(glob.glob(os.path.join(data_dir, "*.xlsx"))):
        basename = os.path.basename(f)
        if basename.startswith("~$"):
            continue
        files.append(f)
    return files


def parse_excel(filepath):
    """Parse one Excel file, return list of 5-level rows.

    Each row: (prov_code, prov_name, city_code, city_name,
               dist_code, dist_name, town_code, town_name,
               vill_code, urban_rural_code, vill_name)
    """
    wb = openpyxl.load_workbook(filepath, data_only=True, read_only=True)
    ws = wb.active
    rows = []
    for i, row in enumerate(ws.iter_rows(min_row=3, values_only=True)):
        if row[0] is None:
            continue
        rows.append(row)
    wb.close()
    return rows


def build_unique_divisions(all_rows):
    """Extract unique divisions by level from raw rows.

    Returns: dict[int, list[tuple]]  level -> [(code, name, parent_code, urban_rural_code), ...]
    """
    levels = {1: {}, 2: {}, 3: {}, 4: {}, 5: {}}

    for row in all_rows:
        prov_code, prov_name = row[0], row[1]
        city_code, city_name = row[2], row[3]
        dist_code, dist_name = row[4], row[5]
        town_code, town_name = row[6], row[7]
        vill_code, urban_rural = row[8], row[9]
        vill_name = row[10]

        if prov_code is not None:
            levels[1][str(prov_code)] = (str(prov_code), prov_name, None, None)
        if city_code is not None:
            levels[2][str(city_code)] = (str(city_code), city_name, str(prov_code), None)
        if dist_code is not None:
            levels[3][str(dist_code)] = (str(dist_code), dist_name, str(city_code), None)
        if town_code is not None:
            levels[4][str(town_code)] = (str(town_code), town_name, str(dist_code), None)
        if vill_code is not None:
            ur_code = str(urban_rural) if urban_rural is not None else None
            levels[5][str(vill_code)] = (str(vill_code), vill_name, str(town_code), ur_code)

    return {lvl: list(divs.values()) for lvl, divs in levels.items()}


def insert_divisions(conn, divisions_by_level, code_to_id, id_to_path):
    """Insert divisions level by level, return total_inserted.

    Uses in-memory code_to_id and id_to_path maps to avoid per-row queries.
    All records get path = '/' as placeholder first; paths are corrected
    in a single UPDATE pass per level after batch insert.
    """
    total_inserted = 0

    with conn.cursor() as cur:
        for level in range(1, 6):
            divs = divisions_by_level.get(level, [])
            if not divs:
                continue

            batch = []
            for div in divs:
                code, name, parent_code, urban_rural = div
                parent_id = code_to_id.get(parent_code) if parent_code else None
                batch.append((parent_id, level, name, code, "/", urban_rural))

            # Batch insert
            cur.executemany(
                "INSERT IGNORE INTO md_administrative_division "
                "(parent_id, level, name, code, path, sort_order, status, "
                "submission_status, urban_rural_code, created_by) "
                "VALUES (%s, %s, %s, %s, %s, 0, 1, 2, %s, 0)",
                batch
            )
            conn.commit()

            # Retrieve newly inserted rows to build mapping and fix paths
            codes_this_level = [div[0] for div in divs]
            placeholders = ",".join(["%s"] * len(codes_this_level))
            cur.execute(
                f"SELECT id, code, parent_id FROM md_administrative_division "
                f"WHERE code IN ({placeholders})",
                codes_this_level
            )
            new_rows = cur.fetchall()

            # Update paths in batch
            for row_id, code, parent_id in new_rows:
                code_to_id[code] = row_id
                if parent_id is not None and parent_id in id_to_path:
                    new_path = id_to_path[parent_id] + f"{row_id}/"
                else:
                    new_path = f"/{row_id}/"
                id_to_path[row_id] = new_path
                cur.execute(
                    "UPDATE md_administrative_division SET path = %s WHERE id = %s",
                    (new_path, row_id)
                )
                total_inserted += 1

            conn.commit()

    return total_inserted


def main():
    print("=" * 60)
    print("行政区划五级数据导入工具")
    print("=" * 60)

    conn = get_connection()

    # Task 1.1: Ensure urban_rural_code column
    print("\n[Step 1] 检查/添加 urban_rural_code 列...")
    ensure_urban_rural_column(conn)

    # Truncate table
    print("\n[Step 2] 清空 md_administrative_division 表...")
    truncate_table(conn)

    # Scan files
    print(f"\n[Step 3] 扫描文件目录: {DATA_DIR}")
    files = scan_excel_files(DATA_DIR)
    print(f"  找到 {len(files)} 个Excel文件")

    if not files:
        print("[ERROR] 未找到任何Excel文件，退出")
        conn.close()
        return

    # Process files
    print(f"\n[Step 4] 开始导入数据...\n")
    results = []
    total_inserted = 0
    total_expected = 0
    code_to_id = {}
    id_to_path = {}
    mismatch_count = 0

    for i, filepath in enumerate(files, 1):
        fname = os.path.basename(filepath)
        print(f"[{i}/{len(files)}] 处理: {fname}")

        try:
            # Parse Excel
            rows = parse_excel(filepath)
            expected = len(rows)
            total_expected += expected

            if expected == 0:
                print(f"  [WARN] 文件无数据行")
                results.append({"file": fname, "expected": 0, "actual": 0, "status": "EMPTY"})
                continue

            # Build unique divisions
            divisions_by_level = build_unique_divisions(rows)
            unique_count = sum(len(divs) for divs in divisions_by_level.values())

            # Insert
            inserted = insert_divisions(conn, divisions_by_level, code_to_id, id_to_path)
            total_inserted += inserted

            # Validate
            if inserted == unique_count:
                status = "OK"
            else:
                status = "MISMATCH"
                mismatch_count += 1

            results.append({
                "file": fname,
                "expected": unique_count,
                "actual": inserted,
                "data_rows": expected,
                "status": status,
            })
            print(f"  数据行: {expected}, 去重后: {unique_count}, 插入: {inserted} [{status}]")

        except Exception as e:
            print(f"  [ERROR] {e}")
            results.append({"file": fname, "expected": -1, "actual": -1, "status": "ERROR"})
            mismatch_count += 1

    # Summary report
    print("\n" + "=" * 60)
    print("导入汇总报告")
    print("=" * 60)
    ok_count = sum(1 for r in results if r["status"] == "OK")
    print(f"总文件数: {len(files)}")
    print(f"成功: {ok_count}")
    print(f"异常/差异: {mismatch_count}")
    print(f"总数据行: {total_expected}")
    print(f"总去重后记录: {sum(r.get('expected', 0) for r in results if r['status'] != 'ERROR')}")
    print(f"总实际插入: {total_inserted}")

    print(f"\n--- 各文件明细 ---")
    for r in results:
        status_tag = r["status"]
        if status_tag == "OK":
            print(f"  [OK] {r['file']}: {r['expected']} 条")
        elif status_tag == "MISMATCH":
            print(f"  [MISMATCH] {r['file']}: 预期 {r['expected']}, 实际 {r['actual']}")
        elif status_tag == "ERROR":
            print(f"  [ERROR] {r['file']}: 导入失败")
        else:
            print(f"  [EMPTY] {r['file']}: 无数据")

    conn.close()
    print("\n完成!")


if __name__ == "__main__":
    main()
