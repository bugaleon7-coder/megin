package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"
)

type SysVersion struct {
	base.Repository[model.SysVersion]
}

func NewSysVersion(ctx *api.Context) *SysVersion {
	r := &SysVersion{}
	r.Initialize(ctx)
	return r
}
