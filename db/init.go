package db

import (
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/robfig/cron"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

const (
	// MysqlCnnPattern 启用对time.Time类型的解析
	MysqlCnnPattern = `%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True`
	//MysqlCnnPattern         = `%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local`
	defaultMysqlAddr        = `127.0.0.1`
	defaultMysqlPort uint16 = 3306
)

// InitConn 初始化数据库的链接
func InitConn(username, password, dbname string, addr *string, port *uint16) (godb *gorm.DB) {
	mysqlAddr := defaultMysqlAddr
	mysqlPort := defaultMysqlPort
	if addr != nil {
		mysqlAddr = *addr
	}
	if port != nil {
		mysqlPort = *port
	}
	dsn := fmt.Sprintf(MysqlCnnPattern, username, password, mysqlAddr, mysqlPort, dbname)
	var err error
	godb, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		CreateBatchSize: 1000, // 分批插入时，1k条 column * row 必须< 65536
		NamingStrategy:  schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		panic(err)
	}
	d, err := godb.DB()
	if err != nil {
		panic(err)
	}
	// 设置链接参数
	d.SetMaxOpenConns(200)
	d.SetMaxIdleConns(200)
	d.SetConnMaxIdleTime(2 * time.Second)
	return
}

func Startup(updater core.DatabaseUpdater) {
	if updater == nil {
		panic("updater is nil")
	}
	// 表必须存在
	updater.MustTableExit()
	// 第一次启动时，更新到数据库
	updater.UpdateIntoDB()
	// 开启定时，更新数据到数据库
	c := cron.New()
	schedule := updater.CrontabSchedule()
	err := c.AddFunc(schedule, updater.UpdateIntoDB)
	if err != nil {
		panic(err)
	}
	c.Start()
}
