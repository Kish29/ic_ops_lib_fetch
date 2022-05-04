package db

import (
	"encoding/json"
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/integrate"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"sync"
	"time"
)

const tableName = `t_bs_lib_info`

type BaseDatabaseUpdater struct {
	integrator integrate.Integrator
	dbConn     *gorm.DB
}

func (d *BaseDatabaseUpdater) MustTableExit() {
	exist := d.dbConn.Migrator().HasTable(tableName)
	if !exist {
		err := d.dbConn.AutoMigrate(&TBsLibInfo{})
		if err != nil {
			panic(fmt.Errorf("table %s not exist, and create error occurred, error=>%v", tableName, err))
		}
	}
}

func NewBaseDatabaseUpdater(integrator integrate.Integrator, dbConn *gorm.DB) *BaseDatabaseUpdater {
	if integrator == nil || dbConn == nil {
		panic("integrator or dbConn is nil")
	}
	return &BaseDatabaseUpdater{integrator: integrator, dbConn: dbConn}
}

func (d *BaseDatabaseUpdater) UpdateIntoDB() {
	if d.integrator == nil {
		panic("integrator is nil")
	}
	start := time.Now()
	// 获取所有的libInfo
	items := d.integrator.Items()
	// 需要更新的item
	needUpdate := make([]*TBsLibInfo, 0, len(items))
	// 需要插入的item
	needInsert := make([]*TBsLibInfo, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		// 构建数据库模型
		dbBean := libInfoConvert2DbBean(item)
		var info = TBsLibInfo{}
		var first *gorm.DB
		if item.VerDetail != nil {
			first = d.dbConn.Where("name=? and version=?", item.Name, item.VerDetail.Ver).First(&info)
		} else {
			first = d.dbConn.Where("name=?", item.Name).First(&info)
		}
		// 如果没有找到该lib的记录
		if first.Error != nil && first.Error == gorm.ErrRecordNotFound {
			// 插入
			needInsert = append(needInsert, dbBean)
		} else { // 更新
			dbBean.Id = info.Id
			dbBean.CreateTime = info.CreateTime
			needUpdate = append(needUpdate, dbBean)
		}
	}
	// 插入与更新
	wg := sync.WaitGroup{}
	for _, info := range needInsert {
		if info == nil {
			continue
		}
		wg.Add(1)
		core.GlobalPool.Do(&pool.TaskHandler{
			Fn: func(i interface{}) error {
				insert := i.(*TBsLibInfo)
				create := d.dbConn.Model(&TBsLibInfo{}).Create(insert)
				if create.Error != nil {
					log.Printf("[error] create error, error=>%v", create.Error)
				}
				return nil
			},
			Param: info,
		})
	}
	for _, info := range needUpdate {
		if info == nil {
			continue
		}
		wg.Add(1)
		core.GlobalPool.Do(&pool.TaskHandler{
			Fn: func(i interface{}) error {
				defer wg.Done()

				update := i.(*TBsLibInfo)
				save := d.dbConn.Model(&TBsLibInfo{}).Updates(update)
				if save.Error != nil {
					log.Printf("[error] update error, error=>%v", save.Error)
				}
				return nil
			},
			Param: info,
		})
	}
	wg.Wait()
	log.Printf("End database update, cost=>%v", time.Since(start))
}

func (d *BaseDatabaseUpdater) CrontabSchedule() string {
	return "0 50 0 * * ?" // 每天50分写入数据库
}

func libInfoConvert2DbBean(info *core.LibInfo) *TBsLibInfo {
	dbBean := &TBsLibInfo{
		Name:          info.Name,
		DownloadCount: info.DownloadCount,
		Description:   info.Description,
		Author:        info.Author,
		BaseDBMod: BaseDBMod{
			Id:         rand.Uint64(),
			CreateTime: time.Now(),
			UpdateTime: time.Now(),
		},
	}
	var err error
	var jsonM []byte
	// 1. version detail
	if info.VerDetail != nil {
		dbBean.Version = info.VerDetail.Ver
		dbBean.License = info.VerDetail.License
	}
	// 2. dependencies
	if len(info.Dependencies) > 0 {
		jsonM, err = json.Marshal(info.Dependencies)
		if err != nil {
			log.Printf("[error] json marshal lib=>%s Dependencies error, error=>%v", info.Name, err)
		} else {
			dbBean.Dependencies = util.Bytes2Str(jsonM)
		}
	}
	// 4. contributors
	if len(info.Contributors) > 0 {
		jsonM, err = json.Marshal(info.Contributors)
		if err != nil {
			log.Printf("[error] json marshal lib=>%s Contributors error, error=>%v", info.Name, err)
		} else {
			dbBean.Contributors = util.Bytes2Str(jsonM)
		}
	}
	if info.Homepage != nil {
		dbBean.Homepage = *info.Homepage
	}
	if info.SourceCode != nil {
		dbBean.SourceCode = *info.SourceCode
	}
	if info.Stars != nil {
		dbBean.Stars = *info.Stars
	}
	if info.Watching != nil {
		dbBean.Watching = *info.Watching
	}
	if info.ForkCount != nil {
		dbBean.ForkCount = *info.ForkCount
	}
	return dbBean
}
