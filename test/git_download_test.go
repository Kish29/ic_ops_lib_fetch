package test

import (
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/download"
	"testing"
)

func Test_git_download(t *testing.T) {
	downloader := download.InitGitDownloader(func(url string, succ bool) {
		fmt.Printf("download=>%s success=>%v", url, succ)
	})
	err := downloader.DownloadTo("", "https://github.com/ValveSoftware/openvr")
	if err != nil {
		panic(err)
	}
}
