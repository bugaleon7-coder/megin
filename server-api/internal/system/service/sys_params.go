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

type SysParams struct {
	base.Service
	Repo *repo.SysParams
}

func NewSysParams(ctx *api.Context) *SysParams {
	s := &SysParams{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysParams(ctx)
	return s
}

func (s *SysParams) CreateSysParams(req *systemDto.CreateParamsReq) error {
	now := time.Now()
	param := model.SysParams{
		Name:  req.Name,
		Key:   req.Key,
		Value: req.Value,
		Desc:  req.Desc,
		SystemModel: base.SystemModel{
			CreatedAt: utils.TimePtr(now),
			UpdatedAt: utils.TimePtr(now),
		},
	}
	return s.Repo.Create(&param)
}

func (s *SysParams) DeleteSysParams(id uint) error {
	param, err := s.Repo.GetById(id)
	if err != nil || param.ID == 0 {
		return s.ErrorMessage("参数不存在")
	}
	return s.Repo.DeleteById(id)
}

func (s *SysParams) UpdateSysParams(req *systemDto.UpdateParamsReq) error {
	param, err := s.Repo.GetById(req.ID)
	if err != nil || param.ID == 0 {
		return s.ErrorMessage("参数不存在")
	}
	param.Name = req.Name
	param.Key = req.Key
	param.Value = req.Value
	param.Desc = req.Desc
	param.UpdatedAt = utils.TimePtr(time.Now())
	return s.Repo.Save(&param)
}

func (s *SysParams) GetSysParams(id uint) (*systemDto.SysParams, error) {
	param, err := s.Repo.GetById(id)
	if err != nil {
		return nil, err
	}
	if param.ID == 0 {
		return nil, s.ErrorMessage("参数不存在")
	}
	return s.toDTO(param), nil
}

func (s *SysParams) GetSysParamsByKey(key string) (*systemDto.SysParams, error) {
	param, err := s.Repo.GetByKey(key)
	if err != nil {
		return nil, s.Error(err, "查询参数失败")
	}
	if param == nil {
		return nil, s.ErrorMessage("参数不存在")
	}
	return s.toDTO(*param), nil
}

func (s *SysParams) GetSysParamsInfoList(req *systemDto.ParamsSearchReq) (*systemDto.PageResult[systemDto.SysParams], error) {
	query := s.Repo.DB().Model(&model.SysParams{})
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Key != "" {
		query = query.Where("`key` LIKE ?", "%"+req.Key+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, s.Error(err, "查询参数列表失败")
	}

	var params []model.SysParams
	if err := query.Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).Find(&params).Error; err != nil {
		return nil, s.Error(err, "查询参数列表失败")
	}

	dtos := make([]systemDto.SysParams, len(params))
	for i, p := range params {
		dtos[i] = *s.toDTO(p)
	}

	return &systemDto.PageResult[systemDto.SysParams]{
		PageNo:    req.PageNo,
		PageSize:  req.PageSize,
		TotalSize: total,
		List:      dtos,
	}, nil
}

func (s *SysParams) DeleteSysParamsByIds(ids []int) error {
	return s.Repo.DeleteByIds(ids)
}

func (s *SysParams) toDTO(p model.SysParams) *systemDto.SysParams {
	return &systemDto.SysParams{
		ID:        p.ID,
		Name:      p.Name,
		Key:       p.Key,
		Value:     p.Value,
		Desc:      p.Desc,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
