package schedule

import (
	"errors"
	"fmt"
	commonDto "megin/internal/dto"
	"megin/pkg/errs"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type TaskReq struct {
	Name     string `json:"name" binding:"required,min=1,max=100"`      // 任务名称
	JobKey   string `json:"job_key" binding:"required,min=1,max=100"`   // 已注册执行器标识
	CronExpr string `json:"cron_expr" binding:"required,min=1,max=100"` // 标准五段 Cron 表达式
	Status   int    `json:"status" binding:"oneof=0 1"`                 // 状态：0禁用 1启用
	Remark   string `json:"remark" binding:"omitempty,max=500"`         // 备注
}

type UpdateTaskReq struct {
	ID uint `json:"id" binding:"required,min=1"`
	TaskReq
}
type TaskIDReq struct {
	ID uint `json:"id" form:"id" binding:"required,min=1"`
}
type TaskPageReq struct {
	commonDto.PageQuery
	Name   string `form:"name" json:"name" binding:"omitempty,max=100"`
	Status *int   `form:"status" json:"status" binding:"omitempty,oneof=0 1"`
}
type TaskLogPageReq struct {
	commonDto.PageQuery
	TaskID uint `form:"task_id" json:"task_id" binding:"required,min=1"`
}
type TaskPageResult struct {
	PageNo    int    `json:"page_no"`
	PageSize  int    `json:"page_size"`
	TotalSize int64  `json:"total_size"`
	TotalPage int64  `json:"total_page"`
	List      []Task `json:"list"`
}
type TaskLogPageResult struct {
	PageNo    int       `json:"page_no"`
	PageSize  int       `json:"page_size"`
	TotalSize int64     `json:"total_size"`
	TotalPage int64     `json:"total_page"`
	List      []TaskLog `json:"list"`
}

func TaskOptions() []JobOption { return jobOptions }

func CreateTask(req *TaskReq, operatorID uint) (Task, error) {
	if err := validateTaskReq(req); err != nil {
		return Task{}, err
	}
	var count int64
	if err := manager.db.Model(&Task{}).Where("job_key = ?", req.JobKey).Count(&count).Error; err != nil {
		return Task{}, err
	}
	if count > 0 {
		return Task{}, errs.NewBusinessError(4000, "同一任务执行器只能配置一条任务")
	}
	now := time.Now()
	task := Task{Name: strings.TrimSpace(req.Name), JobKey: req.JobKey, CronExpr: strings.TrimSpace(req.CronExpr), Status: req.Status, Remark: strings.TrimSpace(req.Remark), CreatedBy: operatorID, UpdatedBy: operatorID, CreatedAt: now, UpdatedAt: now}
	if err := manager.db.Create(&task).Error; err != nil {
		return Task{}, err
	}
	return task, manager.reload()
}

func UpdateTask(req *UpdateTaskReq, operatorID uint) (Task, error) {
	if err := validateTaskReq(&req.TaskReq); err != nil {
		return Task{}, err
	}
	var task Task
	if err := manager.db.First(&task, req.ID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return Task{}, errs.NewBusinessError(404, "定时任务不存在")
	} else if err != nil {
		return Task{}, err
	}
	var count int64
	if err := manager.db.Model(&Task{}).Where("job_key = ? AND id <> ?", req.JobKey, req.ID).Count(&count).Error; err != nil {
		return Task{}, err
	}
	if count > 0 {
		return Task{}, errs.NewBusinessError(4000, "同一任务执行器只能配置一条任务")
	}
	task.Name, task.JobKey, task.CronExpr, task.Status, task.Remark = strings.TrimSpace(req.Name), req.JobKey, strings.TrimSpace(req.CronExpr), req.Status, strings.TrimSpace(req.Remark)
	task.UpdatedBy, task.UpdatedAt = operatorID, time.Now()
	if err := manager.db.Save(&task).Error; err != nil {
		return Task{}, err
	}
	return task, manager.reload()
}

func DeleteTask(id uint) error {
	result := manager.db.Delete(&Task{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errs.NewBusinessError(404, "定时任务不存在")
	}
	return manager.reload()
}

func GetTask(id uint) (Task, error) {
	var task Task
	err := manager.db.First(&task, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Task{}, errs.NewBusinessError(404, "定时任务不存在")
	}
	return task, err
}

func ExecuteTask(id uint) (TaskLog, error) {
	task, err := GetTask(id)
	if err != nil {
		return TaskLog{}, err
	}
	if _, ok := executors[task.JobKey]; !ok {
		return TaskLog{}, errs.NewBusinessError(4000, "任务执行器未注册")
	}
	return manager.execute(task, "manual"), nil
}

func PageTasks(req *TaskPageReq) (TaskPageResult, error) {
	var tasks []Task
	query := manager.db.Model(&Task{})
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	total, err := pageTasks(query.Order("id DESC"), req.PageQuery, &tasks)
	return TaskPageResult{PageNo: normalizedPage(req.PageNo), PageSize: normalizedSize(req.PageSize), TotalSize: total, TotalPage: pageCount(total, normalizedSize(req.PageSize)), List: tasks}, err
}
func PageTaskLogs(req *TaskLogPageReq) (TaskLogPageResult, error) {
	var logs []TaskLog
	total, err := pageTasks(manager.db.Model(&TaskLog{}).Where("task_id = ?", req.TaskID).Order("id DESC"), req.PageQuery, &logs)
	return TaskLogPageResult{PageNo: normalizedPage(req.PageNo), PageSize: normalizedSize(req.PageSize), TotalSize: total, TotalPage: pageCount(total, normalizedSize(req.PageSize)), List: logs}, err
}

func pageTasks(query *gorm.DB, page commonDto.PageQuery, result any) (int64, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return 0, err
	}
	err := query.Offset((normalizedPage(page.PageNo) - 1) * normalizedSize(page.PageSize)).Limit(normalizedSize(page.PageSize)).Find(result).Error
	return total, err
}
func normalizedPage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}
func normalizedSize(size int) int {
	if size < 1 {
		return 20
	}
	return size
}
func pageCount(total int64, size int) int64 { return (total + int64(size) - 1) / int64(size) }

func validateTaskReq(req *TaskReq) error {
	if _, ok := executors[req.JobKey]; !ok {
		return errs.NewBusinessError(4000, "请选择已注册的任务执行器")
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if _, err := parser.Parse(strings.TrimSpace(req.CronExpr)); err != nil {
		return errs.NewBusinessError(4000, fmt.Sprintf("Cron 表达式无效：%v", err))
	}
	return nil
}
