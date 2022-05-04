package cron

import (
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/download"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/antchfx/htmlquery"
	"sync"
)

const (
	HunterWebUrl           = `https://hunter.readthedocs.io/`
	HunterPackages         = `en/latest/packages/`
	HunterAllPackages      = `all.html`
	HunterPackageDetailFmt = `pkg/%s.html`
	HunterFixMe            = `https://__FIXME__`
)

type HunterPackage struct {
	Name     string
	Homepage string
	Detail   *download.GitDetail
}

type HunterFetcher struct {
	*core.BaseAsyncCronFetcher
}

func (h *HunterFetcher) Fetch() ([]*core.LibInfo, error) {
	packages := h.FetchAllPackages(HunterWebUrl + HunterPackages + HunterAllPackages)
	if len(packages) <= 0 {
		return nil, nil
	}
	infos := make([]*core.LibInfo, 0, len(packages))
	for _, hunterPackage := range packages {
		if hunterPackage == nil {
			continue
		}
		if hunterPackage.Detail == nil {
			infos = append(infos, &core.LibInfo{Name: hunterPackage.Name})
			continue
		}
		if len(hunterPackage.Detail.Tags) <= 0 {
			infos = append(infos, &core.LibInfo{
				Name: hunterPackage.Name,
				VerDetail: &core.LibVer{
					Ver:     "",
					License: hunterPackage.Detail.License,
				},
				Homepage:     &hunterPackage.Homepage,
				Author:       hunterPackage.Detail.Owner,
				Stars:        &hunterPackage.Detail.Star,
				Watching:     &hunterPackage.Detail.Watch,
				ForkCount:    &hunterPackage.Detail.Fork,
				Contributors: hunterPackage.Detail.Contributors,
			})
			continue
		}
		for _, tag := range hunterPackage.Detail.Tags {
			infos = append(infos, &core.LibInfo{
				Name: hunterPackage.Name,
				VerDetail: &core.LibVer{
					Ver:     tag.Ver,
					License: hunterPackage.Detail.License,
				},
				Homepage:     &hunterPackage.Homepage,
				Author:       hunterPackage.Detail.Owner,
				Stars:        &hunterPackage.Detail.Star,
				Watching:     &hunterPackage.Detail.Watch,
				ForkCount:    &hunterPackage.Detail.Fork,
				Contributors: hunterPackage.Detail.Contributors,
				SourceCode:   &tag.Zip,
			})
		}
	}
	return infos, nil
}

func (h *HunterFetcher) Name() string {
	return "hunter"
}

func NewHunterFetcher() *HunterFetcher {
	return &HunterFetcher{BaseAsyncCronFetcher: &core.BaseAsyncCronFetcher{}}
}

func (h *HunterFetcher) FetchAllPackages(url string) []*HunterPackage {
	doc := util.HttpGETNode(url)
	// 获取所有的package name
	treeWrap := htmlquery.FindOne(doc, `//div[@class='toctree-wrapper compound'])`)
	if treeWrap == nil {
		return nil
	}
	nodes := htmlquery.Find(treeWrap, `//li[@class='toctree-l1']`)
	packages := make([]*HunterPackage, len(nodes))
	wg := sync.WaitGroup{}
	for i, node := range nodes {
		// 获取包名
		packages[i] = &HunterPackage{
			Name: htmlquery.InnerText(htmlquery.FindOne(node, `/a/text()`)),
		}
		// 获取git page或者其他网站
		wg.Add(1)
		core.GlobalPool.Do(&pool.TaskHandler{
			Fn: func(idx interface{}) error {
				defer wg.Done()

				index := idx.(int)
				detailUrl := fmt.Sprintf(HunterWebUrl+HunterPackages+HunterPackageDetailFmt, packages[index].Name)
				detail := util.HttpGETNode(detailUrl)
				if detail == nil {
					return nil
				}
				one := htmlquery.FindOne(detail, `//ul[@class='simple']`)
				if one == nil {
					return nil
				}
				findOne := htmlquery.FindOne(one, `/li[1]/a/@href`)
				if findOne == nil {
					return nil
				}
				gitPage := htmlquery.InnerText(findOne)
				if gitPage == HunterFixMe {
					return nil
				}
				packages[index].Homepage = gitPage
				packages[index].Detail = download.GetRepoDetailByUrl(gitPage)
				return nil
			},
			Param: i,
		})
	}
	wg.Wait()
	return packages
}
