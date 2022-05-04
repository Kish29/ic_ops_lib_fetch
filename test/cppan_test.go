package test

import (
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/antchfx/htmlquery"
	"github.com/gookit/goutil/mathutil"
	"golang.org/x/net/html"
	"log"
	"strings"
	"sync"
	"testing"
)

const (
	CppanWebUrl           = `https://cppget.org/`
	CppanPageFmt          = `https://cppget.org/?packages&p=%d`
	CppanPackageDetailFmt = `%s/%s`
)

var workers *pool.WorkPool = pool.New(256)

type CppanPackage struct {
	Name         string
	Version      string
	License      string
	Homepage     string
	Dependencies []*core.LibDep
	SourceCode   string
	Desc         string
}

func ParseCppanFromDetails(urls []string) []*CppanPackage {
	wg := sync.WaitGroup{}
	lock := sync.RWMutex{}
	packages := make([]*CppanPackage, 0, len(urls))
	for _, url := range urls {
		wg.Add(1)
		workers.Do(&pool.TaskHandler{
			Fn: func(i interface{}) error {
				defer wg.Done()

				du := i.(string)
				doc := util.HttpGETNode(du)
				if doc == nil {
					return nil
				}
				// 解析所有的version
				vers := htmlquery.Find(doc, `//table[@class='proplist version']`)
				if len(vers) <= 0 {
					return nil
				}
				pkgs := make([]*CppanPackage, 0, len(vers))
				for _, ver := range vers {
					v := htmlquery.FindOne(ver, `/tbody/tr[1]/td/span/a/text()`).Data
					detail := ParseCppanDetail(du[strings.LastIndex(du, `/`)+1:], v)
					if detail != nil {
						pkgs = append(pkgs, detail)
					}
				}
				lock.Lock()
				packages = append(packages, pkgs...)
				lock.Unlock()
				return nil
			},
			Param: url,
		})
	}
	wg.Wait()
	log.Printf("cppan get all %d detailed packages", len(packages))
	return packages
}

func ParseCppanDetail(name, version string) *CppanPackage {
	detailUrl := fmt.Sprintf(CppanWebUrl+CppanPackageDetailFmt, name, version)
	doc := util.HttpGETNode(detailUrl)
	if doc == nil {
		return nil
	}
	p := &CppanPackage{Name: name, Version: version}
	// 解析description
	descNode := htmlquery.FindOne(doc, `//div[@id='description']`)
	if descNode != nil {
		desc := htmlquery.FindOne(descNode, `/pre/text()`)
		if desc != nil {
			index := strings.Index(desc.Data, `For more information see`)
			if index != -1 {
				p.Desc = desc.Data[:index]
			} else {
				p.Desc = desc.Data
			}
		}
	}
	// 解析license
	licenseNode := htmlquery.FindOne(doc, `//tr[@class='license']`)
	if licenseNode != nil {
		license := htmlquery.FindOne(licenseNode, `//span[@class='value']`)
		if license != nil {
			p.License = htmlquery.InnerText(license)
		}
	}
	// 解析source code
	downloadNode := htmlquery.FindOne(doc, `//tr[@class='download']`)
	if downloadNode != nil {
		download := htmlquery.FindOne(downloadNode, `//span[@class='value']`)
		if download != nil {
			p.SourceCode = htmlquery.InnerText(htmlquery.FindOne(download, `/a/@href`))
		}
	}
	// 解析homepage
	homepageNode := htmlquery.FindOne(doc, `//tr[@class='url']`)
	if homepageNode != nil {
		homepage := htmlquery.FindOne(homepageNode, `//span[@class='value']`)
		if homepage != nil {
			p.Homepage = htmlquery.InnerText(htmlquery.FindOne(homepage, `/a/@href`))
		}
	}
	// 解析dependencies
	depsNodes := htmlquery.Find(doc, `//tr[@class='depends']`)
	if len(depsNodes) > 0 {
		for _, node := range depsNodes {
			depName := htmlquery.FindOne(node, `/td/span/a[1]/text()`)
			depVer := htmlquery.FindOne(node, `/td/span/a[2]/text()`)
			var dep *core.LibDep
			if depName != nil {
				dep = &core.LibDep{Name: depName.Data}
				if depVer != nil {
					dep.Version = depVer.Data
				}
			}
			if dep != nil {
				p.Dependencies = append(p.Dependencies, dep)
			}
		}
	}
	return p
}

func CppanGetNameUrl(packageNode *html.Node) string {
	// 查找class为name的tr
	nameNode := htmlquery.FindOne(packageNode, `//tr[@class='name']`)
	if nameNode == nil {
		return ""
	}
	one := htmlquery.FindOne(nameNode, `//span[@class='value']`)
	if one == nil {
		return ""
	}
	return htmlquery.InnerText(one)
}

func CppanFetchAll(url string) []*CppanPackage {
	// 获取根目录
	doc := util.HttpGETNode(url)
	// 获取包总数量
	counterStr := htmlquery.InnerText(htmlquery.FindOne(doc, `//div[@id='count']`))
	counter := mathutil.MustInt(counterStr[:strings.Index(counterStr, ` `)])
	// 获取总页数
	pageNum := counter / 20
	if counter%20 != 0 {
		pageNum++
	}
	// 第一页直接获取到了
	packagesUrls := make([]string, 0, counter)
	nodes := htmlquery.Find(doc, `//table[@class='proplist package']`)
	for _, node := range nodes {
		nu := CppanGetNameUrl(node)
		if nu == "" {
			continue
		}
		packagesUrls = append(packagesUrls, CppanWebUrl+nu)
	}
	// 从第二页开始
	wg := sync.WaitGroup{}
	lock := sync.RWMutex{}
	for i := 1; i <= pageNum; i++ {
		wg.Add(1)
		workers.Do(&pool.TaskHandler{
			Fn: func(i interface{}) error {
				defer wg.Done()

				index := i.(int)
				pageUrl := fmt.Sprintf(CppanPageFmt, index)
				node := util.HttpGETNode(pageUrl)
				if node == nil {
					return nil
				}
				pkgs := htmlquery.Find(node, `//table[@class='proplist package']`)
				if len(pkgs) <= 0 {
					return nil
				}
				lock.Lock()
				for _, pkg := range pkgs {
					name := CppanGetNameUrl(pkg)
					if name == "" {
						continue
					}
					packagesUrls = append(packagesUrls, CppanWebUrl+name)
				}
				lock.Unlock()
				return nil
			},
			Param: i,
		})
	}
	wg.Wait()
	log.Printf("cppan get %d packages\n", len(packagesUrls))
	return ParseCppanFromDetails(packagesUrls)
}

func Test_cppan(t *testing.T) {
	doc := util.HttpGETNode("https://cppget.org/")
	counterStr := htmlquery.InnerText(htmlquery.FindOne(doc, `//div[@id='count']`))
	spaceIdx := strings.Index(counterStr, ` `)
	counter := mathutil.MustInt(counterStr[:spaceIdx])
	println(counter)
	pageNum := counter / 20
	if counter%20 != 0 {
		pageNum++
	}
	println(pageNum)
	nodes := htmlquery.Find(doc, `//table[@class='proplist package']`)
	println(len(nodes))
}

func Test_all(t *testing.T) {
	all := CppanFetchAll(CppanWebUrl)
	for _, cppanPackage := range all {
		fmt.Printf("%v\n", cppanPackage)
	}
}

func Test_cppan_detail(t *testing.T) {
	fmt.Printf("%v", ParseCppanDetail("bbot", `0.14.0`))
}

func Test_cppan_vers(t *testing.T) {
	ParseCppanFromDetails([]string{"https://cppget.org/libsqlite3"})
}
