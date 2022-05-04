package test

import (
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/download"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/antchfx/htmlquery"
	"sync"
	"testing"
)

const (
	HunterWebUrl           = `https://hunter.readthedocs.io/`
	HunterPackages         = `en/latest/packages/`
	HunterAllPackages      = `all.html`
	HunterPackageDetailFmt = `pkg/%s.html`
	HunterFixMe            = `https://__FIXME__`
)

type HunterPackage struct {
	Name   string
	Detail *download.GitDetail
}

func HunterFetchAll(url string) []*HunterPackage {
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
		workers.Do(&pool.TaskHandler{
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
				packages[index].Detail = download.GetRepoDetailByUrl(gitPage)
				return nil
			},
			Param: i,
		})
	}
	wg.Wait()
	return packages
}

func Test_hunter(t *testing.T) {
	for _, hunterPackage := range HunterFetchAll(HunterWebUrl + HunterPackages + HunterAllPackages) {
		fmt.Printf("%v\n", hunterPackage)
	}
}

func Test_git_page(t *testing.T) {
	detail := util.HttpGETNode(fmt.Sprintf(HunterWebUrl+HunterPackages+HunterPackageDetailFmt, `ARM_NEON_2_x86_SSE`))
	if detail == nil {
		return
	}
	one := htmlquery.FindOne(detail, `//ul[@class='simple']`)
	if one == nil {
		return
	}
	findOne := htmlquery.FindOne(one, `/li[1]/a/@href`)
	if findOne == nil {
		return
	}
	gitUrl := htmlquery.InnerText(findOne)
	if gitUrl == HunterFixMe {
		return
	}
	// git detail
	fmt.Printf("%v\n", download.GetRepoDetailByUrl(gitUrl))
}
