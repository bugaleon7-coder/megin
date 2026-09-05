package schedule

import (
	"context"
	"errors"
	"fmt"
	taskCache "megin/internal/cache"
	"megin/internal/config"
	rateLimitRuntime "megin/internal/system/runtime"
	"megin/pkg/logger"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	taskStatusDisabled = 0
	taskStatusEnabled  = 1

	logStatusSuccess = 1
	logStatusFailed  = 2
	logStatusSkipped = 3

	taskLockTTL = time.Minute
)

// Task 是可由后台配置的定时任务。JobKey 必须是服务端已注册的任务标识。
type Task struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement;comment:任务ID" json:"id"`
	Name      string    `gorm:"column:name;size:100;not null;comment:任务名称" json:"name"`
	JobKey    string    `gorm:"column:job_key;size:100;not null;uniqueIndex;comment:执行器标识" json:"job_key"`
	CronExpr  string    `gorm:"column:cron_expr;size:100;not null;comment:Cron表达式" json:"cron_expr"`
	Status    int       `gorm:"column:status;not null;default:1;index;comment:状态 0禁用 1启用" json:"status"`
	Remark    string    `gorm:"column:remark;size:500;not null;default:'';comment:备注" json:"remark"`
	CreatedBy uint      `gorm:"column:created_by;not null;default:0;comment:创建管理员ID" json:"created_by"`
	UpdatedBy uint      `gorm:"column:updated_by;not null;default:0;comment:最后修改管理员ID" json:"updated_by"`
	CreatedAt time.Time `gorm:"column:created_at;not null;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;comment:修改时间" json:"updated_at"`
}

func (Task) TableName() string { return "scheduled_tasks" }

// TaskLog 保存每次自动或手动执行的完整输出。
type TaskLog struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement;comment:执行日志ID" json:"id"`
	TaskID    uint      `gorm:"column:task_id;not null;index;comment:任务ID" json:"task_id"`
	TaskName  string    `gorm:"column:task_name;size:100;not null;comment:任务名称快照" json:"task_name"`
	Trigger   string    `gorm:"column:trigger_type;size:20;not null;comment:触发方式 cron或manual" json:"trigger_type"`
	Status    int       `gorm:"column:status;not null;comment:执行状态 1成功 2失败" json:"status"`
	Output    string    `gorm:"column:output;type:text;not null;comment:执行输出" json:"output"`
	StartedAt time.Time `gorm:"column:started_at;not null;index;comment:开始时间" json:"started_at"`
	EndedAt   time.Time `gorm:"column:ended_at;not null;comment:结束时间" json:"ended_at"`
	Duration  int64     `gorm:"column:duration_ms;not null;comment:耗时毫秒" json:"duration_ms"`
}

func (TaskLog) TableName() string { return "scheduled_task_logs" }

// JobOption 是后台可选择的安全任务执行器。
type JobOption struct {
	Key  string `json:"key"`  // 任务标识
	Name string `json:"name"` // 任务名称
}

type taskExecutor func() (string, error)

var executors = map[string]taskExecutor{
	"goroutine_watcher": func() (string, error) {
		num := runtime.NumGoroutine()
		logger.Info("NumGoroutine", zap.Int("num", num))
		return fmt.Sprintf("当前 Goroutine 数量：%d", num), nil
	},
	"reload_rate_limit": func() (string, error) {
		count, err := rateLimitRuntime.ReloadDefault()
		if err != nil {
			return "刷新 API 限流规则失败：" + err.Error(), err
		}
		return fmt.Sprintf("已刷新 %d 条启用的 API 限流规则", count), nil
	},
}

var jobOptions = []JobOption{
	{Key: "goroutine_watcher", Name: "运行时 Goroutine 巡检"},
	{Key: "reload_rate_limit", Name: "刷新 API 限流规则"},
}

// TaskManager 管理内存中的 cron 调度器，并将每次执行记录到数据库。
type TaskManager struct {
	db      *gorm.DB
	cron    *cron.Cron
	mu      sync.Mutex
	started bool
}

var manager = &TaskManager{}

func EnsureSchema(db *gorm.DB) error {
	if db == nil {
		return errors.New("mysql database is nil")
	}
	if err := db.AutoMigrate(&Task{}, &TaskLog{}); err != nil {
		return err
	}
	if err := seedDefaultTasks(db); err != nil {
		return err
	}
	return ensureManagementMenu(db)
}

// seedDefaultTasks 将原先写死在服务启动代码中的两个任务迁移为可在后台维护的默认任务。
// 仅补齐缺少的执行器，后续 Cron、状态和备注均以后台配置为准。
func seedDefaultTasks(db *gorm.DB) error {
	now := time.Now()
	defaults := []Task{
		{
			Name: "运行时 Goroutine 巡检", JobKey: "goroutine_watcher", CronExpr: "* * * * *", Status: taskStatusEnabled,
			Remark: "原服务内置任务：每分钟记录当前 Goroutine 数量", CreatedAt: now, UpdatedAt: now,
		},
		{
			Name: "API 限流规则刷新", JobKey: "reload_rate_limit", CronExpr: "* * * * *", Status: taskStatusEnabled,
			Remark: "原服务内置任务：每分钟从数据库刷新启用的 API 限流规则", CreatedAt: now, UpdatedAt: now,
		},
	}
	for _, task := range defaults {
		var count int64
		if err := db.Model(&Task{}).Where("job_key = ?", task.JobKey).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.Create(&task).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// ensureManagementMenu 自动补齐后台动态菜单，避免新功能发布后因未导入菜单而不可见。
func ensureManagementMenu(db *gorm.DB) error {
	const menuName = "scheduledTask"
	var count int64
	if err := db.Table("sys_base_menus").Where("name = ? AND deleted_at IS NULL", menuName).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		var parentID uint
		if err := db.Table("sys_base_menus").Where("name = ? AND deleted_at IS NULL", "observability").Select("id").Scan(&parentID).Error; err != nil {
			return err
		}
		// 运行监控父菜单由现有初始化数据提供；不存在时保留在一级菜单，保证页面仍可访问。
		menu := map[string]any{
			"menu_level": 2, "parent_id": parentID, "path": "scheduled-task", "name": menuName,
			"hidden": false, "component": "view/superAdmin/scheduledTask/index.vue", "sort": 3,
			"active_name": "", "keep_alive": false, "default_menu": false, "title": "定时任务管理",
			"icon": "timer", "close_tab": false, "transition_type": "", "created_at": time.Now(), "updated_at": time.Now(),
		}
		if err := db.Table("sys_base_menus").Create(menu).Error; err != nil {
			return err
		}
	}

	// 默认授予超级管理员（角色 888）；其他角色可在后台“角色管理”中按需授权。
	if err := db.Exec(`
		INSERT INTO sys_authority_menus (sys_base_menu_id, sys_authority_authority_id)
		SELECT menu.id, 888
		FROM sys_base_menus AS menu
		WHERE menu.name = ? AND menu.deleted_at IS NULL
		  AND NOT EXISTS (
				SELECT 1 FROM sys_authority_menus AS relation
				WHERE relation.sys_base_menu_id = menu.id
				  AND relation.sys_authority_authority_id = 888
			  )`, menuName).Error; err != nil {
		return err
	}

	// 管理后台接口由 Casbin 单独校验，菜单授权不能替代接口策略。
	policies := []struct {
		path   string
		method string
	}{
		{"/system/scheduled-task/options", "GET"},
		{"/system/scheduled-task/create", "POST"},
		{"/system/scheduled-task/update", "PUT"},
		{"/system/scheduled-task/delete", "DELETE"},
		{"/system/scheduled-task/detail", "GET"},
		{"/system/scheduled-task/pageList", "GET"},
		{"/system/scheduled-task/execute", "POST"},
		{"/system/scheduled-task/logPageList", "GET"},
	}
	for _, policy := range policies {
		if err := db.Exec(`
			INSERT INTO casbin_rule (ptype, v0, v1, v2)
			SELECT 'p', '888', ?, ?
			WHERE NOT EXISTS (
				SELECT 1 FROM casbin_rule
				WHERE ptype = 'p' AND v0 = '888' AND v1 = ? AND v2 = ?
			)`, policy.path, policy.method, policy.path, policy.method).Error; err != nil {
			return err
		}
	}
	return nil
}

// StartTaskManager 从数据库加载所有已启用任务。Cron 使用标准五段表达式，例如：0 2 * * *。
func StartTaskManager() error {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	if manager.started {
		return nil
	}
	manager.db = config.GetMysqlDB()
	if manager.db == nil {
		return errors.New("mysql database is nil")
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	manager.cron = cron.New(cron.WithParser(parser))
	var tasks []Task
	if err := manager.db.Where("status = ?", taskStatusEnabled).Find(&tasks).Error; err != nil {
		return err
	}
	for _, task := range tasks {
		if err := manager.addLocked(task); err != nil {
			return fmt.Errorf("加载定时任务 %s 失败: %w", task.Name, err)
		}
	}
	manager.cron.Start()
	manager.started = true
	return nil
}

func (m *TaskManager) reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.started {
		return nil
	}
	m.cron.Stop()
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	m.cron = cron.New(cron.WithParser(parser))
	var tasks []Task
	if err := m.db.Where("status = ?", taskStatusEnabled).Find(&tasks).Error; err != nil {
		return err
	}
	for _, task := range tasks {
		if err := m.addLocked(task); err != nil {
			return fmt.Errorf("加载定时任务 %s 失败: %w", task.Name, err)
		}
	}
	m.cron.Start()
	return nil
}

func (m *TaskManager) addLocked(task Task) error {
	if _, ok := executors[task.JobKey]; !ok {
		return fmt.Errorf("未注册的任务执行器：%s", task.JobKey)
	}
	_, err := m.cron.AddFunc(task.CronExpr, func() { m.execute(task, "cron") })
	return err
}

func (m *TaskManager) execute(task Task, trigger string) TaskLog {
	startedAt := time.Now()
	lock, err := acquireTaskLock(task.ID)
	if err != nil {
		return m.saveLog(task, trigger, logStatusFailed, "未执行：获取 Redis 分布式锁失败："+err.Error(), startedAt)
	}
	if lock == nil {
		return m.saveLog(task, trigger, logStatusSkipped, "未执行：任务正在其他服务实例中运行", startedAt)
	}
	defer lock.Release()

	output, err := executors[task.JobKey]()
	endedAt := time.Now()
	status := logStatusSuccess
	if err != nil {
		status = logStatusFailed
	}
	return m.saveLog(task, trigger, status, output, startedAt, endedAt)
}

func (m *TaskManager) saveLog(task Task, trigger string, status int, output string, startedAt time.Time, endedAt ...time.Time) TaskLog {
	if len(endedAt) == 0 {
		endedAt = []time.Time{time.Now()}
	}
	log := TaskLog{TaskID: task.ID, TaskName: task.Name, Trigger: trigger, Status: status, Output: output,
		StartedAt: startedAt, EndedAt: endedAt[0], Duration: endedAt[0].Sub(startedAt).Milliseconds()}
	if createErr := m.db.Create(&log).Error; createErr != nil {
		logger.Error("保存定时任务执行日志失败")
	}
	return log
}

// taskLock 用唯一 token 维护 Redis 锁；续期和释放均通过 Lua 校验 token，避免误删其他机器的锁。
type taskLock struct {
	client *taskCache.RedisCacheManager
	key    string
	token  string
	stop   chan struct{}
}

var (
	releaseTaskLockScript = `if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) end return 0`
	renewTaskLockScript   = `if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('pexpire', KEYS[1], ARGV[2]) end return 0`
)

// acquireTaskLock 尝试为一条任务获得集群级执行锁。Redis 不可用时拒绝执行，保证不会退化为多机重复执行。
func acquireTaskLock(taskID uint) (*taskLock, error) {
	redisManager, err := taskCache.DefaultRedisCacheManager()
	if err != nil || redisManager == nil || redisManager.DefaultStore == nil {
		return nil, errors.New("Redis 锁服务不可用")
	}
	lock := &taskLock{
		client: redisManager,
		key:    fmt.Sprintf("scheduled-task:execute:%d", taskID),
		token:  uuid.NewString(),
		stop:   make(chan struct{}),
	}
	locked, err := lock.client.DefaultStore.SetNX(context.Background(), lock.key, lock.token, taskLockTTL)
	if err != nil || !locked {
		return nil, err
	}
	go lock.renew()
	return lock, nil
}

func (l *taskLock) renew() {
	ticker := time.NewTicker(taskLockTTL / 3)
	defer ticker.Stop()
	for {
		select {
		case <-l.stop:
			return
		case <-ticker.C:
			if _, err := l.client.DefaultStore.Raw().Eval(context.Background(), renewTaskLockScript, []string{l.key}, l.token, taskLockTTL.Milliseconds()).Result(); err != nil {
				logger.Error("定时任务 Redis 锁续期失败", zap.String("key", l.key), zap.Error(err))
			}
		}
	}
}

func (l *taskLock) Release() {
	close(l.stop)
	if _, err := l.client.DefaultStore.Raw().Eval(context.Background(), releaseTaskLockScript, []string{l.key}, l.token).Result(); err != nil {
		logger.Error("释放定时任务 Redis 锁失败", zap.String("key", l.key), zap.Error(err))
	}
}
