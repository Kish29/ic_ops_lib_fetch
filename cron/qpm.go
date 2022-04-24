package cron

import (
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/antchfx/htmlquery"
	"strings"
)

const (
	QPMWebUrl            = `https://www.qpm.io/`
	QPMPackagesUrl       = `packages/index.html`
	QPMPackageInfoUrlFmt = `packages/%s/index.html`
)

type QPMFetcher struct {
	*core.BaseAsyncCronFetcher
	workers *pool.WorkPool
}

func (Q *QPMFetcher) Fetch() ([]*core.LibInfo, error) {
	//allPackages := ParseQPM(QPMPackagesUrl)
	return nil, nil
}

func (Q *QPMFetcher) Name() string {
	return "QPM"
}

type QPMPackage struct {
	Name    string
	Url     string
	Desc    string
	Version string
	License string
}

func ParseQPM(url string) []*QPMPackage {
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
			pkg.Url = htmlquery.InnerText(pkgUrl)
		}
		if pkgName := htmlquery.FindOne(pkgInfo, `/strong/text()`); pkgName != nil {
			pkg.Name = pkgName.Data
		}
		if pkgDesc := htmlquery.FindOne(node, `//p/text()`); pkgDesc != nil {
			pkg.Desc = pkgDesc.Data
		}
		if pkgVer := htmlquery.Find(node, `//small`); len(pkgVer) > 0 && pkgVer[1] != nil {
			pkg.Version = htmlquery.InnerText(pkgVer[1])
		}
		if pkgLicense := htmlquery.FindOne(node, `//div[@class='right']/text()`); pkgLicense != nil {
			pkg.License = strings.TrimSpace(strings.ReplaceAll(pkgLicense.Data, `\n`, ``))
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}
