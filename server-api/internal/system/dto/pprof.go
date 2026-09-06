package dto

import (
	commonDto "megin/internal/dto"
	"megin/internal/profiler"
	"time"
)

// PprofToggleReq 动态启停当前服务实例的 pprof。
type PprofToggleReq struct {
	Enable bool `json:"enable"`
}

// PprofCPUProfileReq 启动一次异步运行时采样。
type PprofCPUProfileReq struct {
	DurationSeconds int `json:"duration_seconds" binding:"required,oneof=30 60 300 600"`
}

type PprofRecordPageReq struct {
	commonDto.PageQuery
	Status string `json:"status" form:"status" binding:"omitempty,oneof=waiting collecting generating completed failed"`
}

type PprofRecordReq struct {
	ID uint `json:"id" form:"id" binding:"required,min=1"`
}

type PprofRecord struct {
	ID              uint       `json:"id"`
	Status          string     `json:"status"`
	DurationSeconds int        `json:"duration_seconds"`
	StartedAt       *time.Time `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	SampleTotal     int64      `json:"sample_total"`
	SampleUnit      string     `json:"sample_unit"`
	ErrorMessage    string     `json:"error_message"`
	CreatedBy       uint       `json:"created_by"`
	CreatedAt       *time.Time `json:"created_at"`
}

type PprofRecordDetail struct {
	Record        PprofRecord                     `json:"record"`
	Profile       *profiler.CPUProfile            `json:"profile,omitempty"`
	Profiles      map[string]*profiler.CPUProfile `json:"profiles,omitempty"`
	ProfileErrors map[string]string               `json:"profile_errors,omitempty"`
}
