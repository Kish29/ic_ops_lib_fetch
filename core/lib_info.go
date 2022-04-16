package core

import "fmt"

type LibVer struct {
	Ver     string `json:"ver"`
	License string `json:"license"`
}

type LibDep struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type LibInfo struct {
	Name          string    `gorm:"column:name" json:"name"`                     // 开源库名称
	VerDetail     *LibVer   `json:"ver_detail"`                                  // 版本详情
	Description   string    `gorm:"column:description" json:"description"`       // 开源库描述
	Homepage      *string   `gorm:"column:homepage" json:"homepage"`             // 开源库网站链接
	DownloadCount int       `gorm:"column:download_count" json:"download_count"` // 下载次数
	SourceCode    *string   `gorm:"column:source_code" json:"source_code"`       // 开源库代码链接
	Dependencies  []*LibDep `gorm:"column:dependencies" json:"dependencies"`     // 依赖的其他开源库
	Author        string    `gorm:"column:author" json:"author"`                 // 开源库作者
	Contributors  []string  `gorm:"column:contributors" json:"contributors"`     // 开源库贡献者
	Stars         *int      `gorm:"column:stars" json:"stars"`                   // 开源库星标数量
	Watching      *int      `gorm:"column:watching" json:"watching"`             // 开源库订阅数量
	ForkCount     *int      `gorm:"column:fork_count" json:"fork_count"`         // 开源库fork数量
}

func (l *LibInfo) Unique() string {
	return fmt.Sprintf("%s::%s", l.Name, l.VerDetail.Ver)
}
