package integrate

import "github.com/robfig/cron"

func Startup(integrator Integrator) {
	if integrator == nil {
		panic("integrator is nil")
	}
	// 开始所有网站的抓取
	integrator.StartCronFetch()
	// 等待所有的信息抓取完后，第一次聚合
	integrator.Integrate()
	// 开启聚合的定时定时
	schedule := integrator.CrontabSchedule()
	c := cron.New()
	err := c.AddFunc(schedule, integrator.Integrate)
	if err != nil {
		panic(err)
	}
	c.Start()
}
