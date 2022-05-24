package test

import (
	"encoding/json"
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/db"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"log"
	"os/exec"
	"testing"
	"time"
)

func Test_cmd_run(t *testing.T) {
	cmd := exec.Command("git", "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	println(string(output))
}

const (
	mysqlUsername = `root`
	mysqlPassword = `jiangaoran`
	mysqlDatabase = `bs`
)

func Test_dep(t *testing.T) {
	// 初始化数据库连接
	dbConn := db.InitConn(mysqlUsername, mysqlPassword, mysqlDatabase, nil, nil)
	//log.Println("Database init connection success")
	log.Printf("开始对 %v 进行依赖分析...", "boost")
	now := time.Now()
	time.Sleep(12 * time.Second)
	var tt *db.TBsLibInfo
	first := dbConn.Model(db.TBsLibInfo{}).Where("name = ? and dependencies <> ''", "boost").First(&tt)
	if first.Error != nil || tt == nil {
		panic(first.Error)
	}
	var ttt []*core.LibDep
	err := json.Unmarshal(util.Str2Bytes(tt.Dependencies), &ttt)
	if err != nil {
		panic(err)
	}
	deps := make([]string, 0, len(ttt))
	for _, dep := range ttt {
		deps = append(deps, dep.Name)
	}
	log.Printf("%v 分析出boost的依赖为 => %v\n", "boost", deps)
	log.Printf("%v 分析耗时=>%v", "boost", time.Since(now))
}
