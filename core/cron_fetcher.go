package core

// AsyncCronFetcher 异步定时加载
type AsyncCronFetcher interface {
	Fetcher
	CronWorker
}

type BaseAsyncCronFetcher struct{}

func (b *BaseAsyncCronFetcher) CrontabSchedule() string {
	return "0 30 0 * * ?" // 每天0点30分更新
}
