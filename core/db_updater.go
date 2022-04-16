package core

type DatabaseUpdater interface {
	CronWorker
	UpdateIntoDB()  // 更新数据到数据库
	MustTableExit() // 表必须存在，否则panic
}
