package main

import (
	"flag"
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/cron"
	"github.com/Kish29/ic_ops_lib_fetch/db"
	"github.com/Kish29/ic_ops_lib_fetch/download"
	"github.com/Kish29/ic_ops_lib_fetch/integrate"
	"github.com/gookit/goutil/fsutil"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	mysqlUsername = `root`
	mysqlPassword = `jiangaoran`
	mysqlDatabase = `bs`
)

var (
	flagVersion             bool
	flagDownloadSourceCode  bool
	flagOnlyDownloadSoource bool
)

func init() {
	flag.BoolVar(&flagVersion, "v", false, "show scrap version")
	flag.BoolVar(&flagDownloadSourceCode, "d", false, "whether download source code when fetch done")
	flag.BoolVar(&flagOnlyDownloadSoource, "od", false, "only download source codes from database")
	flag.Parse()
}

func main() {
	// 初始化数据库连接
	dbConn := db.InitConn(mysqlUsername, mysqlPassword, mysqlDatabase, nil, nil)
	log.Println("Database init connection success")

	if !flagOnlyDownloadSoource {
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

		// 创建updater
		updater := db.NewBaseDatabaseUpdater(integrator, dbConn)
		//启动updater的更新routine
		db.Startup(updater)
		log.Println("Updater startup success")
	}
	if flagDownloadSourceCode || flagOnlyDownloadSoource {
		if !fsutil.DirExist(download.SourceCodeDir) {
			err := os.Mkdir(download.SourceCodeDir, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}
		err := os.Chdir(download.SourceCodeDir)
		if err != nil {
			panic(err)
		}
		downloader := download.NewDBDownloader(dbConn)
		downloader.AddWget(download.NewGithubWget())
		downloader.AddWget(download.NewTarGZWget())
		download.StartupCronDownload(downloader)
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
