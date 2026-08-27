package system

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	commonDto "megin/internal/module/common/dto"
	systemDto "megin/internal/system/dto"
	systemModel "megin/internal/system/model"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysVersion @Tag 版本管理
type SysVersion struct{}

func (this *SysVersion) DeleteSysVersion(ctx *api.Context, req *systemDto.GetSysVersionReq) (*api.Result[any], error) {
	err := systemService.NewSysVersion(ctx).DeleteSysVersion(ctx.GinCtx.Request.Context(), strconv.FormatUint(uint64(req.ID), 10))
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

func (this *SysVersion) DeleteSysVersionByIds(ctx *api.Context, req *systemDto.DeleteSysVersionByIdsReq) (*api.Result[any], error) {
	ids := make([]string, 0, len(req.IDs))
	for _, id := range req.IDs {
		ids = append(ids, strconv.FormatUint(uint64(id), 10))
	}
	err := systemService.NewSysVersion(ctx).DeleteSysVersionByIds(ctx.GinCtx.Request.Context(), ids)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

func (this *SysVersion) FindSysVersion(ctx *api.Context, req *systemDto.GetSysVersionReq) (*api.Result[systemDto.SysVersion], error) {
	result, err := systemService.NewSysVersion(ctx).GetSysVersion(ctx.GinCtx.Request.Context(), strconv.FormatUint(uint64(req.ID), 10))
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

func (this *SysVersion) GetSysVersionList(ctx *api.Context, req *systemDto.SysVersionSearch) (*api.Result[systemDto.PageResult[systemDto.SysVersion]], error) {
	result, err := systemService.NewSysVersion(ctx).GetSysVersionInfoList(ctx.GinCtx.Request.Context(), req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

func (this *SysVersion) GetSysVersionPublic(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[any], error) {
	systemService.NewSysVersion(ctx).GetSysVersionPublic(ctx.GinCtx.Request.Context())
	return api.ResultData[any](map[string]any{"info": "不需要鉴权的版本管理接口信息"})
}

func (this *SysVersion) ExportVersion(ctx *api.Context, req *systemDto.ExportVersionRequest) (*api.Result[any], error) {
	svc := systemService.NewSysVersion(ctx)

	var menus []systemModel.SysBaseMenu
	var apis []systemModel.SysApi
	var dicts []systemModel.SysDictionary
	var err error

	if len(req.MenuIds) > 0 {
		menus, err = svc.GetMenusByIds(ctx.GinCtx.Request.Context(), req.MenuIds)
		if err != nil {
			return nil, err
		}
	}
	if len(req.ApiIds) > 0 {
		apis, err = svc.GetApisByIds(ctx.GinCtx.Request.Context(), req.ApiIds)
		if err != nil {
			return nil, err
		}
	}
	if len(req.DictIds) > 0 {
		dicts, err = svc.GetDictionariesByIds(ctx.GinCtx.Request.Context(), req.DictIds)
		if err != nil {
			return nil, err
		}
	}

	jsonData, err := svc.ExportVersionPayload(req.VersionName, req.VersionCode, req.Description, menus, apis, dicts)
	if err != nil {
		return nil, err
	}

	version := systemModel.SysVersion{
		VersionName: strPtr(req.VersionName),
		VersionCode: strPtr(req.VersionCode),
		Description: strPtr(req.Description),
		VersionData: strPtr(string(jsonData)),
	}
	if err := svc.CreateSysVersion(ctx.GinCtx.Request.Context(), &version); err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

func (this *SysVersion) DownloadVersionJson(ctx *api.Context, req *systemDto.GetSysVersionReq) (*api.Result[any], error) {
	svc := systemService.NewSysVersion(ctx)
	version, err := svc.GetSysVersion(ctx.GinCtx.Request.Context(), strconv.FormatUint(uint64(req.ID), 10))
	if err != nil {
		return nil, err
	}

	var jsonData []byte
	if version.VersionData != nil && *version.VersionData != "" {
		jsonData = []byte(*version.VersionData)
	} else {
		basic := systemDto.ExportVersionResponse{
			Version: systemDto.VersionInfo{
				Name:        derefString(version.VersionName),
				Code:        derefString(version.VersionCode),
				Description: derefString(version.Description),
				ExportTime:  time.Now().Format("2006-01-02 15:04:05"),
			},
			Menus: []systemModel.SysBaseMenu{},
			Apis:  []systemModel.SysApi{},
		}
		jsonData, err = json.MarshalIndent(basic, "", "  ")
		if err != nil {
			return nil, err
		}
	}

	filename := fmt.Sprintf("version_%s_%s.json", derefString(version.VersionCode), time.Now().Format("20060102150405"))
	ctx.GinCtx.Writer.Header().Set("Content-Type", "application/json")
	ctx.GinCtx.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	ctx.GinCtx.Writer.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
	ctx.GinCtx.Writer.WriteHeader(http.StatusOK)
	_, _ = ctx.GinCtx.Writer.Write(jsonData)
	return nil, nil
}

func (this *SysVersion) ImportVersion(ctx *api.Context, req *systemDto.ImportVersionRequest) (*api.Result[any], error) {
	svc := systemService.NewSysVersion(ctx)

	if len(req.ExportMenu) > 0 {
		if err := svc.ImportMenus(ctx.GinCtx.Request.Context(), req.ExportMenu); err != nil {
			return nil, err
		}
	}
	if len(req.ExportApi) > 0 {
		if err := svc.ImportApis(ctx.GinCtx.Request.Context(), req.ExportApi); err != nil {
			return nil, err
		}
	}
	if len(req.ExportDictionary) > 0 {
		if err := svc.ImportDictionaries(ctx.GinCtx.Request.Context(), req.ExportDictionary); err != nil {
			return nil, err
		}
	}

	jsonData, _ := json.Marshal(req)
	version := systemModel.SysVersion{
		VersionName: strPtr(req.VersionInfo.Name),
		VersionCode: strPtr(fmt.Sprintf("%s_imported_%s", req.VersionInfo.Code, time.Now().Format("20060102150405"))),
		Description: strPtr(fmt.Sprintf("导入版本: %s", req.VersionInfo.Description)),
		VersionData: strPtr(string(jsonData)),
	}
	_ = svc.CreateSysVersion(ctx.GinCtx.Request.Context(), &version)
	return api.ResultSuccess()
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func strPtr(v string) *string {
	return &v
}
