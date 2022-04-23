package test

import (
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/antchfx/htmlquery"
	"github.com/go-resty/resty/v2"
	"strings"
	"testing"
)

func Test_QPM(t *testing.T) {
	rc := resty.New()
	//get, err := util.HttpRawGET(rc, "https://a.nel.cloudflare.com/report/v3?s=U%2BdLDvKDMh%2FOdu%2Fx5RvH%2BjdzUKIP4AJQbEuVCPvppQTTWbix%2F7f9Ml2HGHikW617ZSdiTtbMdbGV%2F1%2BGiTiB7Ovqc1gnXQvH6CbXQ8WPUSeGZDXEhLP505IiGPZz", nil, nil)
	get, err := util.HttpRawGET(rc, "https://www.qpm.io/packages/index.html", nil, nil)
	if err != nil {
		panic(err)
	}
	println(get)
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

func Test_html_query(t *testing.T) {
	for _, info := range ParseQPM("https://www.qpm.io/packages/index.html") {
		fmt.Printf("%v\n", info)
	}
}
