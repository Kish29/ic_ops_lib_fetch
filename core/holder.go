package core

import (
	"github.com/robfig/cron"
	"log"
	"sync"
	"time"
)

type Identify interface {
	Unique() string
}

type ItemHolder interface {
	Name() string
	Items() []*LibInfo
	ItemByKey(key string) *LibInfo
	Startup()
}

type BaseItemHolder struct {
	AsyncCronFetcher
	storage map[string]*LibInfo
	locker  sync.RWMutex
}

func (b *BaseItemHolder) Name() string {
	return b.AsyncCronFetcher.Name()
}

func NewBaseItemHolder(asyncCronFetcher AsyncCronFetcher) *BaseItemHolder {
	return &BaseItemHolder{AsyncCronFetcher: asyncCronFetcher, storage: make(map[string]*LibInfo)}
}

func (b *BaseItemHolder) Items() []*LibInfo {
	if len(b.storage) <= 0 {
		return nil
	}
	libs := make([]*LibInfo, 0, len(b.storage))
	b.locker.RLock()
	for _, info := range b.storage {
		libs = append(libs, info)
	}
	b.locker.RUnlock()
	return libs
}

func (b *BaseItemHolder) ItemByKey(key string) *LibInfo {
	b.locker.RLock()
	defer b.locker.RUnlock()
	return b.storage[key]
}

func (b *BaseItemHolder) update() {
	start := time.Now()
	log.Printf("Star %s update\n", b.Name())
	fetch, err := b.Fetch()
	if err != nil || fetch == nil {
		log.Printf("Fetcher=>%s fetch data error, error=>%v", b.Name(), err)
		return
	}
	b.locker.Lock()
	defer b.locker.Unlock()
	b.storage = make(map[string]*LibInfo, len(fetch))
	for i := range fetch {
		if fetch[i] == nil {
			continue
		}
		b.storage[fetch[i].Unique()] = fetch[i]
	}
	log.Printf("End %s update, cost=>%v", b.Name(), time.Since(start))
}

func (b *BaseItemHolder) Startup() {
	if b.AsyncCronFetcher == nil {
		panic("Fetcher is nil!")
	}
	b.update()
	c := cron.New()
	schedule := b.CrontabSchedule()
	if schedule == "" {
		panic("Fetcher crontab empty")
	}
	err := c.AddFunc(schedule, b.update)
	if err != nil {
		panic(err)
	}
	c.Start()
}
