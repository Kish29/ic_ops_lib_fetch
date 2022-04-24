package test

import (
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/antchfx/htmlquery"
	"testing"
)

func Test_spack(t *testing.T) {
	doc := util.HttpGETNode("https://spack.readthedocs.io/en/latest/package_list.html")
	// package list
	packageList := htmlquery.FindOne(doc, `//div[@id='package-list']`)
	nodes := htmlquery.Find(packageList, `/div[@class='section']`)
	println(len(nodes))
	for _, node := range nodes {
		println(htmlquery.FindOne(node, `/h1/text()`).Data)
	}
}
