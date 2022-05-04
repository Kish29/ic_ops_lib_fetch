package cron

import (
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/download"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/antchfx/htmlquery"
	"log"
	"regexp"
	"strings"
	"sync"
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

type SpackFetcher struct {
	*core.BaseAsyncCronFetcher
	workers *pool.WorkPool
}

var gitRegexp = regexp.MustCompile(`^https://github.com.*`)

func (s *SpackFetcher) Fetch() ([]*core.LibInfo, error) {
	packages := s.FetchAllPackages()
	wg := sync.WaitGroup{}
	lock := sync.RWMutex{}
	infos := make([]*core.LibInfo, 0, len(packages))

	for _, spackPackage := range packages {
		if gitRegexp.MatchString(spackPackage.Homepage) { // 如果是github项目
			wg.Add(1)
			s.workers.Do(&pool.TaskHandler{
				Fn: func(i interface{}) error {
					defer wg.Done()

					url := i.(string)
					verInfos := make([]*core.LibInfo, 0, len(spackPackage.Versions))
					detail := download.GetRepoDetailByUrl(url)
					if detail != nil {
						for _, tag := range detail.Tags {
							verInfos = append(verInfos, &core.LibInfo{
								Name: spackPackage.Name,
								VerDetail: &core.LibVer{
									Ver:     tag.Ver,
									License: detail.License,
								},
								Homepage:     &url,
								Author:       detail.Owner,
								Stars:        &detail.Star,
								Watching:     &detail.Watch,
								ForkCount:    &detail.Fork,
								Contributors: detail.Contributors,
								SourceCode:   &tag.Zip,
							})
						}
					}
					if len(verInfos) > 0 {
						lock.Lock()
						infos = append(infos, verInfos...)
						lock.Unlock()
					}
					return nil
				},
				Param: spackPackage.Homepage,
			})
		} else {
			verInfos := make([]*core.LibInfo, 0, len(spackPackage.Versions))
			for _, version := range spackPackage.Versions {
				info := &core.LibInfo{
					Name: spackPackage.Name,
					VerDetail: &core.LibVer{
						Ver:     version,
						License: "",
					},
					Description: spackPackage.Description,
					Homepage:    &spackPackage.Homepage,
				}
				for _, dep := range spackPackage.Dependencies {
					info.Dependencies = append(info.Dependencies, &core.LibDep{
						Name:    dep,
						Version: "",
					})
				}
				verInfos = append(verInfos, info)
			}
			if len(verInfos) > 0 {
				lock.Lock()
				infos = append(infos, verInfos...)
				lock.Unlock()
			}
		}
	}
	wg.Wait()
	return infos, nil
}

func (s *SpackFetcher) Name() string {
	return "spack"
}

func NewSpackFetcher() *SpackFetcher {
	return &SpackFetcher{BaseAsyncCronFetcher: &core.BaseAsyncCronFetcher{}, workers: pool.New(2048)}
}

func (s *SpackFetcher) FetchAllPackages() []*SpackPackage {
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
