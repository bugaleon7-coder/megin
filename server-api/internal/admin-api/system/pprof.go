package system

import (
	commonDto "megin/internal/dto"
	"megin/internal/profiler"
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// Pprof @Tag Pprof 性能分析
type Pprof struct{}

// Status @Summary 获取 pprof 状态
// @Description 返回开关状态、监听地址和当前进程的核心运行时指标。
func (h *Pprof) Status(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[profiler.Status], error) {
	return api.ResultData(systemService.PprofStatus())
}

// Toggle @Summary 启停 pprof
// @Description 动态启停仅监听 127.0.0.1 的 pprof 服务；端口冲突不会影响主服务。
func (h *Pprof) Toggle(ctx *api.Context, req *systemDto.PprofToggleReq) (*api.Result[profiler.Status], error) {
	status, err := systemService.TogglePprof(req.Enable)
	if err != nil {
		return nil, err
	}
	return api.ResultData(status)
}

// StartCPUProfile @Summary 启动综合 Pprof 采样
// @Description 一次任务同时收集 CPU、内存、协程、锁、阻塞、线程创建与 Trace，支持 30 秒、1 分钟、5 分钟或 10 分钟。
func (h *Pprof) StartCPUProfile(ctx *api.Context, req *systemDto.PprofCPUProfileReq) (*api.Result[systemDto.PprofRecord], error) {
	createdBy := uint(0)
	if ctx.AdminInfo != nil {
		createdBy = uint(ctx.AdminInfo.UserID)
	}
	job, err := systemService.StartPprofProfile(req.DurationSeconds, createdBy)
	if err != nil {
		return nil, err
	}
	return api.ResultData(job)
}

// CPUProfileResult @Summary 获取 CPU Profile 结果
// @Description 兼容接口，返回最新一条持久化记录及其火焰图。
func (h *Pprof) CPUProfileResult(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[systemDto.PprofRecordDetail], error) {
	detail, err := systemService.LatestPprofRecord()
	if err != nil {
		return nil, err
	}
	return api.ResultData(detail)
}

// Records @Summary 获取 Pprof 采样记录
func (h *Pprof) Records(ctx *api.Context, req *systemDto.PprofRecordPageReq) (*api.Result[commonDto.PageResult[systemDto.PprofRecord]], error) {
	result, err := systemService.PagePprofRecords(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}

// Record @Summary 查看 Pprof 采样记录
// @Description 已完成的记录同时返回持久化火焰图数据。
func (h *Pprof) Record(ctx *api.Context, req *systemDto.PprofRecordReq) (*api.Result[systemDto.PprofRecordDetail], error) {
	detail, err := systemService.GetPprofRecord(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(detail)
}
