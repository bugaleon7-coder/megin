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

type SysError struct {
	base.Service
	Repo *repo.SysError
}

func NewSysError(ctx *api.Context) *SysError {
	s := &SysError{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysError(ctx)
	return s
}

func (s *SysError) CreateSysError(req *model.SysError) error {
	now := time.Now()
	req.CreatedAt = utils.TimePtr(now)
	req.UpdatedAt = utils.TimePtr(now)
	return s.Repo.Create(req)
}

func (s *SysError) DeleteSysError(id uint) error {
	errRecord, err := s.Repo.GetById(id)
	if err != nil || errRecord.ID == 0 {
		return s.ErrorMessage("错误日志不存在")
	}
	return s.Repo.DeleteById(id)
}

func (s *SysError) DeleteSysErrorByIds(ids []string) error {
	return s.Repo.DeleteByIds(ids)
}

func (s *SysError) UpdateSysError(req *systemDto.UpdateSysErrorReq) error {
	errRecord, err := s.Repo.GetById(req.ID)
	if err != nil || errRecord.ID == 0 {
		return s.ErrorMessage("错误日志不存在")
	}
	if req.Solution != nil {
		errRecord.Solution = req.Solution
	}
	if req.Status != "" {
		errRecord.Status = req.Status
	}
	errRecord.UpdatedAt = utils.TimePtr(time.Now())
	return s.Repo.Save(&errRecord)
}

func (s *SysError) GetSysError(id uint) (*systemDto.SysError, error) {
	errRecord, e := s.Repo.GetById(id)
	if e != nil {
		return nil, e
	}
	if errRecord.ID == 0 {
		return nil, s.ErrorMessage("错误日志不存在")
	}
	return s.toDTO(errRecord), nil
}

func (s *SysError) GetSysErrorInfoList(req *systemDto.SysErrorSearchReq) (*systemDto.PageResult[systemDto.SysError], error) {
	query := s.Repo.DB().Model(&model.SysError{}).Order("created_at desc")

	if len(req.CreatedAtRange) == 2 {
		query = query.Where("created_at BETWEEN ? AND ?", req.CreatedAtRange[0], req.CreatedAtRange[1])
	}
	if req.Form != nil && *req.Form != "" {
		query = query.Where("form = ?", *req.Form)
	}
	if req.Info != nil && *req.Info != "" {
		query = query.Where("info LIKE ?", "%"+*req.Info+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, s.Error(err, "查询错误日志列表失败")
	}

	var records []model.SysError
	if err := query.Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).Find(&records).Error; err != nil {
		return nil, s.Error(err, "查询错误日志列表失败")
	}

	dtos := make([]systemDto.SysError, len(records))
	for i, r := range records {
		dtos[i] = *s.toDTO(r)
	}

	return &systemDto.PageResult[systemDto.SysError]{
		PageNo:    req.PageNo,
		PageSize:  req.PageSize,
		TotalSize: total,
		List:      dtos,
	}, nil
}

func (s *SysError) GetSysErrorSolution(id uint) error {
	errRecord, e := s.Repo.GetById(id)
	if e != nil || errRecord.ID == 0 {
		return s.ErrorMessage("错误日志不存在")
	}
	statusProcessing := "处理中"
	errRecord.Status = statusProcessing
	errRecord.UpdatedAt = utils.TimePtr(time.Now())
	if err := s.Repo.Save(&errRecord); err != nil {
		return s.Error(err, "更新处理状态失败")
	}

	// 异步处理：1分钟后更新为处理完成
	go func(recordID uint) {
		time.Sleep(1 * time.Minute)
		s2 := NewSysError(s.Ctx)
		rec, err := s2.Repo.GetById(recordID)
		if err != nil || rec.ID == 0 {
			return
		}
		rec.Status = "处理完成"
		rec.UpdatedAt = utils.TimePtr(time.Now())
		_ = s2.Repo.Save(&rec)
	}(id)

	return nil
}

func (s *SysError) toDTO(r model.SysError) *systemDto.SysError {
	return &systemDto.SysError{
		ID:        r.ID,
		Form:      r.Form,
		Info:      r.Info,
		Level:     r.Level,
		Solution:  r.Solution,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}
