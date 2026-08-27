package service

import (
	"megin/internal/base"
	systemDto "megin/internal/system/dto"
	"megin/internal/system/model"
	repo "megin/internal/system/repository"
	"megin/pkg/context/api"
	"megin/pkg/utils"
	"strconv"
	"time"
)

type SysDictionaryDetail struct {
	base.Service
	Repo *repo.SysDictionaryDetail
}

func NewSysDictionaryDetail(ctx *api.Context) *SysDictionaryDetail {
	s := &SysDictionaryDetail{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysDictionaryDetail(ctx)
	return s
}

func (s *SysDictionaryDetail) CreateSysDictionaryDetail(req *systemDto.CreateDictionaryDetailReq) error {
	now := time.Now()
	status := true
	if req.Status != nil {
		status = *req.Status
	}
	detail := model.SysDictionaryDetail{
		Label:           req.Label,
		Value:           req.Value,
		Extend:          req.Extend,
		Status:          &status,
		Sort:            req.Sort,
		SysDictionaryID: req.SysDictionaryID,
		ParentID:        req.ParentID,
		Level:           0,
		Path:            "",
		SystemModel: base.SystemModel{
			CreatedAt: utils.TimePtr(now),
			UpdatedAt: utils.TimePtr(now),
		},
	}

	// If parent is set, calculate level and path
	if req.ParentID != nil && *req.ParentID > 0 {
		parent, err := s.Repo.GetById(*req.ParentID)
		if err == nil && parent.ID > 0 {
			detail.Level = parent.Level + 1
			detail.Path = parent.Path
		}
	}
	if detail.Path == "" {
		detail.Path = "0"
	}

	return s.Repo.Create(&detail)
}

func (s *SysDictionaryDetail) DeleteSysDictionaryDetail(id uint) error {
	detail, err := s.Repo.GetById(id)
	if err != nil || detail.ID == 0 {
		return s.ErrorMessage("字典详情不存在")
	}
	children, err := s.Repo.GetChildren(id)
	if err != nil {
		return s.Error(err, "删除字典详情失败")
	}
	if len(children) > 0 {
		return s.ErrorMessage("该字典详情下还有子项，无法删除")
	}
	return s.Repo.DeleteById(id)
}

func (s *SysDictionaryDetail) UpdateSysDictionaryDetail(req *systemDto.UpdateDictionaryDetailReq) error {
	detail, err := s.Repo.GetById(req.ID)
	if err != nil || detail.ID == 0 {
		return s.ErrorMessage("字典详情不存在")
	}
	detail.Label = req.Label
	detail.Value = req.Value
	detail.Extend = req.Extend
	detail.Status = req.Status
	detail.Sort = req.Sort
	detail.SysDictionaryID = req.SysDictionaryID
	detail.ParentID = req.ParentID
	if req.ParentID != nil {
		parent, err := s.Repo.GetById(*req.ParentID)
		if err != nil || parent.ID == 0 {
			return s.ErrorMessage("父级字典详情不存在")
		}
		if s.checkCircularReference(req.ID, *req.ParentID) {
			return s.ErrorMessage("不能将字典详情设置为自己或其子项的父级")
		}
		detail.Level = parent.Level + 1
		if parent.Path == "" {
			detail.Path = strconv.Itoa(int(parent.ID))
		} else {
			detail.Path = parent.Path + "," + strconv.Itoa(int(parent.ID))
		}
	} else {
		detail.Level = 0
		detail.Path = ""
	}
	detail.UpdatedAt = utils.TimePtr(time.Now())
	if err := s.Repo.Save(&detail); err != nil {
		return err
	}
	return s.updateChildrenLevelAndPath(detail.ID)
}

func (s *SysDictionaryDetail) GetSysDictionaryDetail(id uint) (*systemDto.SysDictionaryDetail, error) {
	detail, err := s.Repo.GetById(id)
	if err != nil {
		return nil, err
	}
	if detail.ID == 0 {
		return nil, s.ErrorMessage("字典详情不存在")
	}
	return s.toDTO(detail), nil
}

func (s *SysDictionaryDetail) GetSysDictionaryDetailInfoList(req *systemDto.DictionaryDetailSearchReq) (*systemDto.PageResult[systemDto.SysDictionaryDetail], error) {
	query := s.Repo.DB().Model(&model.SysDictionaryDetail{})
	if req.Label != "" {
		query = query.Where("label LIKE ?", "%"+req.Label+"%")
	}
	if req.Value != "" {
		query = query.Where("value LIKE ?", "%"+req.Value+"%")
	}
	if req.SysDictionaryID != 0 {
		query = query.Where("sys_dictionary_id = ?", req.SysDictionaryID)
	}
	if req.ParentID != nil {
		query = query.Where("parent_id = ?", *req.ParentID)
	}
	if req.Level != nil {
		query = query.Where("level = ?", *req.Level)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, s.Error(err, "查询字典详情列表失败")
	}

	var details []model.SysDictionaryDetail
	if err := query.Offset((req.PageNo - 1) * req.PageSize).Limit(req.PageSize).Order("sort ASC").Find(&details).Error; err != nil {
		return nil, s.Error(err, "查询字典详情列表失败")
	}

	dtos := make([]systemDto.SysDictionaryDetail, len(details))
	for i, d := range details {
		dtos[i] = *s.toDTO(d)
	}

	return &systemDto.PageResult[systemDto.SysDictionaryDetail]{
		PageNo:    req.PageNo,
		PageSize:  req.PageSize,
		TotalSize: total,
		List:      dtos,
	}, nil
}

func (s *SysDictionaryDetail) GetDetailsByDictionaryID(dictID int) ([]systemDto.SysDictionaryDetail, error) {
	details, err := s.Repo.GetByDictionaryID(dictID)
	if err != nil {
		return nil, s.Error(err, "查询字典详情失败")
	}
	// Build tree
	tree := s.buildDetailTree(details)
	return tree, nil
}

func (s *SysDictionaryDetail) GetDetailsByParent(req *systemDto.GetDetailsByParentReq) ([]systemDto.SysDictionaryDetail, error) {
	details, err := s.Repo.GetByParentAndDict(req.SysDictionaryID, req.ParentID)
	if err != nil {
		return nil, s.Error(err, "查询字典详情失败")
	}

	if req.IncludeChildren {
		var result []systemDto.SysDictionaryDetail
		for _, d := range details {
			item := s.toDTO(d)
			children, _ := s.Repo.GetByParentAndDict(req.SysDictionaryID, &d.ID)
			item.Children = make([]systemDto.SysDictionaryDetail, len(children))
			for i, c := range children {
				item.Children[i] = *s.toDTO(c)
			}
			result = append(result, *item)
		}
		return result, nil
	}

	dtos := make([]systemDto.SysDictionaryDetail, len(details))
	for i, d := range details {
		dtos[i] = *s.toDTO(d)
	}
	return dtos, nil
}

func (s *SysDictionaryDetail) GetDictionaryTreeList(dictID int) ([]systemDto.SysDictionaryDetail, error) {
	roots, err := s.Repo.GetByParentAndDict(dictID, nil)
	if err != nil {
		return nil, s.Error(err, "查询字典详情失败")
	}
	result := make([]systemDto.SysDictionaryDetail, len(roots))
	for i, root := range roots {
		item := s.toDTO(root)
		children, err := s.loadChildren(root.ID)
		if err != nil {
			return nil, err
		}
		item.Children = children
		result[i] = *item
	}
	return result, nil
}

func (s *SysDictionaryDetail) GetDictionaryTreeListByType(dictType string) ([]systemDto.SysDictionaryDetail, error) {
	var dict model.SysDictionary
	if err := s.Repo.DB().Where("type = ?", dictType).First(&dict).Error; err != nil {
		return nil, s.Error(err, "查询字典详情失败")
	}
	return s.GetDictionaryTreeList(int(dict.ID))
}

func (s *SysDictionaryDetail) GetDictionaryPath(id uint) ([]systemDto.SysDictionaryDetail, error) {
	detail, err := s.Repo.GetById(id)
	if err != nil {
		return nil, s.Error(err, "查询字典详情失败")
	}
	if detail.ID == 0 {
		return nil, s.ErrorMessage("字典详情不存在")
	}
	path := []systemDto.SysDictionaryDetail{*s.toDTO(detail)}
	if detail.ParentID == nil || *detail.ParentID == 0 {
		return path, nil
	}
	parentPath, err := s.GetDictionaryPath(*detail.ParentID)
	if err != nil {
		return nil, err
	}
	return append(parentPath, path...), nil
}

func (s *SysDictionaryDetail) checkCircularReference(id, parentID uint) bool {
	if id == parentID {
		return true
	}
	parent, err := s.Repo.GetById(parentID)
	if err != nil || parent.ID == 0 || parent.ParentID == nil {
		return false
	}
	return s.checkCircularReference(id, *parent.ParentID)
}

func (s *SysDictionaryDetail) updateChildrenLevelAndPath(parentID uint) error {
	children, err := s.Repo.GetChildren(parentID)
	if err != nil {
		return s.Error(err, "更新子级字典详情失败")
	}
	parent, err := s.Repo.GetById(parentID)
	if err != nil || parent.ID == 0 {
		return s.ErrorMessage("父级字典详情不存在")
	}
	for _, child := range children {
		child.Level = parent.Level + 1
		if parent.Path == "" {
			child.Path = strconv.Itoa(int(parent.ID))
		} else {
			child.Path = parent.Path + "," + strconv.Itoa(int(parent.ID))
		}
		child.UpdatedAt = utils.TimePtr(time.Now())
		if err := s.Repo.Save(&child); err != nil {
			return err
		}
		if err := s.updateChildrenLevelAndPath(child.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *SysDictionaryDetail) loadChildren(parentID uint) ([]systemDto.SysDictionaryDetail, error) {
	children, err := s.Repo.GetChildren(parentID)
	if err != nil {
		return nil, s.Error(err, "查询字典详情失败")
	}
	result := make([]systemDto.SysDictionaryDetail, len(children))
	for i, child := range children {
		item := s.toDTO(child)
		grandChildren, err := s.loadChildren(child.ID)
		if err != nil {
			return nil, err
		}
		item.Children = grandChildren
		result[i] = *item
	}
	return result, nil
}

func (s *SysDictionaryDetail) buildDetailTree(details []model.SysDictionaryDetail) []systemDto.SysDictionaryDetail {
	detailMap := make(map[uint]*systemDto.SysDictionaryDetail)
	var roots []systemDto.SysDictionaryDetail

	for _, d := range details {
		item := s.toDTO(d)
		item.Children = []systemDto.SysDictionaryDetail{}
		detailMap[d.ID] = item
	}

	for _, d := range details {
		item := detailMap[d.ID]
		if d.ParentID == nil || *d.ParentID == 0 {
			roots = append(roots, *item)
		} else if parent, ok := detailMap[*d.ParentID]; ok {
			parent.Children = append(parent.Children, *item)
		} else {
			roots = append(roots, *item)
		}
	}

	return roots
}

func (s *SysDictionaryDetail) toDTO(d model.SysDictionaryDetail) *systemDto.SysDictionaryDetail {
	status := false
	if d.Status != nil {
		status = *d.Status
	}
	return &systemDto.SysDictionaryDetail{
		ID:              d.ID,
		Label:           d.Label,
		Value:           d.Value,
		Extend:          d.Extend,
		Status:          &status,
		Disabled:        !status,
		Sort:            d.Sort,
		SysDictionaryID: d.SysDictionaryID,
		ParentID:        d.ParentID,
		Level:           d.Level,
		Path:            d.Path,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}
