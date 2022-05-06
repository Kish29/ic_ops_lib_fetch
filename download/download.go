package download

import (
	"errors"
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/db"
	"github.com/robfig/cron"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

const (
	SourceCodeDir = "source_code"
)

type Downloader interface {
	core.CronWorker
	Download() error
}

type Wget interface {
	Get() error
	CanGetAndPut(url string) bool // 是否可对该url进行下载
}

// 检查当前目录下是否有该文件
func FileExist(dir, filename string) bool {
	p := filepath.Join(dir, filename)
	var exist = true
	if _, err := os.Stat(p); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func Exists(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	} else {
		return true
	}
}

func IsDir(dir string) bool {
	s, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func IsFile(filename string) bool {
	return !IsDir(filename)
}

type DBDownloader struct {
	wgets []Wget
	conn  *gorm.DB
}

func NewDBDownloader(conn *gorm.DB) *DBDownloader {
	return &DBDownloader{conn: conn, wgets: make([]Wget, 0, 8)}
}

func (d *DBDownloader) AddWget(wget Wget) {
	d.wgets = append(d.wgets, wget)
}

func (d *DBDownloader) CrontabSchedule() string {
	return "0 0 */3 * * ?" // 每隔3小时
}

func (d *DBDownloader) Download() error {
	// 从DB加载所有的可下载的组件
	if d.conn == nil {
		return errors.New("database connect is nil")
	}
	var downloads []*db.TBsLibInfo
	res := d.conn.Or("homepage <> ?", "").Or("source_code <> ?", "").Find(&downloads)
	if res.Error != nil {
		panic(res.Error)
	}
	for _, download := range downloads {
		// 添加进wget组件
		for _, wget := range d.wgets {
			if wget.CanGetAndPut(download.Homepage) || wget.CanGetAndPut(download.SourceCode) {
				break
			}
		}
	}
	for _, wget := range d.wgets {
		_ = wget.Get()
	}
	return nil
}

func StartupCronDownload(dl Downloader) {
	_ = dl.Download()
	c := cron.New()
	err := c.AddFunc(dl.CrontabSchedule(), func() {
		_ = dl.Download()
	})
	if err != nil {
		panic(err)
	}
	c.Start()
}
