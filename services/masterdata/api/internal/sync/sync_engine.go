package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/guxiao/community-and-home/services/masterdata/model"
	"github.com/zeromicro/go-zero/core/logx"
)

type SyncStatus string

const (
	SyncStatusRunning   SyncStatus = "running"
	SyncStatusCompleted SyncStatus = "completed"
	SyncStatusFailed    SyncStatus = "failed"
)

type SyncProgress struct {
	mu                 sync.Mutex
	TaskId             string     `json:"task_id"`
	Status             SyncStatus `json:"status"`
	TotalCounties      int32      `json:"total_counties"`
	CurrentCountyIdx   int32      `json:"current_county"`
	CurrentCountyName  string     `json:"current_county_name"`
	TotalKeywords      int32      `json:"total_keywords"`
	CurrentKeywordIdx  int32      `json:"current_keyword"`
	CurrentKeyword     string     `json:"current_keyword"`
	TotalPages         int32      `json:"total_pages"`
	CurrentPage        int32      `json:"current_page"`
	TotalFound         int32      `json:"total_found"`
	TotalSynced        int32      `json:"total_synced"`
	TotalSkipped       int32      `json:"total_skipped"`
	TotalFailed        int32      `json:"total_failed"`
	ErrorMessage       string     `json:"error_message,omitempty"`
}

type SyncEngine struct {
	mu        sync.Mutex
	tasks     map[string]*SyncProgress
	amapKey   string
	divModel  model.MdAdministrativeDivisionModel
	areaModel model.MdResidentialAreaModel
	client    *http.Client
}

func NewSyncEngine(amapKey string, divModel model.MdAdministrativeDivisionModel, areaModel model.MdResidentialAreaModel) *SyncEngine {
	return &SyncEngine{
		tasks:     make(map[string]*SyncProgress),
		amapKey:   amapKey,
		divModel:  divModel,
		areaModel: areaModel,
		client:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (e *SyncEngine) StartSync(ctx context.Context, divisionId int64) string {
	taskId := strconv.FormatInt(time.Now().UnixNano(), 10)
	p := &SyncProgress{
		TaskId: taskId,
		Status: SyncStatusRunning,
	}
	e.mu.Lock()
	e.tasks[taskId] = p
	e.mu.Unlock()
	go e.runMultiCountySync(context.Background(), p, divisionId)
	return taskId
}

func (e *SyncEngine) GetProgress(taskId string) *SyncProgress {
	e.mu.Lock()
	defer e.mu.Unlock()
	p, ok := e.tasks[taskId]
	if !ok {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	cp := *p
	return &cp
}

// AMap v3 text search response types

type amapTextSearchResp struct {
	Status string    `json:"status"`
	Count  string    `json:"count"`
	POIs   []amapPOI `json:"pois"`
}

type amapPOI struct {
	Name     string          `json:"name"`
	Address  json.RawMessage `json:"address"`
	Location string          `json:"location"`
	Adcode   string          `json:"adcode"`
}

func (p *amapPOI) GetAddress() string {
	if p.Address == nil {
		return ""
	}
	var s string
	if err := json.Unmarshal(p.Address, &s); err == nil {
		return s
	}
	var arr []string
	if err := json.Unmarshal(p.Address, &arr); err == nil && len(arr) > 0 {
		return arr[0]
	}
	return string(p.Address)
}

func ptrInt64(v int64) *int64 { return &v }

// resolveCounties returns all county-level (level=3) divisions under the given division.
func (e *SyncEngine) resolveCounties(ctx context.Context, divisionId int64) ([]*model.MdAdministrativeDivision, error) {
	div, err := e.divModel.FindOne(ctx, divisionId)
	if err != nil {
		return nil, err
	}

	switch div.Level {
	case 4:
		// Street level - return the street itself for direct sync
		return []*model.MdAdministrativeDivision{div}, nil
	case 3:
		return []*model.MdAdministrativeDivision{div}, nil
	case 2:
		return e.divModel.FindChildrenWithFilter(ctx, divisionId, ptrInt64(3), nil)
	case 1:
		cities, err := e.divModel.FindChildrenWithFilter(ctx, divisionId, ptrInt64(2), nil)
		if err != nil {
			return nil, err
		}
		var counties []*model.MdAdministrativeDivision
		for _, city := range cities {
			cityCounties, err := e.divModel.FindChildrenWithFilter(ctx, city.Id, ptrInt64(3), nil)
			if err != nil {
				return nil, fmt.Errorf("failed to load counties for city %s (id=%d): %w", city.Name, city.Id, err)
			}
			counties = append(counties, cityCounties...)
		}
		return counties, nil
	default:
		return nil, fmt.Errorf("unsupported division level %d (must be 1/2/3/4)", div.Level)
	}
}

func (e *SyncEngine) runMultiCountySync(ctx context.Context, p *SyncProgress, divisionId int64) {
	defer func() {
		if r := recover(); r != nil {
			p.mu.Lock()
			p.Status = SyncStatusFailed
			p.ErrorMessage = fmt.Sprintf("panic: %v", r)
			p.mu.Unlock()
		}
	}()

	counties, err := e.resolveCounties(ctx, divisionId)
	if err != nil {
		p.mu.Lock()
		p.Status = SyncStatusFailed
		p.ErrorMessage = "获取区县列表失败: " + err.Error()
		p.mu.Unlock()
		return
	}

	if len(counties) == 0 {
		p.mu.Lock()
		p.TotalCounties = 0
		p.Status = SyncStatusCompleted
		p.mu.Unlock()
		return
	}

	p.mu.Lock()
	p.TotalCounties = int32(len(counties))
	p.mu.Unlock()

	for i, county := range counties {
		p.mu.Lock()
		p.CurrentCountyIdx = int32(i + 1)
		p.CurrentCountyName = county.Name
		p.TotalPages = 0
		p.CurrentPage = 0
		p.mu.Unlock()

		logx.Infof("[AMap Sync] %d/%d 开始同步: %s (code=%s)", i+1, len(counties), county.Name, county.Code)

		e.runSyncSingleCounty(ctx, p, county.Id)

		if i < len(counties)-1 {
			delay := time.Duration(5+rand.Intn(56)) * time.Second
			logx.Infof("[AMap Sync] 等待 %v 后处理下一个区县...", delay)
			time.Sleep(delay)
		}
	}

	p.mu.Lock()
	p.CurrentCountyIdx = p.TotalCounties
	p.CurrentCountyName = ""
	p.Status = SyncStatusCompleted
	p.mu.Unlock()
}

func (e *SyncEngine) runSyncSingleCounty(ctx context.Context, p *SyncProgress, countyId int64) {
	county, err := e.divModel.FindOne(ctx, countyId)
	if err != nil {
		logx.Errorf("[AMap Sync] find county %d failed: %v", countyId, err)
		return
	}

	var countyCode string
	var cityId int64

	// Handle both county (level 3) and street (level 4)
	if county.Level == 4 {
		// Street level: use parent county's code
		if !county.ParentId.Valid {
			logx.Errorf("[AMap Sync] street %d has no parent county", countyId)
			return
		}
		parentCounty, err := e.divModel.FindOne(ctx, county.ParentId.Int64)
		if err != nil {
			logx.Errorf("[AMap Sync] find parent county %d failed: %v", county.ParentId.Int64, err)
			return
		}
		countyCode = parentCounty.Code
		if parentCounty.ParentId.Valid {
			cityId = parentCounty.ParentId.Int64
		}
		logx.Infof("[AMap Sync] Street-level sync: using county code %s", countyCode)
	} else {
		// County level: use county code directly
		countyCode = county.Code
		if county.ParentId.Valid {
			cityId = county.ParentId.Int64
		}
	}

	// 20 common keywords for residential area names
	keywords := []string{
		"小区", "花园", "苑", "公寓", "村", "家园", "居", "城", "府", "庭",
		"轩", "阁", "园", "坊", "里", "邸", "郡", "湾", "台", "座",
	}

	p.mu.Lock()
	p.TotalKeywords = int32(len(keywords))
	p.mu.Unlock()

	// Memory-based deduplication using POI ID
	seenPOIs := make(map[string]bool)

	maxCode, err := e.areaModel.GetMaxCodeByCountyId(ctx, countyId, countyCode)
	var nextSeq int
	if err != nil || maxCode == "" {
		nextSeq = 1
	} else {
		seqStr := maxCode[len(maxCode)-4:]
		nextSeq, _ = strconv.Atoi(seqStr)
		nextSeq++
	}

	for kwIdx, keyword := range keywords {
		p.mu.Lock()
		p.CurrentKeywordIdx = int32(kwIdx + 1)
		p.CurrentKeyword = keyword
		p.TotalPages = 0
		p.CurrentPage = 0
		p.mu.Unlock()

		logx.Infof("[AMap Sync] %d/%d 关键词: '%s'", kwIdx+1, len(keywords), keyword)

		firstResp, err := e.searchResidentialAreas(countyCode, 1, keyword)
		if err != nil {
			logx.Errorf("[AMap Sync] search with keyword '%s' failed: %v", keyword, err)
			continue
		}

		totalCount, _ := strconv.Atoi(firstResp.Count)
		if totalCount == 0 {
			continue
		}

		totalPages := totalCount / 25
		if totalCount%25 > 0 {
			totalPages++
		}
		// AMap API limit: max 18 pages (450 results)
		if totalPages > 18 {
			totalPages = 18
		}

		p.mu.Lock()
		p.TotalPages = int32(totalPages)
		p.mu.Unlock()

		for page := 1; page <= totalPages; page++ {
			p.mu.Lock()
			p.CurrentPage = int32(page)
			p.mu.Unlock()

			var resp *amapTextSearchResp
			if page == 1 {
				resp = firstResp
			} else {
				resp, err = e.searchResidentialAreas(countyCode, page, keyword)
				if err != nil {
					logx.Errorf("[AMap Sync] search page %d with keyword '%s' failed: %v", page, keyword, err)
					continue
				}
			}

			if len(resp.POIs) == 0 {
				break
			}

			for _, poi := range resp.POIs {
				// Deduplicate by POI ID (from AMap)
				poiId := poi.Name + "|" + poi.Location
				if seenPOIs[poiId] {
					p.mu.Lock()
					p.TotalSkipped++
					p.mu.Unlock()
					continue
				}
				seenPOIs[poiId] = true

				// Check if already exists in database
				existing, err := e.areaModel.FindByNameAndCountyId(ctx, poi.Name, countyId)
				if err == nil && existing != nil {
					p.mu.Lock()
					p.TotalSkipped++
					p.mu.Unlock()
					continue
				}

				code := fmt.Sprintf("%s%04d", countyCode, nextSeq)
				for {
					codeExists, _ := e.areaModel.FindByCode(ctx, code)
					if codeExists == nil {
						break
					}
					nextSeq++
					code = fmt.Sprintf("%s%04d", countyCode, nextSeq)
					if nextSeq > 9999 {
						logx.Errorf("[AMap Sync] cannot generate unique code for county %s", countyCode)
						p.mu.Lock()
						p.TotalFailed++
						p.mu.Unlock()
						break
					}
				}

				var longitude, latitude sql.NullFloat64
				if poi.Location != "" {
					parts := strings.Split(poi.Location, ",")
					if len(parts) == 2 {
						if lng, err := strconv.ParseFloat(parts[0], 64); err == nil {
							longitude = sql.NullFloat64{Float64: lng, Valid: true}
						}
						if lat, err := strconv.ParseFloat(parts[1], 64); err == nil {
							latitude = sql.NullFloat64{Float64: lat, Valid: true}
						}
					}
				}

				now := time.Now()
				area := &model.MdResidentialArea{
					CountyId:         sql.NullInt64{Int64: countyId, Valid: true},
					CityId:           sql.NullInt64{Int64: cityId, Valid: true},
					Code:             sql.NullString{String: code, Valid: true},
					Name:             poi.Name,
					Address:          poi.GetAddress(),
					Longitude:        longitude,
					Latitude:         latitude,
					DataSource:       1,
					CommunityType:    1,
					SubmissionStatus: 2,
					SubmitterId:      0,
					CreatedTime:      now,
					UpdatedTime:      now,
				}

				_, err = e.areaModel.Insert(ctx, area)
				if err != nil {
					logx.Errorf("[AMap Sync] insert residential area failed: %v", err)
					p.mu.Lock()
					p.TotalFailed++
					p.mu.Unlock()
					continue
				}

				nextSeq++
				p.mu.Lock()
				p.TotalSynced++
				p.TotalFound++
				p.mu.Unlock()
			}

			if page < totalPages {
				delay := time.Duration(5+rand.Intn(6)) * time.Second
				time.Sleep(delay)
			}
		}

		// Delay between keywords
		if kwIdx < len(keywords)-1 {
			delay := time.Duration(3+rand.Intn(5)) * time.Second
			logx.Infof("[AMap Sync] 等待 %v 后处理下一个关键词...", delay)
			time.Sleep(delay)
		}
	}

	logx.Infof("[AMap Sync] 区县 %s 完成，共发现 %d 个小区（去重后），同步 %d 个，跳过 %d 个",
		county.Name, len(seenPOIs), p.TotalSynced, p.TotalSkipped)
}

func (e *SyncEngine) searchResidentialAreas(countyCode string, page int, keywords string) (*amapTextSearchResp, error) {
	reqUrl := fmt.Sprintf(
		"https://restapi.amap.com/v3/place/text?keywords=%s&types=120300&city=%s&citylimit=true&offset=25&page=%d&key=%s",
		keywords, countyCode, page, e.amapKey,
	)

	resp, err := e.client.Get(reqUrl)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body failed: %w", err)
	}

	var result amapTextSearchResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	if result.Status != "1" {
		return nil, fmt.Errorf("AMap API returned status: %s", result.Status)
	}

	return &result, nil
}
