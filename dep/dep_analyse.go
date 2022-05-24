package dep

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

// 依赖分析
// 读取target_list中的文件名称

const (
	TargetListFilename = `target_list.json`
)

func ReaderTargetList() []string {
	f, err := os.Open(TargetListFilename)
	if err != nil {
		panic(err)
	}
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	var filenames []string
	err = json.Unmarshal(bytes, &filenames)
	if err != nil {
		panic(err)
	}
	return filenames
}

func StartDepAnalyse(sourceDir string) {
	// 读取文件中对依赖文件的配置
}
