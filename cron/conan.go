package cron

import (
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/go-resty/resty/v2"
	"log"
	"strings"
	"sync"
)

const (
	webUrlConan          = `https://conan.io`
	allPackagesPathConan = `/center/api/ui/allpackages`
	// ?name=7zip&user=_&channel=_
	packageDetailPathConan = `/center/api/ui/details`
	// name=7zip&version=19.00&user=_&channel=_
	packageRevisionPathConan = `/center/api/ui/revisions?`
	// name=7zip&version=19.00&user=_&channel=_&revision=7e1fae98e076cf075bea743444b918dd
	packageDependenciesPathConan = `/center/api/ui/dependencies?`
)

var (
	conanClient = resty.New()
)

type ConanPackage struct {
	Name          string `json:"name"`
	LatestVersion string `json:"latest_version"`
	User          string `json:"user"`
	Channel       string `json:"channel"`
	License       string `json:"license"`
	Description   string `json:"description"`
	Topics        string `json:"topics"`
}

type ConanPackageVersion struct {
	Version string `json:"version"`
	License string `json:"license"`
}

type ConanDependency struct {
	NameVersion string `json:"name_version"`
}

type ConanAllPackagesResp struct {
	Count    int             `json:"count"`
	Packages []*ConanPackage `json:"packages"`
}

type ConanPackageDetailResp struct {
	LatestVersion  string                 `json:"latest_version"`
	Versions       []*ConanPackageVersion `json:"versions"`
	Downloads      int                    `json:"downloads"`
	Homepage       string                 `json:"homepage"`
	SourceLocation string                 `json:"source_location"`
	Topics         string                 `json:"topics"`
}

type ConanRevisionsResp struct {
	Count     int      `json:"count"`
	Revisions []string `json:"revisions"`
}

type ConanDependenciesResp struct {
	Dependencies []*ConanDependency `json:"dependencies"`
}

type ConanFetcher struct {
	*core.BaseAsyncCronFetcher
	workers  *pool.WorkPool
	workers2 *pool.WorkPool
}

func NewConanFetcher() *ConanFetcher {
	fetcher := &ConanFetcher{BaseAsyncCronFetcher: &core.BaseAsyncCronFetcher{}}
	fetcher.workers = pool.New(128)
	fetcher.workers2 = pool.New(128)
	return fetcher
}

func (c *ConanFetcher) Fetch() ([]*core.LibInfo, error) {
	// 获取所有的包预览信息
	defaultHeaderAttr := map[string]string{
		util.HttpHeadKeyUserAgent: util.RandomFakeAgent(),
	}
	resp := &ConanAllPackagesResp{}
	err := util.HttpGETToJson(
		conanClient,
		webUrlConan+allPackagesPathConan,
		nil,
		defaultHeaderAttr,
		resp,
	)
	if err != nil || resp.Packages == nil {
		return nil, err
	}
	// 根据所有的包获取详细信息
	packages := resp.Packages
	log.Printf("Fetcher=>%s will update %d packages!", c.Name(), len(packages))
	libInfo := make([]*core.LibInfo, 0, 4*len(packages))
	wg := sync.WaitGroup{}
	lock := sync.RWMutex{}
	for idx := range packages {
		if packages[idx] == nil {
			continue
		}
		wg.Add(1)
		c.workers.Do(&pool.TaskHandler{
			Fn: func(i interface{}) error {
				defer wg.Done()
				index := i.(int)
				details := c.libInfoDetails(packages[index])
				if len(details) <= 0 {
					return nil
				}
				lock.Lock()
				for _, detail := range details {
					if detail == nil {
						continue
					}
					libInfo = append(libInfo, detail)
				}
				lock.Unlock()
				return nil
			},
			Param: idx,
		})
	}
	wg.Wait()
	return libInfo, nil
}

func (c *ConanFetcher) Name() string {
	return "conan"
}

func (c *ConanFetcher) fulfillLinInfoDetail(libInfo *core.LibInfo, user, channel string) {
	if libInfo == nil {
		return
	}
	headerAttr := map[string]string{
		util.HttpHeadKeyUserAgent: util.RandomFakeAgent(),
	}
	param := map[string]string{
		"name":    libInfo.Name,
		"user":    user,
		"channel": channel,
	}
	// 2. 获取revision，参数: 最新版本
	revisions := &ConanRevisionsResp{}
	param["version"] = libInfo.VerDetail.Ver
	err := util.HttpGETToJson(
		conanClient,
		webUrlConan+packageRevisionPathConan,
		param,
		headerAttr,
		revisions,
	)
	if err != nil || len(revisions.Revisions) <= 0 {
		log.Printf("[error] Fethcer=>%s fetch into libInfo error=>%v or revision empty", c.Name(), err)
		return
	}
	// 3. 获取dependencies, 参数：第一个revision
	param["revision"] = revisions.Revisions[0]
	dependencies := &ConanDependenciesResp{}
	err = util.HttpGETToJson(
		conanClient,
		webUrlConan+packageDependenciesPathConan,
		param,
		headerAttr,
		dependencies,
	)
	// 4. 添加dependency
	for _, dependency := range dependencies.Dependencies {
		if len(dependency.NameVersion) <= 0 {
			continue
		}
		index := strings.Index(dependency.NameVersion, `/`)
		if index == -1 {
			libInfo.Dependencies = append(libInfo.Dependencies, &core.LibDep{Name: dependency.NameVersion, Version: dependency.NameVersion})
		} else {
			libInfo.Dependencies = append(libInfo.Dependencies, &core.LibDep{
				Name:    dependency.NameVersion[0:index],
				Version: dependency.NameVersion[index+1:],
			})
		}
	}
}

func (c *ConanFetcher) libInfoDetails(preInfo *ConanPackage) []*core.LibInfo {
	if preInfo == nil {
		return nil
	}
	defer log.Printf("Fetcher=>%s fetch libInfo=>%s done", c.Name(), preInfo.Name)
	headerAttr := map[string]string{
		util.HttpHeadKeyUserAgent: util.RandomFakeAgent(),
	}
	param := map[string]string{
		"name":    preInfo.Name,
		"user":    preInfo.User,
		"channel": preInfo.Channel,
	}
	// 1. 获取detail，得到所有的version
	detail := &ConanPackageDetailResp{}
	err := util.HttpGETToJson(
		conanClient,
		webUrlConan+packageDetailPathConan,
		param,
		headerAttr,
		detail,
	)
	if err != nil {
		log.Printf("[error] Fethcer=>%s fetch into libInfo error=>%v", c.Name(), err)
		return nil
	}
	// 2. 获取所有version的有关信息
	libs := make([]*core.LibInfo, len(detail.Versions))
	wg := sync.WaitGroup{}
	for i := range detail.Versions {
		if detail.Versions[i] == nil {
			continue
		}
		libs[i] = &core.LibInfo{}
		wg.Add(1)
		c.workers2.Do(&pool.TaskHandler{
			Fn: func(index interface{}) error {
				defer wg.Done()
				idx := index.(int)
				libs[idx].Name = preInfo.Name
				libs[idx].VerDetail = &core.LibVer{
					Ver:     detail.Versions[idx].Version,
					License: detail.Versions[idx].License,
				}
				libs[idx].Description = preInfo.Description
				libs[idx].Homepage = &detail.Homepage
				libs[idx].DownloadCount = detail.Downloads
				if detail.SourceLocation == "" {
					libs[idx].SourceCode = &detail.Homepage
				} else {
					libs[idx].SourceCode = &detail.Homepage
				}
				c.fulfillLinInfoDetail(libs[idx], preInfo.User, preInfo.Channel)
				return nil
			},
			Param: i,
		})
	}
	wg.Wait()
	return libs
}
