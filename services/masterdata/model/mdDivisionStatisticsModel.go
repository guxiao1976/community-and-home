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
