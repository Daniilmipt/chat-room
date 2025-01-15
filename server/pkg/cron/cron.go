package cron

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type cronSchedule struct {
	jobs []CronJob
}

func NewSchedule(jobs []CronJob) *cronSchedule {
	return &cronSchedule{
		jobs: jobs,
	}
}

func (sch *cronSchedule) Start() {
	for _, job := range sch.jobs {
		job.Start()
	}
}

type CronJob struct {
	logger *zap.Logger
	period time.Duration
	task   func()
	mu     *sync.Mutex
}

func NewJob(logger *zap.Logger, period time.Duration, name string, task func()) CronJob {
	return CronJob{
		logger: logger.With(zap.String("name", name)),
		period: period,
		task:   task,
		mu:     &sync.Mutex{},
	}
}

func (cj CronJob) Start() {
	go func() {
		ticker := time.NewTicker(cj.period)
		defer ticker.Stop()

		for range ticker.C {
			cj.runTask()
		}
	}()
}

func (cj CronJob) runTask() {
	cj.mu.Lock()
	defer cj.mu.Unlock()

	if cj.logger != nil {
		cj.logger.Info("start job")
		defer cj.logger.Info("end job")
	}
	cj.task()
}
