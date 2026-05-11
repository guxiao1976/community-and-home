package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MdDivisionStatisticsModel = (*customMdDivisionStatisticsModel)(nil)

type (
	DivisionCountRow struct {
		Id             int64  `db:"id"`
		Name           string `db:"name"`
		Level          int64  `db:"level"`
		CommunityCount int64  `db:"community_count"`
		VillageCount   int64  `db:"village_count"`
		TotalCount     int64  `db:"total_count"`
	}

	MdDivisionStatisticsModel interface {
		mdDivisionStatisticsModel
		DeleteByDate(ctx context.Context, statDate time.Time) error
		InsertBatch(ctx context.Context, rows []MdDivisionStatistics) error
		FindCountsByParentId(ctx context.Context, parentId *int64, statDate time.Time) ([]DivisionCountRow, error)
		FindLatestDate(ctx context.Context) (time.Time, error)
		FindRealtimeCountsByParentId(ctx context.Context, parentId *int64) ([]DivisionCountRow, error)
		RefreshStatistics(ctx context.Context) error
	}

	customMdDivisionStatisticsModel struct {
		*defaultMdDivisionStatisticsModel
	}
)

func NewMdDivisionStatisticsModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MdDivisionStatisticsModel {
	return &customMdDivisionStatisticsModel{
		defaultMdDivisionStatisticsModel: newMdDivisionStatisticsModel(conn, c, opts...),
	}
}

func (m *customMdDivisionStatisticsModel) DeleteByDate(ctx context.Context, statDate time.Time) error {
	query := fmt.Sprintf("delete from %s where stat_date = ?", m.table)
	_, err := m.ExecNoCacheCtx(ctx, query, statDate)
	return err
}

func (m *customMdDivisionStatisticsModel) InsertBatch(ctx context.Context, rows []MdDivisionStatistics) error {
	if len(rows) == 0 {
		return nil
	}
	placeholders := strings.Repeat("(?,?,?,?,?,?,?),", len(rows)-1) + "(?,?,?,?,?,?,?)"
	query := fmt.Sprintf("insert into %s (division_id, level, community_count, village_count, total_count, stat_date, created_at) values %s",
		m.table, placeholders)
	args := make([]interface{}, 0, len(rows)*7)
	for _, r := range rows {
		args = append(args, r.DivisionId, r.Level, r.CommunityCount, r.VillageCount, r.TotalCount, r.StatDate, r.CreatedAt)
	}
	_, err := m.ExecNoCacheCtx(ctx, query, args...)
	return err
}

func (m *customMdDivisionStatisticsModel) FindCountsByParentId(ctx context.Context, parentId *int64, statDate time.Time) ([]DivisionCountRow, error) {
	var query string
	var args []interface{}

	if parentId != nil {
		query = fmt.Sprintf(`SELECT d.id, d.name, d.level,
			COALESCE(s.community_count, 0) as community_count,
			COALESCE(s.village_count, 0) as village_count,
			COALESCE(s.total_count, 0) as total_count
			FROM md_administrative_division d
			LEFT JOIN %s s ON s.division_id = d.id AND s.stat_date = ?
			WHERE d.parent_id = ? AND d.submission_status = 2 AND d.delete_time IS NULL
			ORDER BY COALESCE(s.total_count, 0) DESC`, m.table)
		args = []interface{}{statDate, *parentId}
	} else {
		query = fmt.Sprintf(`SELECT d.id, d.name, d.level,
			COALESCE(s.community_count, 0) as community_count,
			COALESCE(s.village_count, 0) as village_count,
			COALESCE(s.total_count, 0) as total_count
			FROM md_administrative_division d
			LEFT JOIN %s s ON s.division_id = d.id AND s.stat_date = ?
			WHERE d.parent_id IS NULL AND d.level = 1 AND d.submission_status = 2 AND d.delete_time IS NULL
			ORDER BY COALESCE(s.total_count, 0) DESC`, m.table)
		args = []interface{}{statDate}
	}

	var rows []DivisionCountRow
	err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *customMdDivisionStatisticsModel) FindLatestDate(ctx context.Context) (time.Time, error) {
	var row struct {
		StatDate time.Time `db:"stat_date"`
	}
	query := fmt.Sprintf("select stat_date from %s order by stat_date desc limit 1", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &row, query)
	if err != nil {
		return time.Time{}, err
	}
	return row.StatDate, nil
}

func (m *customMdDivisionStatisticsModel) FindRealtimeCountsByParentId(ctx context.Context, parentId *int64) ([]DivisionCountRow, error) {
	// residential_area 关联字段说明：
	//   小区(community_type=1): county_id → level=3(区县), street_id → NULL
	//   村(community_type=2):   county_id → level=2(市),   street_id → level=3(区县)
	// 层级：省(1) → 市(2) → 区县(3) → 街道(4) → 社区(5)
	// 下钻：省 → 市 → 区县 → 街道(无数据)

	var query string
	var args []interface{}

	if parentId == nil {
		// 省级统计：小区按 county_id(区县) → 区县.parent_id(市) → 市.parent_id(省)
		//           村按 street_id(区县) → 区县.parent_id(市) → 市.parent_id(省)
		query = `SELECT p.id, p.name, p.level,
			COALESCE(sub.community_count, 0) AS community_count,
			COALESCE(sub.village_count, 0) AS village_count,
			COALESCE(sub.total_count, 0) AS total_count
			FROM md_administrative_division p
			LEFT JOIN (
				SELECT city.parent_id AS province_id,
					SUM(t.community_count) AS community_count,
					SUM(t.village_count) AS village_count,
					SUM(t.total_count) AS total_count
				FROM (
					SELECT ct.parent_id AS city_id,
						SUM(ra.community_count) AS community_count,
						SUM(ra.village_count) AS village_count,
						SUM(ra.total_count) AS total_count
					FROM (
						SELECT r.county_id AS div_id,
							COUNT(CASE WHEN r.community_type = 1 THEN 1 END) AS community_count,
							COUNT(CASE WHEN r.community_type = 2 THEN 1 END) AS village_count,
							COUNT(*) AS total_count
						FROM md_residential_area r
						WHERE r.submission_status = 2 AND r.delete_time IS NULL AND r.community_type = 1
						GROUP BY r.county_id
						UNION ALL
						SELECT r.street_id AS div_id,
							COUNT(CASE WHEN r.community_type = 1 THEN 1 END) AS community_count,
							COUNT(CASE WHEN r.community_type = 2 THEN 1 END) AS village_count,
							COUNT(*) AS total_count
						FROM md_residential_area r
						WHERE r.submission_status = 2 AND r.delete_time IS NULL AND r.community_type = 2
						GROUP BY r.street_id
					) ra
					INNER JOIN md_administrative_division ct ON ct.id = ra.div_id
						AND ct.level = 3 AND ct.submission_status = 2 AND ct.delete_time IS NULL
					GROUP BY ct.parent_id
				) t
				INNER JOIN md_administrative_division city ON city.id = t.city_id
					AND city.level = 2 AND city.submission_status = 2 AND city.delete_time IS NULL
				GROUP BY city.parent_id
			) sub ON sub.province_id = p.id
			WHERE p.level = 1 AND p.parent_id IS NULL AND p.submission_status = 2 AND p.delete_time IS NULL
			ORDER BY COALESCE(sub.total_count, 0) DESC`
	} else {
		var parent struct {
			Level int64 `db:"level"`
		}
		parentErr := m.QueryRowNoCacheCtx(ctx, &parent,
			"SELECT level FROM md_administrative_division WHERE id = ? LIMIT 1", *parentId)
		if parentErr != nil {
			return nil, parentErr
		}

		switch parent.Level {
		case 1:
			// 省 → 市：小区按 county_id(区县) → 区县.parent_id(市)；村按 county_id(市) 直接聚合
			query = `SELECT d.id, d.name, d.level,
				COALESCE(sub.community_count, 0) AS community_count,
				COALESCE(sub.village_count, 0) AS village_count,
				COALESCE(sub.total_count, 0) AS total_count
				FROM md_administrative_division d
				LEFT JOIN (
					SELECT city_id,
						SUM(community_count) AS community_count,
						SUM(village_count) AS village_count,
						SUM(total_count) AS total_count
					FROM (
						SELECT ct.parent_id AS city_id,
							COUNT(CASE WHEN r.community_type = 1 THEN 1 END) AS community_count,
							COUNT(CASE WHEN r.community_type = 2 THEN 1 END) AS village_count,
							COUNT(*) AS total_count
						FROM md_residential_area r
						INNER JOIN md_administrative_division ct ON ct.id = r.county_id
							AND ct.level = 3 AND ct.submission_status = 2 AND ct.delete_time IS NULL
						WHERE r.submission_status = 2 AND r.delete_time IS NULL AND r.community_type = 1
						GROUP BY ct.parent_id
						UNION ALL
						SELECT r.county_id AS city_id,
							COUNT(CASE WHEN r.community_type = 1 THEN 1 END) AS community_count,
							COUNT(CASE WHEN r.community_type = 2 THEN 1 END) AS village_count,
							COUNT(*) AS total_count
						FROM md_residential_area r
						WHERE r.submission_status = 2 AND r.delete_time IS NULL AND r.community_type = 2
						GROUP BY r.county_id
					) combined
					GROUP BY city_id
				) sub ON sub.city_id = d.id
				WHERE d.parent_id = ? AND d.submission_status = 2 AND d.delete_time IS NULL
				ORDER BY COALESCE(sub.total_count, 0) DESC`
			args = []interface{}{*parentId}

		case 2:
			// 市 → 区县：小区按 county_id(区县) 聚合；村按 street_id(区县) 聚合
			query = `SELECT d.id, d.name, d.level,
				COALESCE(sub.community_count, 0) AS community_count,
				COALESCE(sub.village_count, 0) AS village_count,
				COALESCE(sub.total_count, 0) AS total_count
				FROM md_administrative_division d
				LEFT JOIN (
					SELECT div_id,
						SUM(community_count) AS community_count,
						SUM(village_count) AS village_count,
						SUM(total_count) AS total_count
					FROM (
						SELECT r.county_id AS div_id,
							COUNT(CASE WHEN r.community_type = 1 THEN 1 END) AS community_count,
							COUNT(CASE WHEN r.community_type = 2 THEN 1 END) AS village_count,
							COUNT(*) AS total_count
						FROM md_residential_area r
						WHERE r.submission_status = 2 AND r.delete_time IS NULL AND r.community_type = 1
						GROUP BY r.county_id
						UNION ALL
						SELECT r.street_id AS div_id,
							COUNT(CASE WHEN r.community_type = 1 THEN 1 END) AS community_count,
							COUNT(CASE WHEN r.community_type = 2 THEN 1 END) AS village_count,
							COUNT(*) AS total_count
						FROM md_residential_area r
						WHERE r.submission_status = 2 AND r.delete_time IS NULL AND r.community_type = 2
						GROUP BY r.street_id
					) combined
					GROUP BY div_id
				) sub ON sub.div_id = d.id
				WHERE d.parent_id = ? AND d.submission_status = 2 AND d.delete_time IS NULL
				ORDER BY COALESCE(sub.total_count, 0) DESC`
			args = []interface{}{*parentId}

		default:
			return []DivisionCountRow{}, nil
		}
	}

	var rows []DivisionCountRow
	err := m.QueryRowsNoCacheCtx(ctx, &rows, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (m *customMdDivisionStatisticsModel) RefreshStatistics(ctx context.Context) error {
	// Step 1: 统计 level 3 区县（按 county_id）
	countySQL := fmt.Sprintf(`INSERT INTO %s (division_id, level, community_count, village_count, total_count, stat_date, created_at)
		SELECT d.id, 3,
			COALESCE(SUM(CASE WHEN r.community_type = 1 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN r.community_type = 2 THEN 1 ELSE 0 END), 0),
			COUNT(r.id),
			CURDATE(), NOW()
		FROM md_administrative_division d
		INNER JOIN md_residential_area r ON r.county_id = d.id
			AND r.submission_status = 2 AND r.delete_time IS NULL
		WHERE d.level = 3 AND d.submission_status = 2 AND d.delete_time IS NULL
		GROUP BY d.id`, m.table)
	if _, err := m.ExecNoCacheCtx(ctx, countySQL); err != nil {
		return err
	}

	// Step 2: 统计 level 4 街道（按 street_id）
	streetSQL := fmt.Sprintf(`INSERT INTO %s (division_id, level, community_count, village_count, total_count, stat_date, created_at)
		SELECT d.id, 4,
			COALESCE(SUM(CASE WHEN r.community_type = 1 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN r.community_type = 2 THEN 1 ELSE 0 END), 0),
			COUNT(r.id),
			CURDATE(), NOW()
		FROM md_administrative_division d
		INNER JOIN md_residential_area r ON r.street_id = d.id
			AND r.submission_status = 2 AND r.delete_time IS NULL
		WHERE d.level = 4 AND d.submission_status = 2 AND d.delete_time IS NULL
		GROUP BY d.id`, m.table)
	if _, err := m.ExecNoCacheCtx(ctx, streetSQL); err != nil {
		return err
	}

	// Step 3: 聚合 level 2 市（通过 parent_id 关联区县统计）
	citySQL := fmt.Sprintf(`INSERT INTO %s (division_id, level, community_count, village_count, total_count, stat_date, created_at)
		SELECT d.id, 2,
			SUM(s.community_count), SUM(s.village_count), SUM(s.total_count),
			CURDATE(), NOW()
		FROM md_division_statistics s
		INNER JOIN md_administrative_division ct ON ct.id = s.division_id
			AND ct.level = 3 AND ct.submission_status = 2 AND ct.delete_time IS NULL
		INNER JOIN md_administrative_division d ON d.id = ct.parent_id
			AND d.level = 2 AND d.submission_status = 2 AND d.delete_time IS NULL
		WHERE s.level = 3 AND s.stat_date = CURDATE()
		GROUP BY d.id`, m.table)
	if _, err := m.ExecNoCacheCtx(ctx, citySQL); err != nil {
		return err
	}

	// Step 4: 聚合 level 1 省（通过 parent_id 关联市统计）
	provinceSQL := fmt.Sprintf(`INSERT INTO %s (division_id, level, community_count, village_count, total_count, stat_date, created_at)
		SELECT d.id, 1,
			SUM(s.community_count), SUM(s.village_count), SUM(s.total_count),
			CURDATE(), NOW()
		FROM md_division_statistics s
		INNER JOIN md_administrative_division ct ON ct.id = s.division_id
			AND ct.level = 2 AND ct.submission_status = 2 AND ct.delete_time IS NULL
		INNER JOIN md_administrative_division d ON d.id = ct.parent_id
			AND d.level = 1 AND d.submission_status = 2 AND d.delete_time IS NULL
		WHERE s.level = 2 AND s.stat_date = CURDATE()
		GROUP BY d.id`, m.table)
	if _, err := m.ExecNoCacheCtx(ctx, provinceSQL); err != nil {
		return err
	}

	return nil
}
