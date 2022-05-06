package test

import (
	"github.com/Kish29/ic_ops_lib_fetch/download"
	"github.com/gookit/goutil/fsutil"
	"os"
	"path/filepath"
	"testing"
)

func Test_file_exist(t *testing.T) {
	_ = os.Chdir("../" + download.SourceCodeDir)
	println(fsutil.FileExist(filepath.Join(`tengo`, "v2.0.0.zip")))
}
