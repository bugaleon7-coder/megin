package service

import (
	"encoding/json"
	"megin/internal/base"
	systemDto "megin/internal/system/dto"
	"megin/internal/system/model"
	repo "megin/internal/system/repository"
	"megin/pkg/context/api"
	"megin/pkg/utils"
	"time"

	"gorm.io/gorm"
)

type SysDictionary struct {
	base.Service
	Repo *repo.SysDictionary
}

func NewSysDictionary(ctx *api.Context) *SysDictionary {
	s := &SysDictionary{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysDictionary(ctx)
	return s
}

func (s *SysDictionary) CreateSysDictionary(req *systemDto.CreateDictionaryReq) error {
	now := time.Now()
	status := true
	if req.Status != nil {
		status = *req.Status
	}
	dict := model.SysDictionary{
		Name:     req.Name,
		Type:     req.Type,
		Status:   &status,
		Desc:     req.Desc,
		ParentID: req.ParentID,
		SystemModel: base.SystemModel{
			CreatedAt: utils.TimePtr(now),
			UpdatedAt: utils.TimePtr(now),
		},
	}
	return s.Repo.Create(&dict)
}

func (s *SysDictionary) DeleteSysDictionary(id uint) error {
	dict, err := s.Repo.GetById(id)
	if err != nil || dict.ID == 0 {
		return s.ErrorMessage("字典不存在")
	}
	return s.Repo.DeleteById(id)
}

func (s *SysDictionary) UpdateSysDictionary(req *systemDto.UpdateDictionaryReq) error {
	dict, err := s.Repo.GetById(req.ID)
	if err != nil || dict.ID == 0 {
		return s.ErrorMessage("字典不存在")
	}
	dict.Name = req.Name
	dict.Type = req.Type
	dict.Status = req.Status
	dict.Desc = req.Desc
	dict.ParentID = req.ParentID
	dict.UpdatedAt = utils.TimePtr(time.Now())
	return s.Repo.Save(&dict)
}

func (s *SysDictionary) GetSysDictionary(id uint) (*systemDto.SysDictionary, error) {
	dict, err := s.Repo.GetById(id)
	if err != nil {
		return nil, err
	}
	if dict.ID == 0 {
		return nil, s.ErrorMessage("字典不存在")
	}
	return s.toDTO(dict), nil
}

func (s *SysDictionary) FindSysDictionary(req *systemDto.FindDictionaryReq) (*systemDto.SysDictionary, error) {
	status := true
	if req.Status != nil {
		status = *req.Status
	}
	var dict model.SysDictionary
	err := s.Repo.DB().Where("(type = ? OR id = ?) AND status = ?", req.Type, req.ID, status).
		Preload("SysDictionaryDetails", "status = ? AND deleted_at IS NULL", true).
		First(&dict).Error
	if err != nil {
		return nil, s.Error(err, "字典未创建或未开启")
	}
	return s.toDTO(dict), nil
}

func (s *SysDictionary) GetSysDictionaryInfoList(req *systemDto.DictionarySearchReq) (*systemDto.PageResult[systemDto.SysDictionary], error) {
	query := s.Repo.DB().Model(&model.SysDictionary{})
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Type != "" {
		query = query.Where("type LIKE ?", "%"+req.Type+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, s.Error(err, "查询字典列表失败")
	}

	var dicts []model.SysDictionary
	if err := query.Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).Find(&dicts).Error; err != nil {
		return nil, s.Error(err, "查询字典列表失败")
	}

	dtos := make([]systemDto.SysDictionary, len(dicts))
	for i, d := range dicts {
		dtos[i] = *s.toDTO(d)
	}

	return &systemDto.PageResult[systemDto.SysDictionary]{
		PageNo:    req.PageNo,
		PageSize:  req.PageSize,
		TotalSize: total,
		List:      dtos,
	}, nil
}

func (s *SysDictionary) GetSysDictionaryList(req *systemDto.DictionaryListReq) ([]systemDto.SysDictionary, error) {
	query := s.Repo.DB().Preload("Children")
	if req.Name != "" {
		keyword := "%" + req.Name + "%"
		query = query.Where("name LIKE ? OR type LIKE ?", keyword, keyword)
	}

	var dicts []model.SysDictionary
	if err := query.Find(&dicts).Error; err != nil {
		return nil, s.Error(err, "查询字典列表失败")
	}

	result := make([]systemDto.SysDictionary, len(dicts))
	for i, dict := range dicts {
		result[i] = *s.toDTO(dict)
	}
	return result, nil
}

func (s *SysDictionary) ExportSysDictionary() ([]systemDto.SysDictionary, error) {
	var dicts []model.SysDictionary
	if err := s.Repo.DB().Preload("SysDictionaryDetails", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC")
	}).Find(&dicts).Error; err != nil {
		return nil, s.Error(err, "导出字典失败")
	}

	dtos := make([]systemDto.SysDictionary, len(dicts))
	for i, d := range dicts {
		dtos[i] = *s.toDTO(d)
	}
	return dtos, nil
}

func (s *SysDictionary) ImportSysDictionary(jsonStr string) error {
	var dicts []systemDto.SysDictionary
	if err := json.Unmarshal([]byte(jsonStr), &dicts); err != nil {
		return s.ErrorMessage("JSON解析失败")
	}

	for _, d := range dicts {
		now := time.Now()
		dict := model.SysDictionary{
			Name:   d.Name,
			Type:   d.Type,
			Status: d.Status,
			Desc:   d.Desc,
			SystemModel: base.SystemModel{
				CreatedAt: utils.TimePtr(now),
				UpdatedAt: utils.TimePtr(now),
			},
		}
		if err := s.Repo.Create(&dict); err != nil {
			return s.Error(err, "导入字典失败")
		}
	}
	return nil
}

func (s *SysDictionary) toDTO(d model.SysDictionary) *systemDto.SysDictionary {
	details := make([]systemDto.SysDictionaryDetail, len(d.SysDictionaryDetails))
	for i, det := range d.SysDictionaryDetails {
		status := false
		if det.Status != nil {
			status = *det.Status
		}
		details[i] = systemDto.SysDictionaryDetail{
			ID:              det.ID,
			Label:           det.Label,
			Value:           det.Value,
			Extend:          det.Extend,
			Status:          &status,
			Sort:            det.Sort,
			SysDictionaryID: det.SysDictionaryID,
			ParentID:        det.ParentID,
			Level:           det.Level,
			Path:            det.Path,
			CreatedAt:       det.CreatedAt,
			UpdatedAt:       det.UpdatedAt,
		}
	}
	children := make([]systemDto.SysDictionary, len(d.Children))
	for i, child := range d.Children {
		children[i] = *s.toDTO(child)
	}
	return &systemDto.SysDictionary{
		ID:                   d.ID,
		Name:                 d.Name,
		Type:                 d.Type,
		Status:               d.Status,
		Desc:                 d.Desc,
		ParentID:             d.ParentID,
		CreatedAt:            d.CreatedAt,
		UpdatedAt:            d.UpdatedAt,
		Children:             children,
		SysDictionaryDetails: details,
	}
}
