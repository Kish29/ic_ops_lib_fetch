package test

import (
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
)

func Test_dir_del(t *testing.T) {
	dir, err := ioutil.ReadDir(`../source_code`)
	if err != nil {
		panic(err)
	}
	reg := regexp.MustCompile(`^ .*`)
	noSpaceName := make([]string, len(dir))
	for _, d := range dir {
		name := d.Name()
		if !reg.Match(util.Str2Bytes(name)) {
			println(name)
			noSpaceName = append(noSpaceName, `../source_code/`+name)
		}
	}
	println(len(dir))
	println(len(noSpaceName))
	for i := range noSpaceName {
		_ = os.Remove(noSpaceName[i])
	}
}
