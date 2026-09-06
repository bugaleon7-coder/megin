package service

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"megin/internal/config"
	commonDto "megin/internal/dto"
	"megin/internal/profiler"
	systemDto "megin/internal/system/dto"
	"megin/internal/system/model"
	"megin/pkg/errs"

	"gorm.io/gorm"
)

const (
	PprofStatusWaiting    = "waiting"
	PprofStatusCollecting = "collecting"
	PprofStatusGenerating = "generating"
	PprofStatusCompleted  = "completed"
	PprofStatusFailed     = "failed"
)

type pprofJob struct {
	ID      uint
	Seconds int
}

var (
	pprofQueue      = make(chan pprofJob, 100)
	pprofWorkerOnce sync.Once
)

type storedProfileBundle struct {
	Profiles map[string]*profiler.CPUProfile `json:"profiles"`
	Errors   map[string]string               `json:"errors,omitempty"`
}

// InitPprofJobs 启动单实例采样队列，并结束服务重启前未完成的记录。
func InitPprofJobs() error {
	db := config.GetMysqlDB()
	if db == nil {
		return errors.New("数据库连接未初始化")
	}
	finishedAt := time.Now()
	if err := db.Model(&model.SysPprofRecord{}).
		Where("status IN ?", []string{PprofStatusWaiting, PprofStatusCollecting, PprofStatusGenerating}).
		Updates(map[string]any{"status": PprofStatusFailed, "finished_at": finishedAt, "error_message": "服务重启，采样任务已中断"}).Error; err != nil {
		return fmt.Errorf("恢复 Pprof 采样任务状态失败: %w", err)
	}
	pprofWorkerOnce.Do(func() { go runPprofWorker() })
	return nil
}

func PprofStatus() profiler.Status { return profiler.Default().Status() }

func TogglePprof(enable bool) (profiler.Status, error) {
	if err := profiler.Default().SetEnabled(enable); err != nil {
		return profiler.Default().Status(), errs.NewBusinessError(4000, err.Error())
	}
	return profiler.Default().Status(), nil
}

func StartPprofProfile(seconds int, createdBy uint) (systemDto.PprofRecord, error) {
	if seconds != 30 && seconds != 60 && seconds != 300 && seconds != 600 {
		return systemDto.PprofRecord{}, errs.NewBusinessError(4000, "采样时长仅支持 30 秒、1 分钟、5 分钟或 10 分钟")
	}
	if !profiler.Default().Status().Enabled {
		return systemDto.PprofRecord{}, errs.NewBusinessError(4000, "请先开启 pprof")
	}
	now := time.Now()
	record := model.SysPprofRecord{ProfileType: "all", Status: PprofStatusWaiting, DurationSeconds: seconds, CreatedBy: createdBy}
	record.CreatedAt, record.UpdatedAt = &now, &now
	if err := config.GetMysqlDB().Create(&record).Error; err != nil {
		return systemDto.PprofRecord{}, errs.NewBusinessError(500, "创建采样记录失败："+err.Error())
	}
	select {
	case pprofQueue <- pprofJob{ID: record.ID, Seconds: seconds}:
	default:
		finishPprofJob(record.ID, errors.New("采样队列已满"))
		return systemDto.PprofRecord{}, errs.NewBusinessError(4000, "采样队列已满，请稍后重试")
	}
	return toPprofRecord(record), nil
}

func PagePprofRecords(req *systemDto.PprofRecordPageReq) (commonDto.PageResult[systemDto.PprofRecord], error) {
	db := config.GetMysqlDB().Model(&model.SysPprofRecord{})
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return commonDto.PageResult[systemDto.PprofRecord]{}, errs.NewBusinessError(500, "查询采样记录失败："+err.Error())
	}
	var records []model.SysPprofRecord
	if err := db.Order("id DESC").Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).Find(&records).Error; err != nil {
		return commonDto.PageResult[systemDto.PprofRecord]{}, errs.NewBusinessError(500, "查询采样记录失败："+err.Error())
	}
	items := make([]systemDto.PprofRecord, len(records))
	for index := range records {
		items[index] = toPprofRecord(records[index])
	}
	totalPage := int64(0)
	if req.PageSize > 0 {
		totalPage = (total + int64(req.PageSize) - 1) / int64(req.PageSize)
	}
	return commonDto.PageResult[systemDto.PprofRecord]{PageNo: req.PageNo, PageSize: req.PageSize, TotalSize: total, TotalPage: totalPage, List: items}, nil
}

func GetPprofRecord(id uint) (systemDto.PprofRecordDetail, error) {
	var record model.SysPprofRecord
	if err := config.GetMysqlDB().First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return systemDto.PprofRecordDetail{}, errs.NewBusinessError(404, "采样记录不存在")
		}
		return systemDto.PprofRecordDetail{}, errs.NewBusinessError(500, "查询采样记录失败："+err.Error())
	}
	detail := systemDto.PprofRecordDetail{Record: toPprofRecord(record)}
	if record.Status != PprofStatusCompleted || record.FlameFile == "" {
		return detail, nil
	}
	bundle, err := readProfileBundle(record.FlameFile)
	if err != nil {
		return detail, errs.NewBusinessError(500, "读取分析数据失败："+err.Error())
	}
	detail.Profiles = bundle.Profiles
	detail.ProfileErrors = bundle.Errors
	detail.Profile = bundle.Profiles["cpu"]
	return detail, nil
}

func LatestPprofRecord() (systemDto.PprofRecordDetail, error) {
	var record model.SysPprofRecord
	if err := config.GetMysqlDB().Order("id DESC").First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return systemDto.PprofRecordDetail{}, nil
		}
		return systemDto.PprofRecordDetail{}, err
	}
	return GetPprofRecord(record.ID)
}

func runPprofWorker() {
	for job := range pprofQueue {
		runPprofJob(job)
	}
}

func runPprofJob(job pprofJob) {
	startedAt := time.Now()
	db := config.GetMysqlDB()
	if err := db.Model(&model.SysPprofRecord{}).Where("id = ?", job.ID).
		Updates(map[string]any{"status": PprofStatusCollecting, "started_at": startedAt, "error_message": ""}).Error; err != nil {
		finishPprofJob(job.ID, err)
		return
	}
	captures, err := profiler.Default().CaptureAll(context.Background(), job.Seconds)
	if err != nil {
		finishPprofJob(job.ID, err)
		return
	}
	if err := db.Model(&model.SysPprofRecord{}).Where("id = ?", job.ID).Update("status", PprofStatusGenerating).Error; err != nil {
		finishPprofJob(job.ID, err)
		return
	}
	rawFile, flameFile, bundle, err := writePprofFiles(job.ID, captures)
	if err != nil {
		finishPprofJob(job.ID, err)
		return
	}
	finishedAt := time.Now()
	var sampleTotal int64
	var sampleUnit string
	if cpu := bundle.Profiles["cpu"]; cpu != nil {
		sampleTotal = cpu.Total
		sampleUnit = cpu.SampleUnit
	}
	if err := db.Model(&model.SysPprofRecord{}).Where("id = ?", job.ID).Updates(map[string]any{
		"status": PprofStatusCompleted, "finished_at": finishedAt, "raw_file": rawFile,
		"flame_file": flameFile, "sample_total": sampleTotal, "sample_unit": sampleUnit,
	}).Error; err != nil {
		finishPprofJob(job.ID, err)
	}
}

func finishPprofJob(id uint, err error) {
	finishedAt := time.Now()
	_ = config.GetMysqlDB().Model(&model.SysPprofRecord{}).Where("id = ?", id).Updates(map[string]any{
		"status": PprofStatusFailed, "finished_at": finishedAt, "error_message": err.Error(),
	}).Error
}

func writePprofFiles(id uint, captures []profiler.CaptureResult) (string, string, storedProfileBundle, error) {
	directory, err := filepath.Abs(config.GetConfig().PprofDirectory())
	if err != nil {
		return "", "", storedProfileBundle{}, err
	}
	recordDirectory := filepath.Join(directory, fmt.Sprintf("%s_%d", time.Now().Format("20060102_150405"), id))
	if err := os.MkdirAll(recordDirectory, 0o750); err != nil {
		return "", "", storedProfileBundle{}, err
	}
	bundle := storedProfileBundle{Profiles: map[string]*profiler.CPUProfile{}, Errors: map[string]string{}}
	for _, capture := range captures {
		if len(capture.Raw) > 0 {
			rawPath := filepath.Join(recordDirectory, capture.ProfileType+capture.RawExt)
			if err := os.WriteFile(rawPath, capture.Raw, 0o600); err != nil {
				return "", "", storedProfileBundle{}, err
			}
		}
		if capture.Profile != nil {
			bundle.Profiles[capture.ProfileType] = capture.Profile
		}
		if capture.Error != "" {
			bundle.Errors[capture.ProfileType] = capture.Error
		}
	}
	flameFile := filepath.Join(recordDirectory, "profiles.json.gz")
	file, err := os.OpenFile(flameFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return "", "", storedProfileBundle{}, err
	}
	zipWriter := gzip.NewWriter(file)
	encodeErr := json.NewEncoder(zipWriter).Encode(bundle)
	closeZipErr := zipWriter.Close()
	closeFileErr := file.Close()
	if encodeErr != nil {
		return "", "", storedProfileBundle{}, encodeErr
	}
	if closeZipErr != nil {
		return "", "", storedProfileBundle{}, closeZipErr
	}
	if closeFileErr != nil {
		return "", "", storedProfileBundle{}, closeFileErr
	}
	return recordDirectory, flameFile, bundle, nil
}

func readProfileBundle(path string) (storedProfileBundle, error) {
	directory, err := filepath.Abs(config.GetConfig().PprofDirectory())
	if err != nil {
		return storedProfileBundle{}, err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return storedProfileBundle{}, err
	}
	relative, err := filepath.Rel(directory, absPath)
	if err != nil || relative == ".." || filepath.IsAbs(relative) {
		return storedProfileBundle{}, errors.New("非法的分析文件路径")
	}
	file, err := os.Open(absPath)
	if err != nil {
		return storedProfileBundle{}, err
	}
	defer file.Close()
	zipReader, err := gzip.NewReader(file)
	if err != nil {
		return storedProfileBundle{}, err
	}
	defer zipReader.Close()
	var bundle storedProfileBundle
	if err := json.NewDecoder(io.LimitReader(zipReader, 64<<20)).Decode(&bundle); err != nil {
		return storedProfileBundle{}, err
	}
	return bundle, nil
}

func toPprofRecord(record model.SysPprofRecord) systemDto.PprofRecord {
	return systemDto.PprofRecord{
		ID: record.ID, Status: record.Status, DurationSeconds: record.DurationSeconds,
		StartedAt: record.StartedAt, FinishedAt: record.FinishedAt, SampleTotal: record.SampleTotal,
		SampleUnit: record.SampleUnit, ErrorMessage: record.ErrorMessage, CreatedBy: record.CreatedBy,
		CreatedAt: record.CreatedAt,
	}
}
