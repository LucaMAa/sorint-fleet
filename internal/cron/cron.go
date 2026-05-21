package cron

import (
	"log"
)

type Cron struct {
	jobs []Job
}

func New() *Cron {
	return &Cron{
		jobs: []Job{
			NewJollyNotifier(),
		},
	}
}

func (c *Cron) Start() {
	log.Println("[Cron] starting all jobs...")
	for _, j := range c.jobs {
		j.Start()
	}
}

func (c *Cron) Stop() {
	log.Println("[Cron] stopping all jobs...")
	for _, j := range c.jobs {
		j.Stop()
	}
}
