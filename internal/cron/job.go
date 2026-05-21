package cron

type Job interface {
	Start()
	Stop()
}
