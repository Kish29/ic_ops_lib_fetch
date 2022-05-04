package main

import (
	"flag"
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/cron"
	"github.com/Kish29/ic_ops_lib_fetch/db"
	"github.com/Kish29/ic_ops_lib_fetch/download"
	"github.com/Kish29/ic_ops_lib_fetch/integrate"
	"log"
	"os"
	"time"
)

const (
	mysqlUsername = `root`
	mysqlPassword = `jiangaoran`
	mysqlDatabase = `bs`
)

var (
	flagVersion            bool
	flagDownloadSourceCode bool
)

func init() {
	flag.BoolVar(&flagVersion, "v", false, "show scrap version")
	flag.BoolVar(&flagDownloadSourceCode, "d", false, "whether download source code when fetch done")
	flag.Parse()
}

func main() {
	// 注册所有的holder-fetcher
	integrator := integrate.NewLibIntegrator()
	// conan
	integrator.AddHolder(core.NewBaseItemHolder(cron.NewConanFetcher()))
	// vcpkg
	integrator.AddHolder(core.NewBaseItemHolder(cron.NewVcpkgFetcher()))
	// qpm
	integrator.AddHolder(core.NewBaseItemHolder(cron.NewQPMFetcher()))
	// hunter
	integrator.AddHolder(core.NewBaseItemHolder(cron.NewHunterFetcher()))
	// cppan
	integrator.AddHolder(core.NewBaseItemHolder(cron.NewCppanFetcher()))
	// spack
	integrator.AddHolder(core.NewBaseItemHolder(cron.NewSpackFetcher()))
	//聚合器启动，这里需要一直等待所有的聚合器抓取完
	integrate.Startup(integrator)
	log.Println("Integrator startup success")

	// 初始化数据库连接
	dbConn := db.InitConn(mysqlUsername, mysqlPassword, mysqlDatabase, nil, nil)
	log.Println("Database init connection success")
	// 创建updater
	updater := db.NewBaseDatabaseUpdater(integrator, dbConn)
	//启动updater的更新routine
	db.Startup(updater)
	log.Println("Updater startup success")

	if flagDownloadSourceCode {
		go func() {
			gitDownloader := download.InitGitDownloader(func(url string, succ bool) {
				log.Printf("Download for=>%s %v", url, succ)
			})
			_ = os.Mkdir(download.SourceCodeDir, os.ModeDir|os.ModePerm)
			err := os.Chdir(download.SourceCodeDir)
			if err != nil {
				panic(err)
			}
			gitDownloader.DownloadAllInDB(dbConn)
		}()
	}
	// TODO: 启动接口的服务
	time.Sleep(time.Hour)
}
