package download

import (
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/go-resty/resty/v2"
	"github.com/gookit/goutil/fsutil"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	apiTagsFmt = `https://api.github.com/repos/%s/%s/tags`
	apiRepoZip = `https://github.com/%s/%s/archive/refs/tags/%s.zip`
	token      = `token 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`
)

var (
	gitRegexp = regexp.MustCompile(`^https://github.com.*`)
	gitClient = resty.New()
)

type GitTagInfo struct {
	Name       string `json:"name"`
	ZipballUrl string `json:"zipball_url"`
	TarballUrl string `json:"tarball_url"`
	Commit     *struct {
		Sha string `json:"sha"`
		Url string `json:"url"`
	} `json:"commit"`
	NodeId string `json:"node_id"`
}

type GithubWget struct {
	urlRec   map[string]bool
	repoUrls []string
}

func NewGithubWget() *GithubWget {
	return &GithubWget{urlRec: make(map[string]bool)}
}

func (g *GithubWget) Get() error {
	for _, url := range g.repoUrls {
		if strings.TrimSpace(url) == "" {
			continue
		}
		owner, repo := g.parseOwnerRepo(url)
		if strings.TrimSpace(owner) == "" || strings.TrimSpace(repo) == "" {
			return nil
		}
		tags := g.getAllTags(owner, repo)
		if len(tags) > 10 {
			tags = tags[:10]
		}
		for _, tag := range tags {
			if strings.TrimSpace(tag) == "" {
				continue
			}
			// 检查该文件是否存在
			if fsutil.FileExist(filepath.Join(repo, tag+`.tar.gz`)) {
				log.Printf("component=>%v exists! skip...", tag+`.tar.gz`)
				continue
			}
			if fsutil.FileExist(filepath.Join(repo, tag+`.zip`)) {
				log.Printf("component=>%v exists! skip...", tag+`.zip`)
				continue
			}
			// 执行下载
			type GitZipInfo struct {
				Owner string
				Repo  string
				Tag   string
			}
			core.GlobalPool.Do(&pool.TaskHandler{
				Fn: func(i interface{}) error {
					zipInfo, ok := i.(*GitZipInfo)
					if !ok || zipInfo == nil {
						return nil
					}
					err := exec.Command("wget", fmt.Sprintf(apiRepoZip, zipInfo.Owner, zipInfo.Repo, zipInfo.Tag), `-P`, repo).Run()
					if err != nil {
						log.Printf("[error] git wget=>%s::%s error, err=>%v", zipInfo.Repo, zipInfo.Tag, err)
					} else {
						log.Printf("git wget=>%s::%s success", zipInfo.Repo, zipInfo.Tag)
					}
					return nil
				},
				Param: &GitZipInfo{Owner: owner, Repo: repo, Tag: tag},
			})
		}
	}
	return nil
}

func (g *GithubWget) CanGetAndPut(url string) bool {
	if g.urlRec[url] {
		return true
	}
	if gitRegexp.Match(util.Str2Bytes(url)) {
		g.repoUrls = append(g.repoUrls, url)
		g.urlRec[url] = true
		return true
	}
	return false
}

func (g *GithubWget) parseOwnerRepo(url string) (owner string, repo string) {
	// https://github.com/ValveSoftware/openvr
	lastIdx := strings.LastIndex(url, `/`)
	repo = url[lastIdx+1:]
	url = url[:lastIdx]
	lastIdx = strings.LastIndex(url, `/`)
	owner = url[lastIdx+1:]
	return
}

func (g *GithubWget) defaultToken() map[string]string {
	return map[string]string{
		`Authorization`: token,
	}
}

func (g *GithubWget) getAllTags(owner, repo string) []string {
	tagUrl := fmt.Sprintf(apiTagsFmt, owner, repo)
	tagInfo := []*GitTagInfo{}
	err := util.HttpGETToJson(gitClient, tagUrl, nil, g.defaultToken(), &tagInfo)
	if err != nil {
		return nil
	}
	tags := make([]string, 0, len(tagInfo))
	for _, info := range tagInfo {
		if info == nil || strings.TrimSpace(info.Name) == "" {
			continue
		}
		tags = append(tags, info.Name)
	}
	return tags
}
