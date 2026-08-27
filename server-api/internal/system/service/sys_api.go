package service

import (
	"megin/internal/base"
	systemDto "megin/internal/system/dto"
	"megin/internal/system/model"
	repo "megin/internal/system/repository"
	"megin/pkg/context/api"
	"megin/pkg/utils"
	"strings"
	"time"
)

type SysApi struct {
	base.Service
	Repo *repo.SysApi
}

func NewSysApi(ctx *api.Context) *SysApi {
	s := &SysApi{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysApi(ctx)
	return s
}

func (s *SysApi) CreateApi(req *systemDto.CreateApiReq) error {
	now := time.Now()
	api := model.SysApi{
		Path:        req.Path,
		Description: req.Description,
		ApiGroup:    req.ApiGroup,
		Method:      req.Method,
		SystemModel: base.SystemModel{
			CreatedAt: utils.TimePtr(now),
			UpdatedAt: utils.TimePtr(now),
		},
	}
	return s.Repo.Create(&api)
}

func (s *SysApi) DeleteApi(id uint) error {
	api, err := s.Repo.GetById(id)
	if err != nil || api.ID == 0 {
		return s.ErrorMessage("API不存在")
	}
	return s.Repo.DeleteById(id)
}

func (s *SysApi) UpdateApi(req *systemDto.UpdateApiReq) error {
	api, err := s.Repo.GetById(req.ID)
	if err != nil || api.ID == 0 {
		return s.ErrorMessage("API不存在")
	}
	api.Path = req.Path
	api.Description = req.Description
	api.ApiGroup = req.ApiGroup
	api.Method = req.Method
	api.UpdatedAt = utils.TimePtr(time.Now())
	return s.Repo.Save(&api)
}

func (s *SysApi) GetApiById(id uint) (*systemDto.SysApi, error) {
	api, err := s.Repo.GetById(id)
	if err != nil {
		return nil, err
	}
	if api.ID == 0 {
		return nil, s.ErrorMessage("API不存在")
	}
	return s.toDTO(api), nil
}

func (s *SysApi) GetApiList(req *systemDto.GetApiListReq) (*systemDto.PageResult[systemDto.SysApi], error) {
	query := s.Repo.DB().Model(&model.SysApi{})
	if req.Path != "" {
		query = query.Where("path LIKE ?", "%"+req.Path+"%")
	}
	if req.Description != "" {
		query = query.Where("description LIKE ?", "%"+req.Description+"%")
	}
	if req.ApiGroup != "" {
		query = query.Where("api_group = ?", req.ApiGroup)
	}
	if req.Method != "" {
		query = query.Where("method = ?", req.Method)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, s.Error(err, "查询API列表失败")
	}

	var apis []model.SysApi
	if err := query.Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).Find(&apis).Error; err != nil {
		return nil, s.Error(err, "查询API列表失败")
	}

	dtos := make([]systemDto.SysApi, len(apis))
	for i, a := range apis {
		dtos[i] = *s.toDTO(a)
	}

	return &systemDto.PageResult[systemDto.SysApi]{
		PageNo:    req.PageNo,
		PageSize:  req.PageSize,
		TotalSize: total,
		List:      dtos,
	}, nil
}

func (s *SysApi) GetAllApis() ([]systemDto.SysApi, error) {
	apis, err := s.Repo.GetAllApis()
	if err != nil {
		return nil, s.Error(err, "查询所有API失败")
	}
	dtos := make([]systemDto.SysApi, len(apis))
	for i, a := range apis {
		dtos[i] = *s.toDTO(a)
	}
	return dtos, nil
}

func (s *SysApi) GetApiGroups() (*systemDto.ApiGroupResponse, error) {
	apis, err := s.Repo.GetApiGroups()
	if err != nil {
		return nil, s.Error(err, "查询API分组失败")
	}

	groups := make([]string, 0)
	apiGroupMap := make(map[string]string)
	for _, item := range apis {
		isNewGroup := true
		for _, group := range groups {
			if group == item.ApiGroup {
				isNewGroup = false
				break
			}
		}
		if isNewGroup {
			groups = append(groups, item.ApiGroup)
		}

		pathParts := strings.Split(item.Path, "/")
		if len(pathParts) > 1 {
			apiGroupMap[pathParts[1]] = item.ApiGroup
		}
	}

	return &systemDto.ApiGroupResponse{
		Groups:      groups,
		ApiGroupMap: apiGroupMap,
	}, nil
}

func (s *SysApi) DeleteApisByIds(ids []int) error {
	return s.Repo.DeleteByIds(ids)
}

func (s *SysApi) SyncApi() (*systemDto.SyncApiResponse, error) {
	// Get all APIs from DB
	apis, err := s.Repo.GetAllApis()
	if err != nil {
		return nil, s.Error(err, "查询所有API失败")
	}
	// Get ignored APIs
	ignoreApis, err := s.Repo.GetIgnoreApis()
	if err != nil {
		return nil, s.Error(err, "查询忽略API失败")
	}
	ignoreMap := make(map[string]bool)
	for _, ia := range ignoreApis {
		ignoreMap[ia.Path+":"+ia.Method] = true
	}

	newApis := make([]systemDto.SysApi, 0)
	deleteApis := make([]systemDto.SysApi, 0)
	ignoreDTOs := make([]systemDto.SysApi, 0)

	// Build ignoreDtos
	for _, ia := range ignoreApis {
		ignoreDTOs = append(ignoreDTOs, systemDto.SysApi{
			Path:   ia.Path,
			Method: ia.Method,
		})
	}

	// Route-based APIs from the registered routes would come from router info
	// For now, compare DB APIs against a reference (simplified)
	_ = ignoreMap

	_ = apis
	_ = newApis
	_ = deleteApis

	// Simplified: return the data from route registry comparison
	return &systemDto.SyncApiResponse{
		NewApis:    newApis,
		DeleteApis: deleteApis,
		IgnoreApis: ignoreDTOs,
	}, nil
}

func (s *SysApi) IgnoreApi(req *systemDto.IgnoreApiReq) error {
	return s.Repo.UpsertIgnoreApi(req.Path, req.Method, req.Flag)
}

func (s *SysApi) EnterSyncApi(req *systemDto.EnterSyncApiReq) error {
	return s.Repo.Transaction(func(txRepo *repo.SysApi) error {
		// Delete APIs that should be removed
		for _, d := range req.DeleteApis {
			if err := txRepo.DeleteByPathAndMethod(d.Path, d.Method); err != nil {
				return err
			}
		}
		// Create new APIs
		now := time.Now()
		for _, n := range req.NewApis {
			api := model.SysApi{
				Path:   n.Path,
				Method: n.Method,
				SystemModel: base.SystemModel{
					CreatedAt: utils.TimePtr(now),
					UpdatedAt: utils.TimePtr(now),
				},
			}
			if err := txRepo.Create(&api); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SysApi) FreshCasbin() error {
	return NewSysCasbin(s.Ctx).FreshCasbin(0)
}

func (s *SysApi) toDTO(a model.SysApi) *systemDto.SysApi {
	return &systemDto.SysApi{
		ID:          a.ID,
		Path:        a.Path,
		Description: a.Description,
		ApiGroup:    a.ApiGroup,
		Method:      a.Method,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}
