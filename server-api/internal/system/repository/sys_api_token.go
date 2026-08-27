package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"
)

type SysApiToken struct {
	base.Repository[model.SysApiToken]
}

func NewSysApiToken(ctx *api.Context) *SysApiToken {
	r := &SysApiToken{}
	r.Initialize(ctx)
	return r
}
