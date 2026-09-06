package model

import (
	"megin/internal/base"
	"time"
)

const TableNameSysPprofRecord = "sys_pprof_records"

type SysPprofRecord struct {
	base.SystemModel
	ProfileType     string     `gorm:"column:profile_type;size:32;index;not null" json:"profile_type"`
	Status          string     `gorm:"column:status;size:32;index;not null" json:"status"`
	DurationSeconds int        `gorm:"column:duration_seconds;not null" json:"duration_seconds"`
	StartedAt       *time.Time `gorm:"column:started_at" json:"started_at"`
	FinishedAt      *time.Time `gorm:"column:finished_at" json:"finished_at"`
	RawFile         string     `gorm:"column:raw_file;size:500" json:"-"`
	FlameFile       string     `gorm:"column:flame_file;size:500" json:"-"`
	SampleTotal     int64      `gorm:"column:sample_total" json:"sample_total"`
	SampleUnit      string     `gorm:"column:sample_unit;size:32" json:"sample_unit"`
	ErrorMessage    string     `gorm:"column:error_message;type:text" json:"error_message"`
	CreatedBy       uint       `gorm:"column:created_by;not null" json:"created_by"`
}

func (SysPprofRecord) TableName() string { return TableNameSysPprofRecord }
func (m SysPprofRecord) GetID() any      { return m.ID }
