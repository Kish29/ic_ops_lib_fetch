package core

type Fetcher interface {
	Fetch() ([]*LibInfo, error) // 异步获取开源库信息
	Name() string
}

type CronWorker interface {
	CrontabSchedule() string // 异步加载的频率，格式为crontab的格式
}
