package download

import (
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/gookit/goutil/fsutil"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	tarGZDir = "others(tar-gz)"
)

var (
	tarGzRegexp = regexp.MustCompile(`.*\.tar\.gz$`)
)

type TarGZWget struct {
	urlRec  map[string]bool
	tarUrls []string
}

func NewTarGZWget() *TarGZWget {
	return &TarGZWget{urlRec: make(map[string]bool)}
}

func (t *TarGZWget) Get() error {
	for _, url := range t.tarUrls {
		if strings.TrimSpace(url) == "" {
			continue
		}
		core.GlobalPool.Do(&pool.TaskHandler{
			Fn: func(i interface{}) error {
				targzUrl := i.(string)
				index := strings.LastIndex(targzUrl, `/`)
				if index != -1 {
					component := targzUrl[index+1:]
					if fsutil.FileExist(filepath.Join(tarGZDir, component)) {
						log.Printf("component=>%v exist! skip...", component)
						return nil
					}
				}
				err := exec.Command("wget", targzUrl, `-P`, tarGZDir).Run()
				if err != nil {
					log.Printf("[error] targz download=>%v error, err=>%v", targzUrl, err)
				} else {
					log.Printf("targz download=>%v success", targzUrl)
				}
				return nil
			},
			Param: url,
		})
	}
	return nil
}

func (t *TarGZWget) CanGetAndPut(url string) bool {
	if t.urlRec[url] {
		return true
	}
	if tarGzRegexp.Match(util.Str2Bytes(url)) {
		t.tarUrls = append(t.tarUrls, url)
		t.urlRec[url] = true
		return true
	}
	return false
}
