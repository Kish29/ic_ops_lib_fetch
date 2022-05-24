package net

import "github.com/gin-gonic/gin"

var svc *gin.Engine

const targetPathValueKey = `name`

func StartupServiceHTTPService() {
	svc.GET("/info/:name", GetComponentInfo)
	svc.GET("/deps/:name", GetComponentDeps)
	svc.GET("/deps_ui/:name", GetUIComponentDeps)
	svc.GET("/security/:name", GetComponentSecurityInfo)
	err := svc.Run(":9898")
	if err != nil {
		panic(err)
	}
}

func init() {
	if svc == nil {
		svc = gin.Default()
	}
}
