package download

import (
	"errors"
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/db"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	apiTagsFmt = `https://api.github.com/repos/%s/%s/tags`
	apiRepoZip = `https://github.com/%s/%s/archive/refs/tags/%s.zip`
	token      = `token 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`
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

var (
	gitRegexp = regexp.MustCompile(`^https://github.com.*`)
	gitClient = resty.New()
)

type GithubDownloader struct {
	dPool    *pool.WorkPool
	callback func(url string, succ bool)
}

func (g *GithubDownloader) DownloadAllInDB(dbConn *gorm.DB) {
	if dbConn == nil {
		return
	}
	var downloads []*db.TBsLibInfo
	res := dbConn.Or("homepage <> ?", "").Or("source_code <> ?", "").Find(&downloads)
	if res.Error != nil {
		panic(res.Error)
	}
	// 通过lib名的所有版本或者url，保留一个符合github的url
	dMap := make(map[string]string, len(downloads))
	for _, download := range downloads {
		_, ok := dMap[download.Name]
		if ok || (download.SourceCode == "" && download.Homepage == "") {
			continue
		}
		if !gitRegexp.Match(util.Str2Bytes(download.SourceCode)) || !gitRegexp.Match(util.Str2Bytes(download.Homepage)) {
			continue
		}
		if download.Homepage != "" {
			dMap[download.Name] = download.Homepage
		}
		if download.SourceCode != "" {
			dMap[download.Name] = download.SourceCode
		}
	}
	// 下载所有的lib
	for _, url := range dMap {
		_ = g.DownloadAllVersions(url)
	}
}

func (g *GithubDownloader) buildGitUrl(url string) string {
	if url[len(url)-1] == '/' {
		return url[:len(url)-1] + ".git"
	}
	return url + ".git"
}

func (g *GithubDownloader) parseOwnerRepo(url string) (owner string, repo string) {
	// https://github.com/ValveSoftware/openvr
	lastIdx := strings.LastIndex(url, `/`)
	repo = url[lastIdx+1:]
	url = url[:lastIdx]
	lastIdx = strings.LastIndex(url, `/`)
	owner = url[lastIdx+1:]
	return
}

func (g *GithubDownloader) check(dir, url string) error {
	if len(url) <= 0 {
		return errors.New("url is empty")
	}
	// 检查目录状态
	if !Exists(dir) {
		log.Printf("[info] dir=>%s is not exists", dir)
		// 创建要保存的目录
		err := os.Mkdir(dir, fs.ModeDir|fs.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GithubDownloader) DownloadAllVersions(url string) error {
	if len(url) <= 0 {
		return errors.New("url is empty")
	}
	owner, repo := g.parseOwnerRepo(url)
	if owner == "" || repo == "" {
		return errors.New("owner or repo is empty")
	}
	tagUrl := fmt.Sprintf(apiTagsFmt, owner, repo)
	tagInfo := []*GitTagInfo{}
	defaultHeaderAttr := map[string]string{
		`Authorization`: token,
	}
	err := util.HttpGETToJson(gitClient, tagUrl, nil, defaultHeaderAttr, &tagInfo)
	if err != nil {
		return err
	}
	_ = os.Mkdir(repo, os.ModeDir|os.ModePerm)
	targetDirArg := fmt.Sprintf("-P %s", repo)
	for _, info := range tagInfo {
		// 执行wget
		// TODO: 检查是否已经下载
		zipUrl := fmt.Sprintf(apiRepoZip, owner, repo, info.Name)
		g.dPool.Do(&pool.TaskHandler{
			Fn: func(u interface{}) error {
				dUrl := u.(string)
				log.Printf("Start downloadAllVersions souce code for=>%v", dUrl)
				err := exec.Command("wget", targetDirArg, dUrl).Run()
				if g.callback != nil {
					if err != nil {
						g.callback(dUrl, false)
					} else {
						g.callback(dUrl, true)
					}
				}
				return nil
			},
			Param: zipUrl,
		})
	}
	return nil
}

func (g *GithubDownloader) DownloadToWait(dir, url string) error {
	if err := g.check(dir, url); err != nil {
		return err
	}
	url = g.buildGitUrl(url)
	// 执行git clone
	g.dPool.DoWait(&pool.TaskHandler{
		Fn: func(u interface{}) error {
			dUrl := u.(string)
			// enter dir
			err := os.Chdir(dir)
			if err != nil {
				if g.callback != nil {
					g.callback(dUrl, false)
				}
				return nil
			}
			log.Printf("Start downloadAllVersions souce code for=>%v", dUrl)
			err = exec.Command("git", "clone", dUrl).Run()
			if g.callback != nil {
				if err != nil {
					g.callback(dUrl, false)
				} else {
					g.callback(dUrl, true)
				}
			}
			_ = os.Chdir("../")
			return nil
		},
		Param: url,
	})
	return nil
}

func (g *GithubDownloader) DownloadTo(dir, url string) error {
	if err := g.check(dir, url); err != nil {
		return err
	}
	url = g.buildGitUrl(url)
	// 执行git clone
	g.dPool.Do(&pool.TaskHandler{
		Fn: func(u interface{}) error {
			dUrl := u.(string)
			// enter dir
			err := os.Chdir(dir)
			if err != nil {
				if g.callback != nil {
					g.callback(dUrl, false)
				}
				return nil
			}
			log.Printf("Start downloadAllVersions souce code for=>%v", dUrl)
			err = exec.Command("git", "clone", dUrl).Run()
			if g.callback != nil {
				if err != nil {
					g.callback(dUrl, false)
				} else {
					g.callback(dUrl, true)
				}
			}
			_ = os.Chdir("../")
			return nil
		},
		Param: url,
	})
	return nil
}

func InitGitDownloader(callback func(url string, succ bool)) *GithubDownloader {
	// 1. check git是否存在
	err := exec.Command("git", "version").Run()
	if err != nil {
		panic(fmt.Errorf("[fatal] git not installed, error=>%v", err))
	}
	return &GithubDownloader{dPool: pool.New(2048), callback: callback}
}
