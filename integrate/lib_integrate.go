package integrate

import (
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"log"
	"strings"
	"sync"
	"time"
)

type Integrator interface {
	core.CronWorker
	AddHolder(holder core.ItemHolder)
	StartCronFetch() // 启动所有的holder抓取
	Integrate()
	Items() []*core.LibInfo
}

type LibIntegrator struct {
	holders []core.ItemHolder
	items   []*core.LibInfo // 已经做好聚合的item
	lock    sync.RWMutex
}

func NewLibIntegrator() *LibIntegrator {
	return &LibIntegrator{}
}

func (i *LibIntegrator) Items() []*core.LibInfo {
	i.lock.RLock()
	defer i.lock.RUnlock()
	return i.items
}

func (i *LibIntegrator) Integrate() {
	start := time.Now()
	log.Println("Start Integrate!")
	// 获取所有的libs
	recorder := make(map[string]*core.LibInfo, 4096)
	for _, holder := range i.holders {
		libs := holder.Items()
		for _, lib := range libs {
			// 查找是否出现过的，防止漏查，单词全部小些查找
			unique := strings.ToLower(lib.Unique())
			v, find := recorder[unique]
			if !find { // 没有出现过，直接插入
				recorder[unique] = lib
				continue
			}
			// 已经出现过，做数据聚合
			if holder.Name() == "conan" { // 以conan的信息为标准
				recorder[unique] = i.mergeLibs(lib, v)
			} else {
				recorder[unique] = i.mergeLibs(v, lib)
			}
		}
	}
	i.lock.Lock()
	defer i.lock.Unlock()
	i.items = make([]*core.LibInfo, 0, len(recorder))
	for _, info := range recorder {
		i.items = append(i.items, info)
	}
	log.Printf("End Integrate, cost=>%v", time.Since(start))
}

func (i *LibIntegrator) mergeLibs(l1 *core.LibInfo, l2 *core.LibInfo) *core.LibInfo {
	// 一项一项比较，均以第一个为标准
	//1. Name Pass
	//2. Version Pass
	//3.DownloadCount
	if l1.DownloadCount == 0 && l2.DownloadCount != 0 {
		l1.DownloadCount = l2.DownloadCount
	}
	//4.Description
	if l1.Description == "" && l2.Description != "" {
		l1.Description = l2.Description
	}
	//5.Homepage
	if l1.Homepage == nil && l2.Homepage != nil {
		l1.Homepage = l2.Homepage
	}
	//6.SourceCode
	if l1.SourceCode == nil && l2.SourceCode != nil {
		l1.SourceCode = l2.SourceCode
	}
	//7.Dependencies
	if len(l1.Dependencies) <= 0 && len(l2.Dependencies) > 0 {
		l1.Dependencies = l2.Dependencies
	} else if len(l2.Dependencies) > 0 {
		record := make(map[string]*core.LibDep, len(l1.Dependencies))
		for _, dep := range l1.Dependencies {
			if dep.Name == "" {
				continue
			}
			// 小写，防止漏查
			key := strings.ToLower(dep.Name)
			record[key] = dep
		}
		for _, dep := range l2.Dependencies {
			// 小写，防止漏查
			key := strings.ToLower(dep.Name)
			find, ok := record[key]
			if !ok { // 没有出现过
				l1.Dependencies = append(l1.Dependencies, dep)
			} else { //出现过，检查版本，以第一个为准
				if find.Version == "" && dep.Version != "" {
					find.Version = dep.Version
				}
			}
		}
	}
	//8.Author
	if l1.Author == "" && l2.Author != "" {
		l1.Author = l2.Author
	}
	//9.Contributors
	if len(l1.Contributors) <= 0 && len(l2.Contributors) > 0 {
		l1.Contributors = l2.Contributors
	} else if len(l2.Contributors) > 0 {
		record := make(map[string]bool)
		for _, contributor := range l1.Contributors {
			key := strings.ToLower(contributor)
			record[key] = true
		}
		for _, contributor := range l2.Contributors {
			key := strings.ToLower(contributor)
			// 出现过
			if _, ok := record[key]; ok {
				continue
			}
			l1.Contributors = append(l1.Contributors, contributor)
		}
	}
	//10.Stars
	if l1.Stars == nil && l2.Stars != nil {
		l1.Stars = l2.Stars
	}
	//11.Watching
	if l1.Watching == nil && l2.Watching != nil {
		l1.Watching = l2.Watching
	}
	//12.ForkCount
	if l1.ForkCount == nil && l2.ForkCount != nil {
		l1.ForkCount = l2.ForkCount
	}
	return l1
}

func (i *LibIntegrator) StartCronFetch() {
	if len(i.holders) <= 0 {
		panic("holders is empty")
	}
	for idx := range i.holders {
		i.holders[idx].Startup()
	}
}

func (i *LibIntegrator) CrontabSchedule() string {
	return "0 40 0 * * ?" // 每天0点40分聚合
}

func (i *LibIntegrator) AddHolder(holder core.ItemHolder) {
	i.holders = append(i.holders, holder)
}
