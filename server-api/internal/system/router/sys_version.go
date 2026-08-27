package router

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	handler "megin/internal/admin-api/system"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
	"megin/pkg/context/router"

	"github.com/gin-gonic/gin"
)

func SysVersionRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	version := &handler.SysVersion{}
	router.DELETE(adminApiGroup, "/sysVersion/deleteSysVersion", version.DeleteSysVersion)
	router.DELETE(adminApiGroup, "/sysVersion/deleteSysVersionByIds", version.DeleteSysVersionByIds)
	router.GET(adminApiGroup, "/sysVersion/findSysVersion", version.FindSysVersion)
	router.GET(adminApiGroup, "/sysVersion/getSysVersionList", version.GetSysVersionList)
	router.POST(adminApiGroup, "/sysVersion/exportVersion", version.ExportVersion)
	router.POST(adminApiGroup, "/sysVersion/importVersion", version.ImportVersion)
	router.GET(adminApiGroup, "/sysVersion/getSysVersionPublic", version.GetSysVersionPublic)

	adminApiGroup.AddRoute("GET", "/sysVersion/downloadVersionJson", func(c *gin.Context) {
		ctx, err := api.NewContext(c)
		if err != nil {
			c.JSON(http.StatusOK, map[string]any{"code": 5000, "message": err.Error(), "success": false})
			return
		}
		id := c.Query("ID")
		if id == "" {
			c.JSON(http.StatusOK, map[string]any{"code": 5000, "message": "版本ID不能为空", "success": false})
			return
		}
		version, err := systemService.NewSysVersion(ctx).GetSysVersion(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusOK, map[string]any{"code": 5000, "message": "获取版本记录失败:" + err.Error(), "success": false})
			return
		}
		var jsonData []byte
		if version.VersionData != nil && *version.VersionData != "" {
			jsonData = []byte(*version.VersionData)
		} else {
			jsonData = []byte("{}")
		}
		filename := fmt.Sprintf("version_%s_%s.json", deref(version.VersionCode), timeNowString())
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Length", strconv.Itoa(len(jsonData)))
		c.Data(http.StatusOK, "application/json", jsonData)
	}, nil, nil)

	return adminApiGroup
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func timeNowString() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}
