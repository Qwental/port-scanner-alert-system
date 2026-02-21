package scheduler

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type TaskFunc func(ctx context.Context) error

type Scheduler struct {
	interval time.Duration
	task     TaskFunc
	log      *zap.SugaredLogger
}

func New(interval time.Duration, task TaskFunc, log *zap.SugaredLogger) *Scheduler {
	return &Scheduler{
		interval: interval,
		task:     task,
		log:      log,
	}
}

// type SchedulerConfig struct {
// 	Enabled  bool   `yaml:"enabled"`
// 	Interval string `yaml:"interval"`
// }

func (s *Scheduler) Run(ctx context.Context) {
	s.log.Infof("Running first scan...")
	if err := s.task(ctx); err != nil {
		s.log.Errorf("Task failed: %v", err)
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.log.Infof("Next scan in %s", s.interval)

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Scheduler stopped")
			return
		case <-ticker.C:
			s.log.Info("Running scheduled scan...")
			if err := s.task(ctx); err != nil {
				s.log.Errorf("Task failed: %v", err)
			}
			s.log.Infof("Next scan in %s", s.interval)
		}
	}
}