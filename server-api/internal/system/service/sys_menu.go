package service

import (
	"megin/internal/base"
	systemDto "megin/internal/system/dto"
	"megin/internal/system/model"
	repo "megin/internal/system/repository"
	"megin/pkg/context/api"
	"megin/pkg/utils"
	"sort"
	"time"
)

func menuMetaFromRequest(meta systemDto.Meta, title, icon string) model.Meta {
	if meta.Title == "" {
		meta.Title = title
	}
	if meta.Icon == "" {
		meta.Icon = icon
	}
	return model.Meta{
		ActiveName:     meta.ActiveName,
		KeepAlive:      meta.KeepAlive,
		DefaultMenu:    meta.DefaultMenu,
		Title:          meta.Title,
		Icon:           meta.Icon,
		CloseTab:       meta.CloseTab,
		TransitionType: meta.TransitionType,
	}
}

type SysMenu struct {
	base.Service
	Repo        *repo.SysBaseMenu
	MenuBtnRepo *repo.SysMenuBtn
	AuthBtnRepo *repo.SysAuthorityBtn
}

func NewSysMenu(ctx *api.Context) *SysMenu {
	s := &SysMenu{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysBaseMenu(ctx)
	s.MenuBtnRepo = repo.NewSysMenuBtn(ctx)
	s.AuthBtnRepo = repo.NewSysAuthorityBtn(ctx)
	return s
}

func (s *SysMenu) GetMenuTree() ([]systemDto.SysMenu, error) {
	menus, err := s.Repo.GetMenuTree()
	if err != nil {
		return nil, s.Error(err, "查询菜单树失败")
	}
	return s.buildMenuTree(menus), nil
}

func (s *SysMenu) GetBaseMenuTree() ([]systemDto.SysBaseMenu, error) {
	menus, err := s.Repo.GetMenuTree()
	if err != nil {
		return nil, s.Error(err, "查询基础菜单树失败")
	}
	return s.toBaseMenuDTOs(menus), nil
}

func (s *SysMenu) AddBaseMenu(req *systemDto.AddBaseMenuReq) error {
	menu := model.SysBaseMenu{
		ParentId:  req.ParentId,
		Path:      req.Path,
		Name:      req.Name,
		Hidden:    req.Hidden,
		Component: req.Component,
		Sort:      req.Sort,
		Meta:      menuMetaFromRequest(req.Meta, req.Title, req.Icon),
		SystemModel: base.SystemModel{
			CreatedAt: utils.TimePtr(time.Now()),
			UpdatedAt: utils.TimePtr(time.Now()),
		},
	}
	return s.Repo.Create(&menu)
}

func (s *SysMenu) DeleteBaseMenu(id uint) error {
	menu, err := s.Repo.GetById(id)
	if err != nil || menu.ID == 0 {
		return s.ErrorMessage("菜单不存在")
	}
	return s.Repo.DeleteById(id)
}

func (s *SysMenu) UpdateBaseMenu(req *systemDto.UpdateBaseMenuReq) error {
	menu, err := s.Repo.GetById(req.ID)
	if err != nil || menu.ID == 0 {
		return s.ErrorMessage("菜单不存在")
	}

	menu.ParentId = req.ParentId
	menu.Path = req.Path
	menu.Name = req.Name
	menu.Hidden = req.Hidden
	menu.Component = req.Component
	menu.Sort = req.Sort
	menu.Meta = menuMetaFromRequest(req.Meta, req.Title, req.Icon)
	menu.UpdatedAt = utils.TimePtr(time.Now())
	return s.Repo.Save(&menu)
}

func (s *SysMenu) GetBaseMenuById(id uint) (*systemDto.SysBaseMenu, error) {
	menu, err := s.Repo.GetByIdWithDetails(id)
	if err != nil {
		return nil, err
	}
	if menu.ID == 0 {
		return nil, s.ErrorMessage("菜单不存在")
	}
	dto := s.toBaseMenuDTO(menu)
	return &dto, nil
}

func (s *SysMenu) AddMenuAuthority(adminAuthorityID, authorityId uint, menuIds []uint) error {
	return NewSysAuthority(s.Ctx).SetMenuAuthority(adminAuthorityID, authorityId, menuIds)
}

func (s *SysMenu) GetMenuAuthority(authorityId uint) ([]systemDto.SysBaseMenu, error) {
	menus, err := s.Repo.GetMenuByAuthority(authorityId)
	if err != nil {
		return nil, s.Error(err, "查询菜单权限失败")
	}
	return s.toBaseMenuDTOs(menus), nil
}

func (s *SysMenu) GetMenuInfoList() ([]systemDto.SysBaseMenu, error) {
	menus, err := s.Repo.GetMenuTree()
	if err != nil {
		return nil, s.Error(err, "查询菜单列表失败")
	}
	return s.toBaseMenuDTOs(menus), nil
}

// GetAuthoritiesByMenuId 获取拥有指定菜单的角色ID列表
func (s *SysMenu) GetAuthoritiesByMenuId(menuId uint) ([]uint, error) {
	authorityIds, err := s.Repo.GetAuthoritiesByMenuId(menuId)
	if err != nil {
		return nil, s.Error(err, "查询菜单角色失败")
	}
	if authorityIds == nil {
		authorityIds = []uint{}
	}
	return authorityIds, nil
}

// GetDefaultRouterAuthorityIds 获取将指定菜单设为首页的角色ID列表
func (s *SysMenu) GetDefaultRouterAuthorityIds(menuId uint) ([]uint, error) {
	authorityIds, err := s.Repo.GetDefaultRouterAuthorityIds(menuId)
	if err != nil {
		return nil, s.Error(err, "查询首页角色失败")
	}
	if authorityIds == nil {
		authorityIds = []uint{}
	}
	return authorityIds, nil
}

// SetMenuAuthorities 全量覆盖某菜单关联的角色列表
func (s *SysMenu) SetMenuAuthorities(menuId uint, authorityIds []uint) error {
	return s.Repo.SetMenuAuthorities(menuId, authorityIds)
}

func (s *SysMenu) GetMenuTreeByAuthority(authorityId uint) ([]systemDto.SysMenu, error) {
	menus, err := s.Repo.GetMenuByAuthority(authorityId)
	if err != nil {
		return nil, s.Error(err, "查询角色菜单失败")
	}

	roots := s.buildMenuTree(menus)

	// Load buttons for each menu
	for i := range roots {
		s.loadMenuBtns(&roots[i], authorityId)
	}

	return roots, nil
}

func (s *SysMenu) loadMenuBtns(menu *systemDto.SysMenu, authorityId uint) {
	btns, err := s.AuthBtnRepo.GetByAuthority(menu.MenuId, authorityId)
	if err == nil && len(btns) > 0 {
		menu.Btns = make(map[string]uint)
		for _, btn := range btns {
			menu.Btns[btn.SysBaseMenuBtn.Name] = authorityId
		}
	}
	for i := range menu.Children {
		s.loadMenuBtns(&menu.Children[i], authorityId)
	}
}

func (s *SysMenu) buildMenuTree(menus []model.SysBaseMenu) []systemDto.SysMenu {
	ordered := append([]model.SysBaseMenu(nil), menus...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Sort == ordered[j].Sort {
			return ordered[i].ID < ordered[j].ID
		}
		return ordered[i].Sort < ordered[j].Sort
	})

	items := make(map[uint]systemDto.SysMenu, len(ordered))
	children := make(map[uint][]uint, len(ordered))
	rootIDs := make([]uint, 0)
	for _, menu := range ordered {
		item := s.baseToSysMenu(menu)
		item.Children = []systemDto.SysMenu{}
		items[menu.ID] = item
	}
	for _, menu := range ordered {
		if menu.ParentId == 0 {
			rootIDs = append(rootIDs, menu.ID)
			continue
		}
		if _, parentExists := items[menu.ParentId]; !parentExists {
			// Keep an authorized orphan visible instead of silently dropping it.
			rootIDs = append(rootIDs, menu.ID)
			continue
		}
		children[menu.ParentId] = append(children[menu.ParentId], menu.ID)
	}

	visiting := make(map[uint]bool, len(items))
	var materialize func(uint) systemDto.SysMenu
	materialize = func(id uint) systemDto.SysMenu {
		item := items[id]
		if visiting[id] {
			return item
		}
		visiting[id] = true
		for _, childID := range children[id] {
			if visiting[childID] {
				continue
			}
			item.Children = append(item.Children, materialize(childID))
		}
		delete(visiting, id)
		return item
	}

	roots := make([]systemDto.SysMenu, 0, len(rootIDs))
	for _, id := range rootIDs {
		roots = append(roots, materialize(id))
	}
	return roots
}

func (s *SysMenu) toBaseMenuDTO(m model.SysBaseMenu) systemDto.SysBaseMenu {
	btns := make([]systemDto.SysBaseMenuBtn, len(m.MenuBtn))
	for i, btn := range m.MenuBtn {
		btns[i] = systemDto.SysBaseMenuBtn{ID: btn.ID, Name: btn.Name, Desc: btn.Desc}
	}
	params := make([]systemDto.SysBaseMenuParam, len(m.Parameters))
	for i, p := range m.Parameters {
		params[i] = systemDto.SysBaseMenuParam{
			SysBaseMenuID: p.SysBaseMenuID,
			Type:          p.Type,
			Key:           p.Key,
			Value:         p.Value,
		}
	}
	children := make([]systemDto.SysBaseMenu, len(m.Children))
	for i, c := range m.Children {
		children[i] = s.toBaseMenuDTO(c)
	}
	return systemDto.SysBaseMenu{
		ID:         m.ID,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
		MenuLevel:  m.MenuLevel,
		ParentId:   m.ParentId,
		Path:       m.Path,
		Name:       m.Name,
		Hidden:     m.Hidden,
		Component:  m.Component,
		Sort:       m.Sort,
		Meta:       systemDto.Meta{Title: m.Meta.Title, Icon: m.Meta.Icon},
		Children:   children,
		Parameters: params,
		MenuBtn:    btns,
	}
}

func (s *SysMenu) toBaseMenuDTOs(menus []model.SysBaseMenu) []systemDto.SysBaseMenu {
	dtos := make([]systemDto.SysBaseMenu, len(menus))
	for i, m := range menus {
		dtos[i] = s.toBaseMenuDTO(m)
	}
	return dtos
}

func (s *SysMenu) baseToSysMenu(m model.SysBaseMenu) systemDto.SysMenu {
	params := make([]systemDto.SysBaseMenuParam, len(m.Parameters))
	for i, p := range m.Parameters {
		params[i] = systemDto.SysBaseMenuParam{
			SysBaseMenuID: p.SysBaseMenuID,
			Type:          p.Type,
			Key:           p.Key,
			Value:         p.Value,
		}
	}
	btns := make([]systemDto.SysBaseMenuBtn, len(m.MenuBtn))
	for i, btn := range m.MenuBtn {
		btns[i] = systemDto.SysBaseMenuBtn{ID: btn.ID, Name: btn.Name, Desc: btn.Desc}
	}
	return systemDto.SysMenu{
		SysBaseMenu: systemDto.SysBaseMenu{
			ID:         m.ID,
			CreatedAt:  m.CreatedAt,
			UpdatedAt:  m.UpdatedAt,
			MenuLevel:  m.MenuLevel,
			ParentId:   m.ParentId,
			Path:       m.Path,
			Name:       m.Name,
			Hidden:     m.Hidden,
			Component:  m.Component,
			Sort:       m.Sort,
			Meta:       systemDto.Meta{Title: m.Meta.Title, Icon: m.Meta.Icon},
			Parameters: params,
			MenuBtn:    btns,
		},
		MenuId:     m.ID,
		Parameters: params,
		Btns:       make(map[string]uint),
	}
}
