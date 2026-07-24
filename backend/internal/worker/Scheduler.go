package worker

import (
	"context"
	"log"
	"time"
)

type Scheduler struct {
	runner   Runner
	interval time.Duration
}

func NewScheduler(runner Runner, interval time.Duration) *Scheduler {
	return &Scheduler{
		runner:   runner,
		interval: interval,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	log.Printf("Starting scheduler for runner %s (interval %s)" , s.runner.Name(), ticker )
	for {
		select {
		case <-ticker.C:
			err := s.runner.RunOnce(ctx)
			if err != nil {
				log.Printf("run failed with error: %v", err)
			}
		case <-ctx.Done():
			log.Println("Deletion scheduler stopped.")
			return
		}
	}
}
