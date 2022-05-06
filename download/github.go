package download

// duplicated codes

import (
	"errors"
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"gorm.io/gorm"
	"io/fs"
	"log"
	"os"
	"os/exec"
)

type GithubDownloader struct {
	callback func(url string, succ bool)
}

func (g *GithubDownloader) DownloadAllInDB(dbConn *gorm.DB) {
	if dbConn == nil {
		return
	}
}

func (g *GithubDownloader) buildGitUrl(url string) string {
	if url[len(url)-1] == '/' {
		return url[:len(url)-1] + ".git"
	}
	return url + ".git"
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

func (g *GithubDownloader) DownloadToWait(dir, url string) error {
	if err := g.check(dir, url); err != nil {
		return err
	}
	url = g.buildGitUrl(url)
	// 执行git clone
	core.GlobalPool.DoWait(&pool.TaskHandler{
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
	core.GlobalPool.Do(&pool.TaskHandler{
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
	return &GithubDownloader{callback: callback}
}
