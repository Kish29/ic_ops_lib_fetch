package cron

import (
	"fmt"
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"log"
	"reflect"
)

const (
	webUrlVcpkg          = `https://vcpkg.io`
	allPackagesPathVcpkg = `/output.json`
)

type VcpkgPackage struct {
	Name               string        `json:"Name"`
	Version            string        `json:"Version"`
	PortVersion        int           `json:"Port-Version,omitempty"`
	Description        string        `json:"Description,omitempty"`
	Supports           string        `json:"Supports,omitempty"`
	Dependencies       []interface{} `json:"Dependencies,omitempty"`
	Features           interface{}   `json:"Features"`
	Arm64Windows       string        `json:"arm64-windows"`
	ArmUwp             string        `json:"arm-uwp"`
	X64Linux           string        `json:"x64-linux"`
	X64Osx             string        `json:"x64-osx"`
	X64Uwp             string        `json:"x64-uwp"`
	X64Windows         string        `json:"x64-windows"`
	X64WindowsStatic   string        `json:"x64-windows-static"`
	X64WindowsStaticMd string        `json:"x64-windows-static-md"`
	X86Windows         string        `json:"x86-windows"`
	Homepage           string        `json:"Homepage,omitempty"`
	License            *string       `json:"License,omitempty"`
	Stars              int           `json:"Stars,omitempty"`
	Maintainers        string        `json:"Maintainers,omitempty"`
	DefaultFeatures    []string      `json:"Default-Features,omitempty"`
	Documentation      string        `json:"Documentation,omitempty"`
}

type VcpkgAllPackageResp struct {
	GeneratedOn string          `json:"Generated On"`
	Size        int             `json:"Size"`
	Source      []*VcpkgPackage `json:"Source"`
}

type VcpkgFetcher struct {
	*core.BaseAsyncCronFetcher
}

func NewVcpkgFetcher() *VcpkgFetcher {
	return &VcpkgFetcher{BaseAsyncCronFetcher: &core.BaseAsyncCronFetcher{}}
}

func (v *VcpkgFetcher) Fetch() (info []*core.LibInfo, err error) {
	// 获取所有的包预览信息
	defaultHeaderAttr := map[string]string{
		util.HttpHeadKeyUserAgent: util.RandomFakeAgent(),
	}
	resp := &VcpkgAllPackageResp{}
	err = util.HttpGet2Json(
		conanClient,
		webUrlVcpkg+allPackagesPathVcpkg,
		nil,
		defaultHeaderAttr,
		resp,
	)
	if err != nil || resp.Source == nil {
		return nil, err
	}
	info = v.convertToLibInfo(resp.Source)
	log.Printf("Fetcher=>%s fetched %d packages!", v.Name(), len(info))
	return
}

func (v *VcpkgFetcher) Name() string {
	return "vcpkg"
}

func (v *VcpkgFetcher) convertToLibInfo(origin []*VcpkgPackage) []*core.LibInfo {
	if len(origin) <= 0 {
		return nil
	}
	libInfo := make([]*core.LibInfo, 0, len(origin))
	for _, vcpkgPackage := range origin {
		if vcpkgPackage == nil {
			continue
		}
		ver := &core.LibVer{
			Ver: vcpkgPackage.Version,
		}
		if vcpkgPackage.License != nil {
			ver.License = *vcpkgPackage.License
		}
		info := &core.LibInfo{
			Name:         vcpkgPackage.Name,
			VerDetail:    ver,
			Description:  vcpkgPackage.Description,
			Homepage:     &vcpkgPackage.Homepage,
			Dependencies: v.parseDependencies(vcpkgPackage.Dependencies),
			Stars:        &vcpkgPackage.Stars,
		}

		if vcpkgPackage.Maintainers != "" {
			info.Author = vcpkgPackage.Maintainers
		}
		libInfo = append(libInfo, info)
	}
	return libInfo
}

func (v *VcpkgFetcher) parseDependencies(dep []interface{}) []*core.LibDep {
	if dep == nil {
		return nil
	}
	deps := make([]*core.LibDep, 0, len(dep))
	for _, d := range dep {
		libDep := &core.LibDep{}
		value := reflect.ValueOf(d)
		switch value.Kind() {
		case reflect.Map:
			libDep.Name = fmt.Sprint(value.MapIndex(reflect.ValueOf("name")).Interface())
		case reflect.String:
			libDep.Name = value.String()
		}
		if len(libDep.Name) > 0 {
			deps = append(deps, libDep)
		}
	}
	return deps
}
