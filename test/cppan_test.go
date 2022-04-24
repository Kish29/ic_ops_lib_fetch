package test

import (
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/antchfx/htmlquery"
	"github.com/gookit/goutil/mathutil"
	"strings"
	"testing"
)

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
