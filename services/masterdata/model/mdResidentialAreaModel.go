package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ MdResidentialAreaModel = (*customMdResidentialAreaModel)(nil)

type (
	MdResidentialAreaModel interface {
		mdResidentialAreaModel
		FindByCountyId(ctx context.Context, countyId int64, submissionStatus *int32, communityType *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error)
		FindByCountyIds(ctx context.Context, countyIds []int64, submissionStatus *int32, communityType *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error)
		FindByStreetId(ctx context.Context, streetId int64, submissionStatus *int32, communityType *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error)
		FindByCommunityDivId(ctx context.Context, communityDivId int64, submissionStatus *int32, communityType *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error)
		FindBySubmissionStatus(ctx context.Context, status int64, page, pageSize int64) ([]*MdResidentialArea, error)
		FindAll(ctx context.Context, submissionStatus *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error)
		FindByCode(ctx context.Context, code string) (*MdResidentialArea, error)
		SearchByName(ctx context.Context, keyword string, countyId *int64, streetId *int64, communityDivId *int64, countyIds []int64, submissionStatus *int32, communityType *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error)
		Count(ctx context.Context, countyId *int64, streetId *int64, communityDivId *int64, submissionStatus *int32, countyIds []int64, keyword *string, communityType *int32, excludeStatuses ...int32) (int64, error)
		GetMaxCodeByCountyId(ctx context.Context, countyId int64, countyCode string) (string, error)
		FindByNameAndCountyId(ctx context.Context, name string, countyId int64) (*MdResidentialArea, error)
		CountBySubmissionStatus(ctx context.Context, status int64) (int64, error)
		FindPendingBySubmissionStatus(ctx context.Context, status int64, submissionType *int64, page, pageSize int64) ([]*MdResidentialArea, error)
		CountDeleted(ctx context.Context) (int64, error)
		FindDeleted(ctx context.Context, page, pageSize int64) ([]*MdResidentialArea, int64, error)
		Restore(ctx context.Context, id int64) error
		CountByCommunityType(ctx context.Context, communityType int32, excludeStatuses ...int32) (int64, error)
	}

	customMdResidentialAreaModel struct {
		*defaultMdResidentialAreaModel
	}
)

func NewMdResidentialAreaModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) MdResidentialAreaModel {
	return &customMdResidentialAreaModel{
		defaultMdResidentialAreaModel: newMdResidentialAreaModel(conn, c, opts...),
	}
}

func appendSubmissionFilter(conditions *[]string, args *[]interface{}, submissionStatus *int32, excludeStatuses ...int32) {
	if submissionStatus != nil {
		*conditions = append(*conditions, "submission_status = ?")
		*args = append(*args, *submissionStatus)
	} else if len(excludeStatuses) > 0 {
		placeholders := make([]string, len(excludeStatuses))
		for i, s := range excludeStatuses {
			placeholders[i] = "?"
			*args = append(*args, s)
		}
		*conditions = append(*conditions, "submission_status not in ("+strings.Join(placeholders, ",")+")")
	}
}

func (m *customMdResidentialAreaModel) FindByCountyId(ctx context.Context, countyId int64, submissionStatus *int32, communityType *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error) {
	var areas []*MdResidentialArea
	offset := (page - 1) * pageSize
	var conditions []string
	var args []interface{}
	conditions = append(conditions, "county_id = ?")
	args = append(args, countyId)
	appendSubmissionFilter(&conditions, &args, submissionStatus, excludeStatuses...)
	if communityType != nil {
		conditions = append(conditions, "community_type = ?")
		args = append(args, *communityType)
	}
	where := strings.Join(conditions, " and ")
	query := fmt.Sprintf("select %s from %s where %s and delete_time is null order by id desc limit ? offset ?", mdResidentialAreaRows, m.table, where)
	args = append(args, pageSize, offset)
	err := m.QueryRowsNoCacheCtx(ctx, &areas, query, args...)
	if err != nil {
		return nil, err
	}
	return areas, nil
}

func (m *customMdResidentialAreaModel) FindByCountyIds(ctx context.Context, countyIds []int64, submissionStatus *int32, communityType *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error) {
	if len(countyIds) == 0 {
		return nil, nil
	}
	var areas []*MdResidentialArea
	offset := (page - 1) * pageSize
	placeholders := make([]string, len(countyIds))
	args := make([]interface{}, len(countyIds))
	for i, id := range countyIds {
		placeholders[i] = "?"
		args[i] = id
	}
	var conditions []string
	conditions = append(conditions, "county_id in ("+strings.Join(placeholders, ",")+")")
	appendSubmissionFilter(&conditions, &args, submissionStatus, excludeStatuses...)
	if communityType != nil {
		conditions = append(conditions, "community_type = ?")
		args = append(args, *communityType)
	}
	where := strings.Join(conditions, " and ")
	query := fmt.Sprintf("select %s from %s where %s and delete_time is null order by id desc limit ? offset ?", mdResidentialAreaRows, m.table, where)
	args = append(args, pageSize, offset)
	err := m.QueryRowsNoCacheCtx(ctx, &areas, query, args...)
	if err != nil {
		return nil, err
	}
	return areas, nil
}

func (m *customMdResidentialAreaModel) FindByStreetId(ctx context.Context, streetId int64, submissionStatus *int32, communityType *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error) {
	var areas []*MdResidentialArea
	offset := (page - 1) * pageSize
	var conditions []string
	var args []interface{}
	conditions = append(conditions, "street_id = ?")
	args = append(args, streetId)
	appendSubmissionFilter(&conditions, &args, submissionStatus, excludeStatuses...)
	if communityType != nil {
		conditions = append(conditions, "community_type = ?")
		args = append(args, *communityType)
	}
	where := strings.Join(conditions, " and ")
	query := fmt.Sprintf("select %s from %s where %s and delete_time is null order by id desc limit ? offset ?", mdResidentialAreaRows, m.table, where)
	args = append(args, pageSize, offset)
	err := m.QueryRowsNoCacheCtx(ctx, &areas, query, args...)
	if err != nil {
		return nil, err
	}
	return areas, nil
}

func (m *customMdResidentialAreaModel) FindByCommunityDivId(ctx context.Context, communityDivId int64, submissionStatus *int32, communityType *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error) {
	var areas []*MdResidentialArea
	offset := (page - 1) * pageSize
	var conditions []string
	var args []interface{}
	conditions = append(conditions, "community_div_id = ?")
	args = append(args, communityDivId)
	appendSubmissionFilter(&conditions, &args, submissionStatus, excludeStatuses...)
	if communityType != nil {
		conditions = append(conditions, "community_type = ?")
		args = append(args, *communityType)
	}
	where := strings.Join(conditions, " and ")
	query := fmt.Sprintf("select %s from %s where %s and delete_time is null order by id desc limit ? offset ?", mdResidentialAreaRows, m.table, where)
	args = append(args, pageSize, offset)
	err := m.QueryRowsNoCacheCtx(ctx, &areas, query, args...)
	if err != nil {
		return nil, err
	}
	return areas, nil
}

func (m *customMdResidentialAreaModel) FindBySubmissionStatus(ctx context.Context, status int64, page, pageSize int64) ([]*MdResidentialArea, error) {
	var areas []*MdResidentialArea
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where submission_status = ? and delete_time is null order by id desc limit ? offset ?", mdResidentialAreaRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &areas, query, status, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return areas, nil
}

func (m *customMdResidentialAreaModel) FindAll(ctx context.Context, submissionStatus *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error) {
	var areas []*MdResidentialArea
	offset := (page - 1) * pageSize
	var conditions []string
	var args []interface{}
	appendSubmissionFilter(&conditions, &args, submissionStatus, excludeStatuses...)
	query := fmt.Sprintf("select %s from %s where delete_time is null", mdResidentialAreaRows, m.table)
	if len(conditions) > 0 {
		query += " and " + strings.Join(conditions, " and ")
	}
	query += " order by id desc limit ? offset ?"
	args = append(args, pageSize, offset)
	err := m.QueryRowsNoCacheCtx(ctx, &areas, query, args...)
	if err != nil {
		return nil, err
	}
	return areas, nil
}

func (m *customMdResidentialAreaModel) FindByCode(ctx context.Context, code string) (*MdResidentialArea, error) {
	var area MdResidentialArea
	query := fmt.Sprintf("select %s from %s where code = ? and delete_time is null limit 1", mdResidentialAreaRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &area, query, code)
	if err != nil {
		return nil, err
	}
	return &area, nil
}

func (m *customMdResidentialAreaModel) SearchByName(ctx context.Context, keyword string, countyId *int64, streetId *int64, communityDivId *int64, countyIds []int64, submissionStatus *int32, communityType *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error) {
	var areas []*MdResidentialArea
	offset := (page - 1) * pageSize
	like := "%" + keyword + "%"

	var conditions []string
	var args []interface{}
	conditions = append(conditions, "name like ?")
	args = append(args, like)

	if communityDivId != nil {
		conditions = append(conditions, "community_div_id = ?")
		args = append(args, *communityDivId)
	} else if streetId != nil {
		conditions = append(conditions, "street_id = ?")
		args = append(args, *streetId)
	} else if countyId != nil {
		conditions = append(conditions, "county_id = ?")
		args = append(args, *countyId)
	} else if len(countyIds) > 0 {
		placeholders := make([]string, len(countyIds))
		for i, id := range countyIds {
			placeholders[i] = "?"
			args = append(args, id)
		}
		conditions = append(conditions, "county_id in ("+strings.Join(placeholders, ",")+")")
	}

	appendSubmissionFilter(&conditions, &args, submissionStatus, excludeStatuses...)

	if communityType != nil {
		conditions = append(conditions, "community_type = ?")
		args = append(args, *communityType)
	}

	where := strings.Join(conditions, " and ")
	query := fmt.Sprintf("select %s from %s where %s and delete_time is null order by id desc limit ? offset ?", mdResidentialAreaRows, m.table, where)
	args = append(args, pageSize, offset)
	err := m.QueryRowsNoCacheCtx(ctx, &areas, query, args...)
	if err != nil {
		return nil, err
	}
	return areas, nil
}

func (m *customMdResidentialAreaModel) Count(ctx context.Context, countyId *int64, streetId *int64, communityDivId *int64, submissionStatus *int32, countyIds []int64, keyword *string, communityType *int32, excludeStatuses ...int32) (int64, error) {
	var count int64
	var conditions []string
	var args []interface{}

	if communityDivId != nil {
		conditions = append(conditions, "community_div_id = ?")
		args = append(args, *communityDivId)
	} else if streetId != nil {
		conditions = append(conditions, "street_id = ?")
		args = append(args, *streetId)
	} else if countyId != nil {
		conditions = append(conditions, "county_id = ?")
		args = append(args, *countyId)
	} else if len(countyIds) > 0 {
		placeholders := make([]string, len(countyIds))
		for i, id := range countyIds {
			placeholders[i] = "?"
			args = append(args, id)
		}
		conditions = append(conditions, "county_id in ("+strings.Join(placeholders, ",")+")")
	}

	appendSubmissionFilter(&conditions, &args, submissionStatus, excludeStatuses...)

	if keyword != nil && *keyword != "" {
		conditions = append(conditions, "name like ?")
		args = append(args, "%"+*keyword+"%")
	}

	if communityType != nil {
		conditions = append(conditions, "community_type = ?")
		args = append(args, *communityType)
	}

	query := fmt.Sprintf("select count(*) from %s where delete_time is null", m.table)
	if len(conditions) > 0 {
		query += " and " + strings.Join(conditions, " and ")
	}

	err := m.QueryRowNoCacheCtx(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customMdResidentialAreaModel) GetMaxCodeByCountyId(ctx context.Context, countyId int64, countyCode string) (string, error) {
	prefix := countyCode + "%"
	query := fmt.Sprintf("select code from %s where county_id = ? and code like ? order by code desc limit 1", m.table)
	var code string
	err := m.QueryRowNoCacheCtx(ctx, &code, query, countyId, prefix)
	if err != nil {
		return "", nil
	}
	return code, nil
}

func (m *customMdResidentialAreaModel) FindByNameAndCountyId(ctx context.Context, name string, countyId int64) (*MdResidentialArea, error) {
	var area MdResidentialArea
	query := fmt.Sprintf("select %s from %s where name = ? and county_id = ? and delete_time is null limit 1", mdResidentialAreaRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &area, query, name, countyId)
	if err != nil {
		return nil, err
	}
	return &area, nil
}

func (m *customMdResidentialAreaModel) CountBySubmissionStatus(ctx context.Context, status int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where submission_status = ? and delete_time is null", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, status)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customMdResidentialAreaModel) FindPendingBySubmissionStatus(ctx context.Context, status int64, submissionType *int64, page, pageSize int64) ([]*MdResidentialArea, error) {
	var areas []*MdResidentialArea
	offset := (page - 1) * pageSize
	var conditions []string
	var args []interface{}
	conditions = append(conditions, "submission_status = ?")
	args = append(args, status)
	if submissionType != nil {
		conditions = append(conditions, "submission_type = ?")
		args = append(args, *submissionType)
	}
	where := strings.Join(conditions, " and ")
	query := fmt.Sprintf("select %s from %s where %s and delete_time is null order by id desc limit ? offset ?", mdResidentialAreaRows, m.table, where)
	args = append(args, pageSize, offset)
	err := m.QueryRowsNoCacheCtx(ctx, &areas, query, args...)
	if err != nil {
		return nil, err
	}
	return areas, nil
}

func (m *customMdResidentialAreaModel) Restore(ctx context.Context, id int64) error {
	mdResidentialAreaIdKey := fmt.Sprintf("%s%v", cacheMdResidentialAreaIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (sql.Result, error) {
		query := fmt.Sprintf("update %s set delete_time = null, submission_status = 2 where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, mdResidentialAreaIdKey)
	return err
}

func (m *customMdResidentialAreaModel) CountByCommunityType(ctx context.Context, communityType int32, excludeStatuses ...int32) (int64, error) {
	var count int64
	var conditions []string
	var args []interface{}
	conditions = append(conditions, "community_type = ?")
	args = append(args, communityType)
	appendSubmissionFilter(&conditions, &args, nil, excludeStatuses...)
	where := strings.Join(conditions, " and ")
	query := fmt.Sprintf("select count(*) from %s where %s and delete_time is null", m.table, where)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (m *customMdResidentialAreaModel) CountDeleted(ctx context.Context) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where delete_time is not null", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query)
	return count, err
}

func (m *customMdResidentialAreaModel) FindDeleted(ctx context.Context, page, pageSize int64) ([]*MdResidentialArea, int64, error) {
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where delete_time is not null", m.table)
	if err := m.QueryRowNoCacheCtx(ctx, &total, countQuery); err != nil {
		return nil, 0, err
	}
	var areas []*MdResidentialArea
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where delete_time is not null order by delete_time desc limit ? offset ?", mdResidentialAreaRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &areas, query, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	return areas, total, nil
}
