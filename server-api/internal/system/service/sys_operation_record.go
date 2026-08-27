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

type SysOperationRecord struct {
	base.Service
	Repo *repo.SysOperationRecord
}

func NewSysOperationRecord(ctx *api.Context) *SysOperationRecord {
	s := &SysOperationRecord{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysOperationRecord(ctx)
	return s
}

func (s *SysOperationRecord) CreateSysOperationRecord(req *model.SysOperationRecord) error {
	now := time.Now()
	req.CreatedAt = utils.TimePtr(now)
	req.UpdatedAt = utils.TimePtr(now)
	return s.Repo.Create(req)
}

func (s *SysOperationRecord) DeleteSysOperationRecord(id uint) error {
	record, err := s.Repo.GetById(id)
	if err != nil || record.ID == 0 {
		return s.ErrorMessage("操作记录不存在")
	}
	return s.Repo.DeleteById(id)
}

func (s *SysOperationRecord) GetSysOperationRecordInfoList(req *systemDto.OperationRecordSearchReq) (*systemDto.PageResult[systemDto.SysOperationRecord], error) {
	query := s.Repo.DB().Model(&model.SysOperationRecord{}).Preload("User")
	if req.Method != "" {
		query = query.Where("method = ?", req.Method)
	}
	if req.Path != "" {
		query = query.Where("path LIKE ?", "%"+req.Path+"%")
	}
	if req.Status != 0 {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, s.Error(err, "查询操作记录列表失败")
	}

	var records []model.SysOperationRecord
	if err := query.Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).Order("id DESC").Find(&records).Error; err != nil {
		return nil, s.Error(err, "查询操作记录列表失败")
	}

	dtos := make([]systemDto.SysOperationRecord, len(records))
	for i, r := range records {
		dtos[i] = systemDto.SysOperationRecord{
			ID:           r.ID,
			Ip:           r.Ip,
			Method:       r.Method,
			Path:         r.Path,
			Status:       r.Status,
			Latency:      int64(r.Latency),
			Agent:        r.Agent,
			ErrorMessage: r.ErrorMessage,
			Body:         r.Body,
			Resp:         r.Resp,
			UserID:       r.UserID,
			CreatedAt:    r.CreatedAt,
			User: systemDto.SysUser{
				ID:          r.User.ID,
				UUID:        r.User.UUID,
				Username:    r.User.Username,
				NickName:    r.User.NickName,
				HeaderImg:   r.User.HeaderImg,
				AuthorityId: r.User.AuthorityId,
				Phone:       r.User.Phone,
				Email:       r.User.Email,
				Enable:      r.User.Enable,
				CreatedAt:   r.User.CreatedAt,
				UpdatedAt:   r.User.UpdatedAt,
				Authority: systemDto.SysAuthority{
					CreatedAt:       r.User.Authority.CreatedAt,
					UpdatedAt:       r.User.Authority.UpdatedAt,
					AuthorityId:     r.User.Authority.AuthorityId,
					AuthorityName:   r.User.Authority.AuthorityName,
					ParentId:        r.User.Authority.ParentId,
					DefaultRouter:   r.User.Authority.DefaultRouter,
					DataAuthorityId: nil,
					Children:        nil,
				},
				Authorities:   nil,
				OriginSetting: nil,
			},
		}
	}

	return &systemDto.PageResult[systemDto.SysOperationRecord]{
		PageNo:    req.PageNo,
		PageSize:  req.PageSize,
		TotalSize: total,
		List:      dtos,
	}, nil
}

func (s *SysOperationRecord) DeleteSysOperationRecords(ids []int) error {
	return s.Repo.DeleteByIds(ids)
}
