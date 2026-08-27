package service

import (
	"context"
	"encoding/json"
	"megin/internal/base"
	"megin/internal/system/dto"
	"megin/internal/system/model"
	repo "megin/internal/system/repository"
	"megin/pkg/context/api"
	"time"

	"gorm.io/gorm"
)

type SysVersion struct {
	base.Service
	Repo *repo.SysVersion
}

func NewSysVersion(ctx *api.Context) *SysVersion {
	s := &SysVersion{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysVersion(ctx)
	return s
}

func (s *SysVersion) CreateSysVersion(ctx context.Context, sysVersion *model.SysVersion) error {
	return s.Repo.Create(sysVersion)
}

func (s *SysVersion) DeleteSysVersion(ctx context.Context, id string) error {
	return s.Repo.DB().Delete(&model.SysVersion{}, "id = ?", id).Error
}

func (s *SysVersion) DeleteSysVersionByIds(ctx context.Context, ids []string) error {
	return s.Repo.DB().Where("id in ?", ids).Delete(&model.SysVersion{}).Error
}

func (s *SysVersion) GetSysVersion(ctx context.Context, id string) (*dto.SysVersion, error) {
	var version model.SysVersion
	if err := s.Repo.DB().Where("id = ?", id).First(&version).Error; err != nil {
		return nil, err
	}
	return s.toDTO(version), nil
}

func (s *SysVersion) GetSysVersionInfoList(ctx context.Context, req *dto.SysVersionSearch) (*dto.PageResult[dto.SysVersion], error) {
	query := s.Repo.DB().Model(&model.SysVersion{})
	if len(req.CreatedAtRange) == 2 {
		query = query.Where("created_at BETWEEN ? AND ?", req.CreatedAtRange[0], req.CreatedAtRange[1])
	}
	if req.VersionName != nil && *req.VersionName != "" {
		query = query.Where("version_name LIKE ?", "%"+*req.VersionName+"%")
	}
	if req.VersionCode != nil && *req.VersionCode != "" {
		query = query.Where("version_code = ?", *req.VersionCode)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []model.SysVersion
	if err := query.Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]dto.SysVersion, len(rows))
	for i, row := range rows {
		items[i] = *s.toDTO(row)
	}

	return &dto.PageResult[dto.SysVersion]{
		PageNo:    req.PageNo,
		PageSize:  req.PageSize,
		TotalSize: total,
		List:      items,
	}, nil
}

func (s *SysVersion) GetSysVersionPublic(ctx context.Context) {
}

func (s *SysVersion) GetMenusByIds(ctx context.Context, ids []uint) ([]model.SysBaseMenu, error) {
	var menus []model.SysBaseMenu
	if err := s.Repo.DB().Where("id in ?", ids).Preload("Parameters").Preload("MenuBtn").Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}

func (s *SysVersion) GetApisByIds(ctx context.Context, ids []uint) ([]model.SysApi, error) {
	var apis []model.SysApi
	if err := s.Repo.DB().Where("id in ?", ids).Find(&apis).Error; err != nil {
		return nil, err
	}
	return apis, nil
}

func (s *SysVersion) GetDictionariesByIds(ctx context.Context, ids []uint) ([]model.SysDictionary, error) {
	var dicts []model.SysDictionary
	if err := s.Repo.DB().Where("id in ?", ids).Preload("SysDictionaryDetails").Find(&dicts).Error; err != nil {
		return nil, err
	}
	return dicts, nil
}

func (s *SysVersion) ImportMenus(ctx context.Context, menus []model.SysBaseMenu) error {
	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		return s.createMenusRecursively(tx, menus, 0)
	})
}

func (s *SysVersion) ImportApis(ctx context.Context, apis []model.SysApi) error {
	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		for _, api := range apis {
			var existing model.SysApi
			if err := tx.Where("path = ? AND method = ?", api.Path, api.Method).First(&existing).Error; err == nil {
				continue
			}
			newAPI := model.SysApi{
				Path:        api.Path,
				Description: api.Description,
				ApiGroup:    api.ApiGroup,
				Method:      api.Method,
			}
			if err := tx.Create(&newAPI).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SysVersion) ImportDictionaries(ctx context.Context, dicts []model.SysDictionary) error {
	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		for _, dict := range dicts {
			var existing model.SysDictionary
			if err := tx.Where("type = ?", dict.Type).First(&existing).Error; err == nil {
				continue
			}
			newDict := model.SysDictionary{
				Name:                 dict.Name,
				Type:                 dict.Type,
				Status:               dict.Status,
				Desc:                 dict.Desc,
				SysDictionaryDetails: dict.SysDictionaryDetails,
			}
			if err := tx.Create(&newDict).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SysVersion) createMenusRecursively(tx *gorm.DB, menus []model.SysBaseMenu, parentID uint) error {
	for _, menu := range menus {
		var existing model.SysBaseMenu
		if err := tx.Where("name = ? AND path = ?", menu.Name, menu.Path).First(&existing).Error; err == nil {
			if len(menu.Children) > 0 {
				if err := s.createMenusRecursively(tx, menu.Children, existing.ID); err != nil {
					return err
				}
			}
			continue
		}

		newMenu := model.SysBaseMenu{
			ParentId:  parentID,
			Path:      menu.Path,
			Name:      menu.Name,
			Hidden:    menu.Hidden,
			Component: menu.Component,
			Sort:      menu.Sort,
			Meta:      menu.Meta,
		}
		if err := tx.Create(&newMenu).Error; err != nil {
			return err
		}

		for _, param := range menu.Parameters {
			newParam := model.SysBaseMenuParameter{
				SysBaseMenuID: newMenu.ID,
				Type:          param.Type,
				Key:           param.Key,
				Value:         param.Value,
			}
			if err := tx.Create(&newParam).Error; err != nil {
				return err
			}
		}

		for _, btn := range menu.MenuBtn {
			newBtn := model.SysBaseMenuBtn{
				SysBaseMenuID: newMenu.ID,
				Name:          btn.Name,
				Desc:          btn.Desc,
			}
			if err := tx.Create(&newBtn).Error; err != nil {
				return err
			}
		}

		if len(menu.Children) > 0 {
			if err := s.createMenusRecursively(tx, menu.Children, newMenu.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SysVersion) toDTO(version model.SysVersion) *dto.SysVersion {
	return &dto.SysVersion{
		ID:          version.ID,
		VersionName: version.VersionName,
		VersionCode: version.VersionCode,
		Description: version.Description,
		VersionData: version.VersionData,
		CreatedAt:   version.CreatedAt,
		UpdatedAt:   version.UpdatedAt,
	}
}

func (s *SysVersion) ExportVersionPayload(versionName, versionCode, description string, menus []model.SysBaseMenu, apis []model.SysApi, dicts []model.SysDictionary) ([]byte, error) {
	exportData := dto.ExportVersionResponse{
		Version: dto.VersionInfo{
			Name:        versionName,
			Code:        versionCode,
			Description: description,
			ExportTime:  time.Now().Format("2006-01-02 15:04:05"),
		},
		Menus:        menus,
		Apis:         apis,
		Dictionaries: dicts,
	}
	return json.MarshalIndent(exportData, "", "  ")
}

func strPtr(v string) *string {
	return &v
}
