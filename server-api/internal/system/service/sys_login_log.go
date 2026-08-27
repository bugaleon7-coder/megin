package service

import (
	"megin/internal/base"
	systemDto "megin/internal/system/dto"
	"megin/internal/system/model"
	repo "megin/internal/system/repository"
	"megin/pkg/context/api"
	"megin/pkg/utils"
	"time"
)

type SysLoginLog struct {
	base.Service
	Repo *repo.SysLoginLog
}

func NewSysLoginLog(ctx *api.Context) *SysLoginLog {
	s := &SysLoginLog{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysLoginLog(ctx)
	return s
}

func (s *SysLoginLog) CreateSysLoginLog(username string, ip string, agent string, status bool, errorMessage string) error {
	now := time.Now()
	log := model.SysLoginLog{
		Username:     username,
		Ip:           ip,
		Agent:        agent,
		Status:       status,
		ErrorMessage: errorMessage,
		SystemModel: base.SystemModel{
			CreatedAt: utils.TimePtr(now),
			UpdatedAt: utils.TimePtr(now),
		},
	}
	return s.Repo.Create(&log)
}

func (s *SysLoginLog) DeleteSysLoginLog(id uint) error {
	log, err := s.Repo.GetById(id)
	if err != nil || log.ID == 0 {
		return s.ErrorMessage("登录日志不存在")
	}
	return s.Repo.DeleteById(id)
}

func (s *SysLoginLog) GetSysLoginLogInfoList(req *systemDto.LoginLogSearchReq) (*systemDto.PageResult[systemDto.SysLoginLog], error) {
	query := s.Repo.DB().Model(&model.SysLoginLog{})
	if req.Username != "" {
		query = query.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Ip != "" {
		query = query.Where("ip LIKE ?", "%"+req.Ip+"%")
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, s.Error(err, "查询登录日志列表失败")
	}

	var logs []model.SysLoginLog
	if err := query.Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).Order("id DESC").Find(&logs).Error; err != nil {
		return nil, s.Error(err, "查询登录日志列表失败")
	}

	dtos := make([]systemDto.SysLoginLog, len(logs))
	for i, l := range logs {
		dtos[i] = systemDto.SysLoginLog{
			ID:           l.ID,
			Username:     l.Username,
			Ip:           l.Ip,
			Status:       l.Status,
			ErrorMessage: l.ErrorMessage,
			Agent:        l.Agent,
			UserID:       l.UserID,
			CreatedAt:    l.CreatedAt,
		}
	}

	return &systemDto.PageResult[systemDto.SysLoginLog]{
		PageNo:    req.PageNo,
		PageSize:  req.PageSize,
		TotalSize: total,
		List:      dtos,
	}, nil
}

func (s *SysLoginLog) DeleteSysLoginLogs(ids []int) error {
	return s.Repo.DeleteByIds(ids)
}

func (s *SysLoginLog) FindSysLoginLog(id uint) (*systemDto.SysLoginLog, error) {
	log, err := s.Repo.GetById(id)
	if err != nil {
		return nil, s.Error(err, "查询登录日志失败")
	}
	if log.ID == 0 {
		return nil, s.ErrorMessage("登录日志不存在")
	}
	return &systemDto.SysLoginLog{
		ID:           log.ID,
		Username:     log.Username,
		Ip:           log.Ip,
		Status:       log.Status,
		ErrorMessage: log.ErrorMessage,
		Agent:        log.Agent,
		UserID:       log.UserID,
		CreatedAt:    log.CreatedAt,
	}, nil
}
