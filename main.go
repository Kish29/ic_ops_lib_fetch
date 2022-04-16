package main

import (
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/cron"
	"github.com/Kish29/ic_ops_lib_fetch/db"
	"github.com/Kish29/ic_ops_lib_fetch/integrate"
	"log"
	"time"
)

const (
	mysqlUsername = `root`
	mysqlPassword = `jiangaoran`
	mysqlDatabase = `bs`
)

func main() {
	// 注册所有的holder-fetcher
	integrator := integrate.NewLibIntegrator()
	integrator.AddHolder(core.NewBaseItemHolder(cron.NewConanFetcher()))
	integrator.AddHolder(core.NewBaseItemHolder(cron.NewVcpkgFetcher()))
	// 聚合器启动，这里需要一直等待所有的聚合器抓取完
	integrate.Startup(integrator)
	log.Println("Integrator startup success")

	// 初始化数据库连接
	dbConn := db.InitConn(mysqlUsername, mysqlPassword, mysqlDatabase, nil, nil)
	log.Println("Database init connection success")
	// 创建updater
	updater := db.NewBaseDatabaseUpdater(integrator, dbConn)
	// 启动updater的更新routine
	db.Startup(updater)
	log.Println("Updater startup success")

	// 启动接口的服务
	time.Sleep(time.Hour)
}
