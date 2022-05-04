package test

import (
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/antchfx/htmlquery"
	"log"
	"strings"
	"testing"
)

const (
	SpackWebUrl = `https://spack.readthedocs.io/en/latest/package_list.html`
)

type SpackPackage struct {
	Name         string
	Homepage     string
	Versions     []string
	Dependencies []string
	Description  string
}

func SpackFetchAllPackages() []*SpackPackage {
	doc := util.HttpGETNode(SpackWebUrl)
	// package list
	packageList := htmlquery.FindOne(doc, `//div[@id='package-list']`)
	if packageList == nil {
		return nil
	}
	nodes := htmlquery.Find(packageList, `/div[@class='section']`)
	log.Printf("spack get %d packages", len(nodes))
	pkgs := make([]*SpackPackage, 0, len(nodes))
	for _, node := range nodes {
		p := &SpackPackage{
			Name: htmlquery.FindOne(node, `/h1/text()`).Data,
		}
		detailNode := htmlquery.FindOne(node, `//dl[@class='docutils']`)
		if detailNode != nil {
			// homepage
			homepage := htmlquery.FindOne(detailNode, `/dd[1]//a`)
			if homepage != nil {
				p.Homepage = htmlquery.InnerText(homepage)
			}
			// versions
			versions := htmlquery.FindOne(detailNode, `/dd[3]/text()`)
			if versions != nil {
				p.Versions = strings.Split(versions.Data, `,`)
			}
			// dependencies
			dependencies := htmlquery.Find(detailNode, `/dd[5]//a`)
			for _, dependency := range dependencies {
				p.Dependencies = append(p.Dependencies, htmlquery.InnerText(dependency))
			}
			// description
			descNode := htmlquery.FindOne(detailNode, `/dd[6]/text()`)
			if descNode != nil {
				p.Description = descNode.Data
			}
		}
		pkgs = append(pkgs, p)
	}
	return pkgs
}

func Test_spack(t *testing.T) {
	for _, spackPackage := range SpackFetchAllPackages() {
		fmt.Printf("%v\n", spackPackage)
	}
}
