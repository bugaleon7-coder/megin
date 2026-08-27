package service

import (
	"megin/internal/base"
	systemDto "megin/internal/system/dto"
	repo "megin/internal/system/repository"
	"megin/pkg/context/api"
)

type SysAuthorityBtn struct {
	base.Service
	Repo       *repo.SysAuthorityBtn
	MenuBtnRepo *repo.SysMenuBtn
}

func NewSysAuthorityBtn(ctx *api.Context) *SysAuthorityBtn {
	s := &SysAuthorityBtn{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysAuthorityBtn(ctx)
	s.MenuBtnRepo = repo.NewSysMenuBtn(ctx)
	return s
}

func (s *SysAuthorityBtn) GetAuthorityBtn(authorityId, menuID uint) (*systemDto.AuthorityBtnResponse, error) {
	btns, err := s.Repo.GetByAuthority(menuID, authorityId)
	if err != nil {
		return nil, s.Error(err, "查询按钮权限失败")
	}

	selected := make([]uint, len(btns))
	for i, btn := range btns {
		selected[i] = btn.SysBaseMenuBtnID
	}

	return &systemDto.AuthorityBtnResponse{Selected: selected}, nil
}

func (s *SysAuthorityBtn) SetAuthorityBtn(req *systemDto.SysAuthorityBtnReq) error {
	// Delete old buttons for this menu and authority
	if err := s.Repo.DeleteByAuthority(req.SysMenuID, req.AuthorityId); err != nil {
		return s.Error(err, "清除旧按钮权限失败")
	}

	// Insert new button authorizations
	for _, btnID := range req.Selected {
		authBtn := struct {
			AuthorityId      uint
			SysMenuID        uint
			SysBaseMenuBtnID uint
		}{
			AuthorityId:      req.AuthorityId,
			SysMenuID:        req.SysMenuID,
			SysBaseMenuBtnID: btnID,
		}
		if err := s.Repo.DB().Table("sys_authority_btns").Create(&authBtn).Error; err != nil {
			return s.Error(err, "设置按钮权限失败")
		}
	}

	return nil
}

func (s *SysAuthorityBtn) CanRemoveAuthorityBtn(menuID, btnID, authorityId uint) (bool, error) {
	return s.Repo.CanRemoveBtn(menuID, btnID, authorityId)
}