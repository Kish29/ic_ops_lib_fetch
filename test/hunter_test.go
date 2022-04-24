package test

import (
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/antchfx/htmlquery"
	"testing"
)

func Test_hunter(t *testing.T) {
	doc := util.HttpGETNode("https://hunter.readthedocs.io/en/latest/packages/all.html")
	nodes := htmlquery.Find(doc, `//li[@class='toctree-l1']`)
	println(len(nodes))
}
