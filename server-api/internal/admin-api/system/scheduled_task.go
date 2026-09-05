package system

import (
	commonDto "megin/internal/dto"
	"megin/internal/schedule"
	"megin/pkg/context/api"
)

// ScheduledTask @Tag 定时任务管理
type ScheduledTask struct{}

// Options @Summary 获取可配置的任务执行器
// @Description 返回服务端已注册的安全任务执行器；后台不可配置任意系统命令。
func (h *ScheduledTask) Options(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[[]schedule.JobOption], error) {
	return api.ResultData(schedule.TaskOptions())
}

// Create @Summary 新增定时任务
// @Description 配置任务名称、已注册执行器和标准五段 Cron 表达式，例如 0 2 * * *。
func (h *ScheduledTask) Create(ctx *api.Context, req *schedule.TaskReq) (*api.Result[schedule.Task], error) {
	task, err := schedule.CreateTask(req, uint(ctx.AdminInfo.UserID))
	if err != nil {
		return nil, err
	}
	return api.ResultData(task)
}

// Update @Summary 修改定时任务
// @Description 保存后立即重载当前服务实例的 Cron 调度器。
func (h *ScheduledTask) Update(ctx *api.Context, req *schedule.UpdateTaskReq) (*api.Result[schedule.Task], error) {
	task, err := schedule.UpdateTask(req, uint(ctx.AdminInfo.UserID))
	if err != nil {
		return nil, err
	}
	return api.ResultData(task)
}

// Delete @Summary 删除定时任务
func (h *ScheduledTask) Delete(ctx *api.Context, req *schedule.TaskIDReq) (*api.Result[Success], error) {
	if err := schedule.DeleteTask(req.ID); err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// Detail @Summary 查询定时任务详情
func (h *ScheduledTask) Detail(ctx *api.Context, req *schedule.TaskIDReq) (*api.Result[schedule.Task], error) {
	task, err := schedule.GetTask(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(task)
}

// PageList @Summary 分页查询定时任务
func (h *ScheduledTask) PageList(ctx *api.Context, req *schedule.TaskPageReq) (*api.Result[schedule.TaskPageResult], error) {
	result, err := schedule.PageTasks(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}

// Execute @Summary 手动执行定时任务
// @Description 立即同步执行任务，并返回本次执行日志。
func (h *ScheduledTask) Execute(ctx *api.Context, req *schedule.TaskIDReq) (*api.Result[schedule.TaskLog], error) {
	log, err := schedule.ExecuteTask(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(log)
}

// LogPageList @Summary 分页查询任务执行日志
func (h *ScheduledTask) LogPageList(ctx *api.Context, req *schedule.TaskLogPageReq) (*api.Result[schedule.TaskLogPageResult], error) {
	result, err := schedule.PageTaskLogs(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}
