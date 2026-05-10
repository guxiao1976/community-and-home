package task

import (
	"log"

	"github.com/guxiao/community-and-home/services/masterdata/model"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
}

func NewScheduler(divModel model.MdDivisionStatisticsModel) *Scheduler {
	c := cron.New()
	s := &Scheduler{cron: c}

	statTask := NewStatisticsTask(divModel)

	c.AddFunc("0 2 * * *", func() {
		statTask.Run()
	})

	return s
}

func (s *Scheduler) Start() {
	s.cron.Start()
	log.Println("[Scheduler] 定时任务调度器已启动")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("[Scheduler] 定时任务调度器已停止")
}
