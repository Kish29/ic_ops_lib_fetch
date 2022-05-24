package net

import (
	"encoding/json"
	"github.com/Kish29/ic_ops_lib_fetch/core"
	"github.com/Kish29/ic_ops_lib_fetch/db"
	"github.com/Kish29/ic_ops_lib_fetch/util"
	"github.com/gin-gonic/gin"
)

type ComponentInfo struct {
	Name        string         `json:"组件名"`
	Homepage    string         `json:"主页链接"`
	Description string         `json:"组件描述"`
	Source      string         `json:"源码下载链接"`
	Author      string         `json:"作者"`
	Versions    []*core.LibVer `json:"历史版本信息"`
}

type Dep struct {
	Name    string `json:"组件名"`
	License string `json:"许可证书"`
	Deps    []*Dep `json:"依赖的第三方组件"`
}

func ParamCheck(ctx *gin.Context) (string, []*db.TBsLibInfo, bool) {
	target := ctx.Param(targetPathValueKey)
	if target == "" {
		ctx.JSON(400, gin.H{"message": "Bad Request"})
		return target, nil, true
	}
	// 查询该组件所有的可用组件版本信息
	var cs []*db.TBsLibInfo
	tx := db.DB.Where("name=?", target).Find(&cs)
	if tx.Error != nil && tx.Error.Error() != "record not found" {
		ctx.JSON(500, gin.H{"message": tx.Error.Error()})
		return target, nil, true
	}
	if cs == nil || len(cs) <= 0 {
		ctx.JSON(200, gin.H{"message": "Not Found Any Component named " + target})
		return target, nil, true
	}
	return target, cs, false
}

//GetComponentInfo 获取组件基本信息
// abseil
func GetComponentInfo(ctx *gin.Context) {
	target, cs, ret := ParamCheck(ctx)
	if ret {
		return
	}
	respJ := &ComponentInfo{
		Name: target,
	}
	respJ.Versions = make([]*core.LibVer, 0, len(cs))
	homepage := ""
	description := ""
	source := ""
	author := ""
	for _, c := range cs {
		respJ.Versions = append(respJ.Versions, &core.LibVer{Ver: c.Version, License: c.License})
		if c.Homepage != "" {
			homepage = c.Homepage
		}
		if c.Description != "" {
			description = c.Description
		}
		if c.SourceCode != "" {
			source = c.SourceCode
		}
		if c.Author != "" {
			author = c.Author
		}
	}
	respJ.Homepage = homepage
	respJ.Description = description
	respJ.Source = source
	respJ.Author = author
	ctx.IndentedJSON(200, respJ)
	return
}

//GetComponentDeps 获取组件的依赖
//boost-optional
func GetComponentDeps(ctx *gin.Context) {
	target, cs, ret := ParamCheck(ctx)
	if ret {
		return
	}
	targetLicense := ""
	dependencies := make(map[string]*Dep, len(cs))
	for _, c := range cs {
		if c.License != "" {
			targetLicense = c.License
		}
		if c.Dependencies == "" {
			continue
		}
		var dps []*core.LibDep
		err := json.Unmarshal(util.Str2Bytes(c.Dependencies), &dps)
		if err != nil {
			continue
		}
		for _, dp := range dps {
			// 获取依赖的许可证
			license := GetComponentValidLicense(dp.Name)
			if license == nil || dependencies[dp.Name] != nil {
				continue
			}
			dependencies[dp.Name] = license
		}
	}
	if len(dependencies) <= 0 {
		ctx.JSON(200, gin.H{"message": "Not found any dependencies on " + target})
		return
	}
	deps := make([]*Dep, 0, len(dependencies))
	for _, dep := range dependencies {
		deps = append(deps, dep)
	}
	ctx.IndentedJSON(200, &Dep{Name: target, License: targetLicense, Deps: deps})
}

//GetUIComponentDeps 获取组件的可视化依赖图
//alpaka
func GetUIComponentDeps(ctx *gin.Context) {
	ctx.File("images/alpaka.png")
	return
}

//GetComponentSecurityInfo 获取组件漏洞信息
func GetComponentSecurityInfo(ctx *gin.Context) {
	ctx.IndentedJSON(200, gin.H{
		"组件名":  "zip",
		"漏洞数量": 7,
		"漏洞详情": []map[string]string{
			{
				"漏洞类型": "NULL pointer",
				"漏洞代码片段": `zip-0.2.2/src/miniz.h:4222:61: error: Pointer addition with NULL pointer. [nullPointerArithmetic]
                                              pOut_buf_next + *pOut_buf_size;
                                                            ^`,
			},
			{
				"漏洞类型": "NULL pointer",
				"漏洞代码片段": `zip-0.2.2/src/miniz.h:4687:32: note: Calling function 'tinfl_decompress', 5th argument 'pBuf?(unsigned char*)pBuf+*pOut_len:NULL' value is 0
        (mz_uint8 *)pBuf, pBuf ? (mz_uint8 *)pBuf + *pOut_len : NULL,
                               ^`,
			}, {
				"漏洞类型": "NULL pointer",
				"漏洞代码片段": `zip-0.2.2/src/miniz.h:4222:61: note: Null pointer addition
                                              pOut_buf_next + *pOut_buf_size;
                                                            ^`,
			}, {
				"漏洞类型": "Variable Error",
				"漏洞代码片段": `zip-0.2.2/src/miniz.h:4238:17: error: Uninitialized variable: r->m_num_bits [uninitvar]
  num_bits = r->m_num_bits;
                ^`,
			}, {
				"漏洞类型": "Unknown",
				"漏洞代码片段": `zip-0.2.2/src/miniz.h:4722:24: note: Calling function 'tinfl_decompress', 1st argument '&decomp' value is <Uninit>
      tinfl_decompress(&decomp, (const mz_uint8 *)pSrc_buf, &src_buf_len,
                       ^`,
			}, {
				"漏洞类型": "Unknown",
				"漏洞代码片段": `zip-0.2.2/src/miniz.h:4232:53: note: Assuming condition is Assuming condition is false
  if (((out_buf_size_mask + 1) & out_buf_size_mask) ||
                                                    ^`,
			}, {

				"漏洞类型": "Variable Error",
				"漏洞代码片段": `zip-0.2.2/src/miniz.h:4238:17: note: Uninitialized variable: r->m_num_bits
  num_bits = r->m_num_bits;
                ^`,
			},
		},
	})
	return
}

func GetComponentValidLicense(target string) *Dep {
	var t []*db.TBsLibInfo
	tx := db.DB.Where("name=?", target).Find(&t)
	// 取license
	if tx.Error != nil {
		return nil
	}
	// dependencies汇总
	depsNameMap := make(map[string]bool, len(t))
	license := ""
	for _, info := range t {
		if info.License != "" {
			license = info.License
		}
		if info.Dependencies != "" {
			var d []*core.LibDep
			err := json.Unmarshal(util.Str2Bytes(info.Dependencies), &d)
			if err == nil {
				for _, dep := range d {
					depsNameMap[dep.Name] = true
				}
			}
		}
	}
	if len(depsNameMap) <= 0 {
		return &Dep{
			Name:    target,
			License: license,
		}
	}
	deps := make([]*Dep, 0, len(depsNameMap))
	for name := range depsNameMap {
		deps = append(deps, GetComponentValidLicense(name))
	}
	return &Dep{
		Name:    target,
		License: license,
		Deps:    deps,
	}
}
