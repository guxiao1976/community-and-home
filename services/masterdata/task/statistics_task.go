package task

import (
	"context"
	"log"
	"time"

	"github.com/guxiao/community-and-home/services/masterdata/model"
)

type StatisticsTask struct {
	divModel model.MdDivisionStatisticsModel
}

func NewStatisticsTask(divModel model.MdDivisionStatisticsModel) *StatisticsTask {
	return &StatisticsTask{divModel: divModel}
}

func (t *StatisticsTask) Run() {
	ctx := context.Background()
	log.Println("[StatisticsTask] 开始执行统计任务...")
	start := time.Now()

	if err := t.divModel.DeleteByDate(ctx, time.Now()); err != nil {
		log.Printf("[StatisticsTask] 清除当天数据失败: %v\n", err)
		return
	}

	if err := t.divModel.RefreshStatistics(ctx); err != nil {
		log.Printf("[StatisticsTask] 统计写入失败: %v\n", err)
		return
	}

	log.Printf("[StatisticsTask] 统计完成（耗时 %v）\n", time.Since(start))
}
