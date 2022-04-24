package test

import (
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/cron"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/antchfx/htmlquery"
	"strings"
	"sync"
	"testing"
)

type QPMPackage struct {
	Name     string
	Url      string
	Desc     string
	Versions []string
	License  string
	Author   string
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

func Test_ParseVersion(t *testing.T) {
	doc := util.HttpGETNode(`https://www.qpm.io/packages/android.native.pri/index.html`)
	if doc == nil {
		panic("doc is nil")
	}
	info := htmlquery.Find(doc, `//div[@class='collection']`)
	if info == nil {
		panic("info is nil")
	}
	if info[0] != nil {
		// github info
	}
	if info[1] != nil {
		// version info
		vers := htmlquery.Find(info[1], `//a`)
		for _, ver := range vers {
			println(htmlquery.InnerText(ver))
		}
	}
}

func ParseQPMVersions(qpmPackages []*QPMPackage) {
	if len(qpmPackages) <= 0 {
		return
	}
	workers := pool.New(64)
	wg := sync.WaitGroup{}
	for i := range qpmPackages {
		if qpmPackages[i].Name == "" {
			continue
		}
		wg.Add(1)
		workers.Do(&pool.TaskHandler{
			Fn: func(i interface{}) error {
				defer wg.Done()

				idx := i.(int)
				pkg := qpmPackages[idx]
				url := fmt.Sprintf(cron.QPMWebUrl+cron.QPMPackageInfoUrlFmt, pkg.Name)
				doc := util.HttpGETNode(url)
				if doc == nil {
					return nil
				}
				verList := htmlquery.Find(doc, `//a[@class='collection-item']`)
				if verList == nil {
					return nil
				}
				if len(verList) <= 0 {
					return nil
				}
				record := make(map[string]bool)
				for _, ver := range pkg.Versions {
					record[ver] = true
				}
				for _, ver := range verList {
					v := htmlquery.FindOne(ver, `/text()`)
					if v != nil && !record[v.Data] {
						continue
					}
					pkg.Versions = append(pkg.Versions, v.Data)
				}
				return nil
			},
			Param: i,
		})
	}
	wg.Wait()
}

func Test_html_query(t *testing.T) {
	qpmPackages := ParseQPM(cron.QPMWebUrl + cron.QPMPackagesUrl)
	ParseQPMVersions(qpmPackages)
	for _, qpmPackage := range qpmPackages {
		fmt.Printf("%v\n", qpmPackage)
	}
}
