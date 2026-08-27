package service

import (
	"errors"
	"strconv"

	"megin/internal/base"
	"megin/internal/config"
	systemDto "megin/internal/system/dto"
	"megin/internal/system/model"
	repo "megin/internal/system/repository"
	"megin/pkg/context/api"
	"megin/pkg/utils"
	"time"

	"gorm.io/gorm"
)

var ErrRoleExistence = errors.New("存在相同角色id")

type SysAuthority struct {
	base.Service
	Repo      *repo.SysAuthority
	MenuRepo  *repo.SysBaseMenu
	UserRepo  *repo.SysUser
	CasbinSvc *SysCasbin
}

func NewSysAuthority(ctx *api.Context) *SysAuthority {
	s := &SysAuthority{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysAuthority(ctx)
	s.MenuRepo = repo.NewSysBaseMenu(ctx)
	s.UserRepo = repo.NewSysUser(ctx)
	s.CasbinSvc = NewSysCasbin(ctx)
	return s
}

func (s *SysAuthority) authorityIdStr(id uint) string {
	return strconv.Itoa(int(id))
}

func (s *SysAuthority) checkAuthorityIdDuplicate(authorityId uint) error {
	var count int64
	if err := s.Repo.DB().Unscoped().Model(&model.SysAuthority{}).Where("authority_id = ?", authorityId).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrRoleExistence
	}
	return nil
}

// getDefaultMenuIds returns the default menu IDs for a new authority.
func (s *SysAuthority) getDefaultMenuIds() []uint {
	return []uint{1}
}

func (s *SysAuthority) CreateAuthority(adminAuthorityID uint, req *systemDto.CreateAuthorityReq) (*model.SysAuthority, error) {
	if config.GetConfig().System.UseStrictAuth && (req.ParentId == nil || *req.ParentId == 0) {
		req.ParentId = &adminAuthorityID
	}
	if req.ParentId != nil && *req.ParentId != 0 && *req.ParentId != adminAuthorityID {
		if err := s.CheckAuthorityIDAuth(adminAuthorityID, *req.ParentId); err != nil {
			return nil, err
		}
	}
	// Check duplicate authority_id
	if err := s.checkAuthorityIdDuplicate(req.AuthorityId); err != nil {
		return nil, s.Error(err, "创建角色失败")
	}

	now := time.Now()
	auth := model.SysAuthority{
		AuthorityId:   req.AuthorityId,
		AuthorityName: req.AuthorityName,
		ParentId:      req.ParentId,
		DefaultRouter: req.DefaultRouter,
		CreatedAt:     utils.TimePtr(now),
		UpdatedAt:     utils.TimePtr(now),
	}

	err := s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		// 1. Create authority record
		if err := tx.Create(&auth).Error; err != nil {
			return err
		}

		// 2. Set default menus
		menuIds := s.getDefaultMenuIds()
		for _, mid := range menuIds {
			if err := tx.Exec(
				"INSERT INTO sys_authority_menus (sys_base_menu_id, sys_authority_authority_id) VALUES (?, ?)",
				mid, auth.AuthorityId,
			).Error; err != nil {
				return err
			}
		}

		// 3. Set the original least-privilege default policies. Never inherit role 888.
		for _, ci := range systemDto.DefaultCasbinInfos {
			if err := tx.Exec(
				"INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', ?, ?, ?)",
				s.authorityIdStr(auth.AuthorityId), ci.Path, ci.Method,
			).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, s.Error(err, "创建角色失败")
	}
	return &auth, nil
}

func (s *SysAuthority) CopyAuthority(adminAuthorityID uint, req *systemDto.CopyAuthorityReq) (*model.SysAuthority, error) {
	if err := s.CheckAuthorityIDAuth(adminAuthorityID, req.OldAuthorityId); err != nil {
		return nil, err
	}
	if req.Authority.ParentId != nil && *req.Authority.ParentId != 0 && *req.Authority.ParentId != adminAuthorityID {
		if err := s.CheckAuthorityIDAuth(adminAuthorityID, *req.Authority.ParentId); err != nil {
			return nil, err
		}
	}
	if err := s.checkAuthorityIdDuplicate(req.Authority.AuthorityId); err != nil {
		return nil, s.Error(err, "复制角色失败")
	}

	now := time.Now()
	auth := model.SysAuthority{
		AuthorityId:   req.Authority.AuthorityId,
		AuthorityName: req.Authority.AuthorityName,
		ParentId:      req.Authority.ParentId,
		DefaultRouter: req.Authority.DefaultRouter,
		CreatedAt:     utils.TimePtr(now),
		UpdatedAt:     utils.TimePtr(now),
	}

	err := s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		// 1. Create authority
		if err := tx.Create(&auth).Error; err != nil {
			return err
		}

		// 2. Copy menus from old authority
		var menuLinks []struct {
			SysBaseMenuID           uint
			SysAuthorityAuthorityID uint
		}
		if err := tx.Table("sys_authority_menus").
			Where("sys_authority_authority_id = ?", req.OldAuthorityId).
			Find(&menuLinks).Error; err != nil {
			return err
		}
		for _, link := range menuLinks {
			if err := tx.Exec(
				"INSERT INTO sys_authority_menus (sys_base_menu_id, sys_authority_authority_id) VALUES (?, ?)",
				link.SysBaseMenuID, auth.AuthorityId,
			).Error; err != nil {
				return err
			}
		}

		// 3. Copy buttons from old authority
		var oldBtns []model.SysAuthorityBtn
		if err := tx.Where("authority_id = ?", req.OldAuthorityId).Find(&oldBtns).Error; err != nil {
			return err
		}
		for _, btn := range oldBtns {
			if err := tx.Exec(
				"INSERT INTO sys_authority_btns (authority_id, sys_menu_id, sys_base_menu_btn_id) VALUES (?, ?, ?)",
				auth.AuthorityId, btn.SysMenuID, btn.SysBaseMenuBtnID,
			).Error; err != nil {
				return err
			}
		}

		// 4. Copy casbin policies from old authority
		var oldPolicies []struct {
			PType string
			V1    string
			V2    string
		}
		if err := tx.Table("casbin_rule").
			Where("ptype = 'p' AND v0 = ?", s.authorityIdStr(req.OldAuthorityId)).
			Find(&oldPolicies).Error; err != nil {
			return err
		}
		for _, p := range oldPolicies {
			if err := tx.Exec(
				"INSERT INTO casbin_rule (ptype, v0, v1, v2) VALUES ('p', ?, ?, ?)",
				s.authorityIdStr(auth.AuthorityId), p.V1, p.V2,
			).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, s.Error(err, "复制角色失败")
	}
	return &auth, nil
}

func (s *SysAuthority) UpdateAuthority(adminAuthorityID uint, req *systemDto.UpdateAuthorityReq) error {
	if err := s.CheckAuthorityIDAuth(adminAuthorityID, req.AuthorityId); err != nil {
		return err
	}
	if req.ParentId != nil && *req.ParentId != 0 && *req.ParentId != adminAuthorityID {
		if err := s.CheckAuthorityIDAuth(adminAuthorityID, *req.ParentId); err != nil {
			return err
		}
	}
	var auth model.SysAuthority
	err := s.Repo.DB().Where("authority_id = ?", req.AuthorityId).First(&auth).Error
	if err != nil {
		return s.ErrorMessage("查询角色数据失败")
	}
	auth.AuthorityName = req.AuthorityName
	auth.ParentId = req.ParentId
	auth.DefaultRouter = req.DefaultRouter
	auth.UpdatedAt = utils.TimePtr(time.Now())
	return s.Repo.DB().Model(&auth).Updates(map[string]interface{}{
		"authority_name": req.AuthorityName,
		"parent_id":      req.ParentId,
		"default_router": req.DefaultRouter,
		"updated_at":     utils.TimePtr(time.Now()),
	}).Error
}

func (s *SysAuthority) DeleteAuthority(adminAuthorityID, authorityId uint) error {
	if err := s.CheckAuthorityIDAuth(adminAuthorityID, authorityId); err != nil {
		return err
	}
	var auth model.SysAuthority
	err := s.Repo.DB().Where("authority_id = ?", authorityId).First(&auth).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return s.ErrorMessage("该角色不存在")
		}
		return s.Error(err, "查询角色失败")
	}

	// Check user usage first
	var userCount int64
	if err := s.Repo.DB().Model(&model.SysUser{}).Where("authority_id = ?", authorityId).Count(&userCount).Error; err != nil {
		return s.Error(err, "查询角色用户失败")
	}
	if userCount > 0 {
		return s.ErrorMessage("此角色有用户正在使用禁止删除")
	}
	// Also check sys_user_authority
	var uaCount int64
	if err := s.Repo.DB().Model(&model.SysUserAuthority{}).Where("sys_authority_authority_id = ?", authorityId).Count(&uaCount).Error; err != nil {
		return s.Error(err, "查询角色关联失败")
	}
	if uaCount > 0 {
		return s.ErrorMessage("此角色有用户正在使用禁止删除")
	}

	// Check children
	var childCount int64
	if err := s.Repo.DB().Model(&model.SysAuthority{}).Where("parent_id = ?", authorityId).Count(&childCount).Error; err != nil {
		return s.Error(err, "查询子角色失败")
	}
	if childCount > 0 {
		return s.ErrorMessage("此角色存在子角色不允许删除")
	}

	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		// authority_id is unique and reusable after deletion, matching the original hard-delete behavior.
		if err := tx.Unscoped().Where("authority_id = ?", authorityId).Delete(&model.SysAuthority{}).Error; err != nil {
			return err
		}

		statements := []struct {
			query string
			args  []any
		}{
			{"DELETE FROM sys_authority_menus WHERE sys_authority_authority_id = ?", []any{authorityId}},
			{"DELETE FROM sys_data_authority_id WHERE sys_authority_authority_id = ?", []any{authorityId}},
			{"DELETE FROM sys_data_authority_id WHERE data_authority_id_authority_id = ?", []any{authorityId}},
			{"DELETE FROM sys_user_authority WHERE sys_authority_authority_id = ?", []any{authorityId}},
			{"DELETE FROM sys_authority_btns WHERE authority_id = ?", []any{authorityId}},
			{"DELETE FROM casbin_rule WHERE ptype = 'p' AND v0 = ?", []any{s.authorityIdStr(authorityId)}},
		}
		for _, statement := range statements {
			if err := tx.Exec(statement.query, statement.args...).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetAuthorityInfoList returns the role tree (not paginated), matching original behavior.
// adminAuthorityId is used for strict auth scope.
func (s *SysAuthority) GetAuthorityInfoList(adminAuthorityId uint) ([]systemDto.SysAuthority, error) {
	// Get the admin's authority to check parent
	var adminAuth model.SysAuthority
	if err := s.Repo.DB().Where("authority_id = ?", adminAuthorityId).First(&adminAuth).Error; err != nil {
		return nil, s.Error(err, "查询角色失败")
	}

	var authorities []model.SysAuthority
	db := s.Repo.DB().Model(&model.SysAuthority{}).Preload("DataAuthorityId")

	// Apply the original configurable strict-auth behavior.
	if !config.GetConfig().System.UseStrictAuth {
		db = db.Where("parent_id = ?", 0)
	} else if adminAuth.ParentId != nil && *adminAuth.ParentId == 0 {
		// Top-level: can only see self
		db = db.Where("authority_id = ?", adminAuthorityId)
	} else {
		// Non-top-level: see children only
		db = db.Where("parent_id = ?", adminAuthorityId)
	}

	if err := db.Find(&authorities).Error; err != nil {
		return nil, s.Error(err, "查询角色列表失败")
	}

	// Build tree recursively
	for k := range authorities {
		if err := s.findChildren(&authorities[k]); err != nil {
			return nil, s.Error(err, "查询子角色失败")
		}
	}

	// Convert to DTOs preserving tree structure
	result := make([]systemDto.SysAuthority, len(authorities))
	for i, a := range authorities {
		result[i] = s.toDTOWithTree(a)
	}
	return result, nil
}

func (s *SysAuthority) findChildren(authority *model.SysAuthority) error {
	var children []model.SysAuthority
	err := s.Repo.DB().Preload("DataAuthorityId").Where("parent_id = ?", authority.AuthorityId).Find(&children).Error
	if err != nil {
		return err
	}
	if len(children) > 0 {
		authority.Children = make([]model.SysAuthority, len(children))
		copy(authority.Children, children)
		for k := range authority.Children {
			if err = s.findChildren(&authority.Children[k]); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SysAuthority) toDTOWithTree(a model.SysAuthority) systemDto.SysAuthority {
	dto := s.toDTO(a)
	// Copy DataAuthorityId
	if len(a.DataAuthorityId) > 0 {
		dto.DataAuthorityId = make([]*systemDto.SysAuthority, len(a.DataAuthorityId))
		for i, da := range a.DataAuthorityId {
			item := s.toDTO(*da)
			dto.DataAuthorityId[i] = &item
		}
	}
	// Copy children tree
	if len(a.Children) > 0 {
		dto.Children = make([]systemDto.SysAuthority, len(a.Children))
		for i, c := range a.Children {
			dto.Children[i] = s.toDTOWithTree(c)
		}
	}
	return dto
}

func (s *SysAuthority) GetAuthorityInfo(authorityId uint) (*systemDto.SysAuthority, error) {
	var auth model.SysAuthority
	err := s.Repo.DB().Preload("DataAuthorityId").Where("authority_id = ?", authorityId).First(&auth).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, s.ErrorMessage("角色不存在")
		}
		return nil, s.Error(err, "查询角色失败")
	}
	dto := s.toDTOWithTree(auth)
	return &dto, nil
}

func (s *SysAuthority) SetDataAuthority(adminAuthorityID uint, auth model.SysAuthority) error {
	var checkIDs []uint
	checkIDs = append(checkIDs, auth.AuthorityId)
	for i := range auth.DataAuthorityId {
		checkIDs = append(checkIDs, auth.DataAuthorityId[i].AuthorityId)
	}

	for i := range checkIDs {
		err := s.CheckAuthorityIDAuth(adminAuthorityID, checkIDs[i])
		if err != nil {
			return err
		}
	}

	var current model.SysAuthority
	if err := s.Repo.DB().Preload("DataAuthorityId").First(&current, "authority_id = ?", auth.AuthorityId).Error; err != nil {
		return s.ErrorMessage("角色不存在")
	}
	return s.Repo.DB().Model(&current).Association("DataAuthorityId").Replace(&auth.DataAuthorityId)
}

func (s *SysAuthority) SetMenuAuthority(adminAuthorityID, authorityId uint, menuIds []uint) error {
	menuIds = uniqueUintIDs(menuIds)
	if err := s.CheckAuthorityIDAuth(adminAuthorityID, authorityId); err != nil {
		return err
	}
	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("sys_authority_authority_id = ?", authorityId).Delete(&model.SysAuthorityMenu{}).Error; err != nil {
			return err
		}

		for _, menuId := range menuIds {
			if err := tx.Exec(
				"INSERT INTO sys_authority_menus (sys_base_menu_id, sys_authority_authority_id) VALUES (?, ?)",
				menuId, authorityId,
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *SysAuthority) GetUserIdsByAuthorityId(adminAuthorityID, authorityId uint) ([]uint, error) {
	if err := s.CheckAuthorityIDAuth(adminAuthorityID, authorityId); err != nil {
		return nil, err
	}
	return s.Repo.GetUserIdsByAuthorityId(authorityId)
}

func (s *SysAuthority) SetRoleUsers(adminAuthorityID, authorityId uint, userIds []uint) error {
	if err := s.CheckAuthorityIDAuth(adminAuthorityID, authorityId); err != nil {
		return err
	}
	userIds = uniqueUintIDs(userIds)
	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		var authorityCount int64
		if err := tx.Model(&model.SysAuthority{}).Where("authority_id = ?", authorityId).Count(&authorityCount).Error; err != nil {
			return err
		}
		if authorityCount != 1 {
			return s.ErrorMessage("角色不存在")
		}
		if len(userIds) > 0 {
			var userCount int64
			if err := tx.Model(&model.SysUser{}).Where("id IN ?", userIds).Count(&userCount).Error; err != nil {
				return err
			}
			if userCount != int64(len(userIds)) {
				return s.ErrorMessage("存在无效用户ID")
			}
		}
		// 1. Find current users of this role
		var existingRecords []model.SysUserAuthority
		if err := tx.Where("sys_authority_authority_id = ?", authorityId).Find(&existingRecords).Error; err != nil {
			return err
		}

		currentSet := make(map[uint]struct{})
		for _, r := range existingRecords {
			currentSet[r.SysUserId] = struct{}{}
		}

		targetSet := make(map[uint]struct{})
		for _, id := range userIds {
			targetSet[id] = struct{}{}
		}

		// 2. Delete all existing associations for this role
		if err := tx.Delete(&model.SysUserAuthority{}, "sys_authority_authority_id = ?", authorityId).Error; err != nil {
			return err
		}

		// 3. For removed users whose primary role is this one, reassign primary role
		for userId := range currentSet {
			if _, ok := targetSet[userId]; ok {
				continue
			}
			var user model.SysUser
			if err := tx.First(&user, "id = ?", userId).Error; err != nil {
				return err
			}
			if user.AuthorityId == authorityId {
				var another model.SysUserAuthority
				if err := tx.Where("sys_user_id = ?", userId).First(&another).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return s.ErrorMessage("用户至少需要保留一个角色")
					}
					return err
				}
				if err := tx.Model(&model.SysUser{}).Where("id = ?", userId).
					Update("authority_id", another.SysAuthorityAuthorityId).Error; err != nil {
					return err
				}
			}
		}

		// 4. Insert new records
		if len(userIds) > 0 {
			newRecords := make([]model.SysUserAuthority, 0, len(userIds))
			for _, uid := range userIds {
				newRecords = append(newRecords, model.SysUserAuthority{
					SysUserId:               uid,
					SysAuthorityAuthorityId: authorityId,
				})
			}
			if err := tx.Create(&newRecords).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func uniqueUintIDs(ids []uint) []uint {
	seen := make(map[uint]struct{}, len(ids))
	result := make([]uint, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func (s *SysAuthority) GetStructAuthorityList(authorityID uint) ([]uint, error) {
	var authority model.SysAuthority
	if err := s.Repo.DB().Where("authority_id = ?", authorityID).First(&authority).Error; err != nil {
		return nil, err
	}
	var children []model.SysAuthority
	if err := s.Repo.DB().Where("parent_id = ?", authorityID).Find(&children).Error; err != nil {
		return nil, err
	}
	result := make([]uint, 0, len(children)+1)
	for _, child := range children {
		result = append(result, child.AuthorityId)
		descendants, err := s.GetStructAuthorityList(child.AuthorityId)
		if err != nil {
			return nil, err
		}
		result = append(result, descendants...)
	}
	if authority.ParentId != nil && *authority.ParentId == 0 {
		result = append(result, authorityID)
	}
	return result, nil
}

func (s *SysAuthority) CheckAuthorityIDAuth(adminAuthorityID, targetID uint) error {
	if !config.GetConfig().System.UseStrictAuth {
		return nil
	}
	ids, err := s.GetStructAuthorityList(adminAuthorityID)
	if err != nil {
		return err
	}
	for _, id := range ids {
		if id == targetID {
			return nil
		}
	}
	return s.ErrorMessage("您提交的角色ID不合法")
}

func (s *SysAuthority) GetParentAuthorityID(authorityId uint) (*uint, error) {
	var auth model.SysAuthority
	err := s.Repo.DB().Where("authority_id = ?", authorityId).First(&auth).Error
	if err != nil || auth.AuthorityId == 0 {
		return nil, s.ErrorMessage("角色不存在")
	}
	return auth.ParentId, nil
}

func (s *SysAuthority) toDTO(a model.SysAuthority) systemDto.SysAuthority {
	return systemDto.SysAuthority{
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
		AuthorityId:   a.AuthorityId,
		AuthorityName: a.AuthorityName,
		ParentId:      a.ParentId,
		DefaultRouter: a.DefaultRouter,
		Children:      []systemDto.SysAuthority{},
	}
}
