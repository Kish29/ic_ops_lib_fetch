package cron

import (
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/antchfx/htmlquery"
	"github.com/gookit/goutil/mathutil"
	"strings"
	"sync"
)

const (
	QPMWebUrl            = `https://www.qpm.io/`
	QPMPackagesUrl       = `packages/index.html`
	QPMPackageInfoUrlFmt = `packages/%s/index.html`
)

const (
	QPMVerTag = `collection-item `
)

type QPMFetcher struct {
	*core.BaseAsyncCronFetcher
	workers *pool.WorkPool
}

func NewQPMFetcher() *QPMFetcher {
	return &QPMFetcher{BaseAsyncCronFetcher: &core.BaseAsyncCronFetcher{}, workers: pool.New(128)}
}

func (Q *QPMFetcher) Fetch() ([]*core.LibInfo, error) {
	// 获取所有包
	qpmPackages := Q.FetchAllPackages(QPMWebUrl + QPMPackagesUrl)
	// 解析详情
	Q.ParseQPMDetail(qpmPackages)
	// 构建libInfo
	allInfos := make([]*core.LibInfo, 0, len(qpmPackages)*2)
	for _, qpmPackage := range qpmPackages {
		if qpmPackages == nil {
			continue
		}
		infoList := make([]*core.LibInfo, 0, len(qpmPackage.Versions))
		for _, version := range qpmPackage.Versions {
			infoList = append(infoList, &core.LibInfo{
				Name: qpmPackage.Name,
				VerDetail: &core.LibVer{
					Ver:     version,
					License: qpmPackage.License,
				},
				Description:   qpmPackage.Desc,
				Homepage:      &qpmPackage.GitHomepage,
				DownloadCount: qpmPackage.DownloadCount,
				Author:        qpmPackage.Author,
			})
		}
		allInfos = append(allInfos, infoList...)
	}
	return allInfos, nil
}

func (Q *QPMFetcher) Name() string {
	return "QPM"
}

type QPMPackage struct {
	Name          string
	QPMDetailUrl  string
	Desc          string
	Versions      []string
	License       string
	Author        string
	GitHomepage   string
	DownloadCount int
}

func (Q *QPMFetcher) FetchAllPackages(url string) []*QPMPackage {
	doc := util.HttpGETNode(url)
	nodes := htmlquery.Find(doc, `//li[@class="collection-item"]`)
	pkgs := make([]*QPMPackage, 0, len(nodes))
	for _, node := range nodes {
		pkgInfo := htmlquery.FindOne(node, `//a[@class="orange-text"]`)
		if pkgInfo == nil {
			continue
		}
		pkg := &QPMPackage{}
		if pkgUrl := htmlquery.FindOne(pkgInfo, `/@href`); pkgUrl != nil {
			pkg.QPMDetailUrl = htmlquery.InnerText(pkgUrl)
		}
		if pkgName := htmlquery.FindOne(pkgInfo, `/strong/text()`); pkgName != nil {
			pkg.Name = pkgName.Data
		}
		if pkgDesc := htmlquery.FindOne(node, `//p/text()`); pkgDesc != nil {
			pkg.Desc = pkgDesc.Data
		}
		if pkgVer := htmlquery.Find(node, `//small`); len(pkgVer) > 0 {
			if pkgVer[0] != nil {
				pkg.Author = htmlquery.InnerText(pkgVer[0])
			}
			if pkgVer[1] != nil {
				pkg.Versions = append(pkg.Versions, htmlquery.InnerText(pkgVer[1]))
			}
		}
		if pkgLicense := htmlquery.FindOne(node, `//div[@class='right']/text()`); pkgLicense != nil {
			pkg.License = strings.TrimSpace(strings.ReplaceAll(pkgLicense.Data, `\n`, ``))
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

func (Q *QPMFetcher) ParseQPMDetail(qpmPackages []*QPMPackage) {
	if len(qpmPackages) <= 0 {
		return
	}
	wg := sync.WaitGroup{}
	for i := range qpmPackages {
		if qpmPackages[i].Name == "" {
			continue
		}
		wg.Add(1)
		Q.workers.Do(&pool.TaskHandler{
			Fn: func(i interface{}) error {
				defer wg.Done()

				idx := i.(int)
				pkg := qpmPackages[idx]
				url := fmt.Sprintf(QPMWebUrl+QPMPackageInfoUrlFmt, pkg.Name)
				doc := util.HttpGETNode(url)
				if doc == nil {
					return nil
				}
				// 获取所有的版本号
				verList := htmlquery.Find(doc, fmt.Sprintf(`//a[@class='%s']`, QPMVerTag))
				if verList == nil {
					return nil
				}
				if len(verList) <= 0 {
					return nil
				}
				// 去重
				record := make(map[string]bool)
				for _, ver := range pkg.Versions {
					record[ver] = true
				}
				// 添加所有的版本号
				for _, ver := range verList {
					v := htmlquery.FindOne(ver, `/text()`)
					if v != nil && record[v.Data] {
						continue
					}
					pkg.Versions = append(pkg.Versions, v.Data)
				}
				// 获取githubHomepage
				gitPage := htmlquery.FindOne(doc, `//a[@class='collection-item']`)
				if gitPage != nil {
					page := htmlquery.InnerText(htmlquery.FindOne(gitPage, `/@href`))
					index := strings.Index(page, `?`)
					if index != -1 {
						pkg.GitHomepage = page[:index]
					} else {
						pkg.GitHomepage = page
					}
				}
				// 解析下载总次数
				one := htmlquery.FindOne(doc, `//td[@id='total_stat']/text()`)
				if one != nil {
					pkg.DownloadCount = mathutil.MustInt(one.Data)
				}
				return nil
			},
			Param: i,
		})
	}
	wg.Wait()
}
