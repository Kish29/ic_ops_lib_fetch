package download

import "os"

const (
	SourceCodeDir = "source_code"
)

type Downloader interface {
	DownloadTo(dir, name string) error
	DownloadToWait(dir, name string) error
}

func Exists(dir string) bool {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	} else {
		return true
	}
}

func IsDir(dir string) bool {
	s, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func IsFile(filename string) bool {
	return !IsDir(filename)
}
