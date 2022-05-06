package download

import (
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/pool"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"os/exec"
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
				_ = exec.Command("wget", i.(string), `-P`, tarGZDir).Run()
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
