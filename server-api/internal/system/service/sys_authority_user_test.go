package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"megin/internal/base"
	"megin/internal/config"
	systemDto "megin/internal/system/dto"
	"megin/internal/system/model"
	"megin/pkg/context/api"
	"megin/pkg/logger"
)

func newSystemTestContext(t *testing.T) (*api.Context, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err = db.AutoMigrate(
		&model.SysAuthority{},
		&model.SysBaseMenu{},
		&model.SysBaseMenuParameter{},
		&model.SysBaseMenuBtn{},
		&model.SysUser{},
		&model.SysUserAuthority{},
		&model.SysAuthorityBtn{},
		&model.JwtBlacklist{},
		&CasbinRule{},
	); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	return &api.Context{Tx: db, Log: logger.New()}, db
}

func uintPtr(value uint) *uint { return &value }

func TestCreateAuthorityUsesLeastPrivilegeDefaultsAndHardDelete(t *testing.T) {
	ctx, db := newSystemTestContext(t)
	service := NewSysAuthority(ctx)
	config.GetConfig().System.UseStrictAuth = false
	t.Cleanup(func() { config.GetConfig().System.UseStrictAuth = false })

	if err := db.Create(&model.SysBaseMenu{SystemModel: modelSystemModel(1), Name: "dashboard"}).Error; err != nil {
		t.Fatalf("create default menu: %v", err)
	}
	if err := db.Create(&CasbinRule{PType: "p", V0: "888", V1: "/admin-only", V2: "POST"}).Error; err != nil {
		t.Fatalf("create admin policy: %v", err)
	}

	req := &systemDto.CreateAuthorityReq{
		AuthorityId: 100, AuthorityName: "operator", ParentId: uintPtr(0), DefaultRouter: "dashboard",
	}
	if _, err := service.CreateAuthority(888, req); err != nil {
		t.Fatalf("create authority: %v", err)
	}

	var menuIDs []uint
	if err := db.Table("sys_authority_menus").Where("sys_authority_authority_id = ?", 100).
		Pluck("sys_base_menu_id", &menuIDs).Error; err != nil {
		t.Fatalf("query authority menus: %v", err)
	}
	if len(menuIDs) != 1 || menuIDs[0] != 1 {
		t.Fatalf("default menus = %v, want [1]", menuIDs)
	}

	var policies []CasbinRule
	if err := db.Where("ptype = 'p' AND v0 = ?", "100").Find(&policies).Error; err != nil {
		t.Fatalf("query policies: %v", err)
	}
	if len(policies) != len(systemDto.DefaultCasbinInfos) {
		t.Fatalf("default policy count = %d, want %d", len(policies), len(systemDto.DefaultCasbinInfos))
	}
	for _, policy := range policies {
		if policy.V1 == "/admin-only" {
			t.Fatal("new authority inherited administrator policy")
		}
		if strings.HasPrefix(policy.V1, "/admin-api/") {
			t.Fatalf("default casbin policy path should not contain admin-api prefix: %s", policy.V1)
		}
	}

	if err := service.DeleteAuthority(888, 100); err != nil {
		t.Fatalf("delete authority: %v", err)
	}
	var count int64
	if err := db.Unscoped().Model(&model.SysAuthority{}).Where("authority_id = ?", 100).Count(&count).Error; err != nil {
		t.Fatalf("count deleted authority: %v", err)
	}
	if count != 0 {
		t.Fatalf("authority was soft-deleted; unscoped count = %d", count)
	}
	if _, err := service.CreateAuthority(888, req); err != nil {
		t.Fatalf("recreate authority with same ID: %v", err)
	}
}

func TestAuthorityTreeHonorsStrictAuthConfiguration(t *testing.T) {
	ctx, db := newSystemTestContext(t)
	service := NewSysAuthority(ctx)
	authorities := []model.SysAuthority{
		{AuthorityId: 888, AuthorityName: "root", ParentId: uintPtr(0)},
		{AuthorityId: 100, AuthorityName: "child", ParentId: uintPtr(888)},
		{AuthorityId: 200, AuthorityName: "grandchild", ParentId: uintPtr(100)},
		{AuthorityId: 999, AuthorityName: "other root", ParentId: uintPtr(0)},
	}
	if err := db.Create(&authorities).Error; err != nil {
		t.Fatalf("create authorities: %v", err)
	}
	t.Cleanup(func() { config.GetConfig().System.UseStrictAuth = false })

	config.GetConfig().System.UseStrictAuth = false
	list, err := service.GetAuthorityInfoList(888)
	if err != nil {
		t.Fatalf("non-strict list: %v", err)
	}
	if len(list) != 2 || len(list[0].Children) != 1 || len(list[0].Children[0].Children) != 1 {
		t.Fatalf("unexpected non-strict tree: %#v", list)
	}

	config.GetConfig().System.UseStrictAuth = true
	list, err = service.GetAuthorityInfoList(100)
	if err != nil {
		t.Fatalf("strict list: %v", err)
	}
	if len(list) != 1 || list[0].AuthorityId != 200 {
		t.Fatalf("strict child list = %#v, want authority 200", list)
	}
	if err := service.CheckAuthorityIDAuth(100, 999); err == nil {
		t.Fatal("strict authorization accepted an out-of-scope authority")
	}
}

func TestSetUserAuthorityRejectsUnmatchedDefaultRouter(t *testing.T) {
	ctx, db := newSystemTestContext(t)
	service := NewSysUser(ctx)
	if err := db.Create(&model.SysAuthority{
		AuthorityId: 100, AuthorityName: "operator", ParentId: uintPtr(0), DefaultRouter: "dashboard",
	}).Error; err != nil {
		t.Fatalf("create authority: %v", err)
	}
	menu := model.SysBaseMenu{SystemModel: modelSystemModel(1), Name: "settings"}
	if err := db.Create(&menu).Error; err != nil {
		t.Fatalf("create menu: %v", err)
	}
	if err := db.Exec(
		"INSERT INTO sys_authority_menus (sys_base_menu_id, sys_authority_authority_id) VALUES (?, ?)", 1, 100,
	).Error; err != nil {
		t.Fatalf("assign menu: %v", err)
	}
	user := model.SysUser{Username: "tester", AuthorityId: 100, Enable: 1}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&model.SysUserAuthority{SysUserId: user.ID, SysAuthorityAuthorityId: 100}).Error; err != nil {
		t.Fatalf("assign authority: %v", err)
	}
	if err := service.SetUserAuthority(user.ID, 100); err == nil {
		t.Fatal("role switch succeeded without a menu matching default_router")
	}
}

func TestGetUserAuthoritiesJoinAndSelfSetting(t *testing.T) {
	ctx, db := newSystemTestContext(t)
	service := NewSysUser(ctx)
	if err := db.Create(&model.SysAuthority{AuthorityId: 100, AuthorityName: "operator", ParentId: uintPtr(0)}).Error; err != nil {
		t.Fatalf("create authority: %v", err)
	}
	user := model.SysUser{Username: "tester", AuthorityId: 100, Enable: 1}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&model.SysUserAuthority{SysUserId: user.ID, SysAuthorityAuthorityId: 100}).Error; err != nil {
		t.Fatalf("assign authority: %v", err)
	}
	authorities, err := service.GetUserAuthorities(user.ID)
	if err != nil {
		t.Fatalf("get user authorities: %v", err)
	}
	if len(authorities) != 1 || authorities[0].AuthorityId != 100 {
		t.Fatalf("authorities = %#v", authorities)
	}

	var request systemDto.SetSelfSettingReq
	if err := json.Unmarshal([]byte(`{"theme":"dark","compact":true}`), &request); err != nil {
		t.Fatalf("decode root setting object: %v", err)
	}
	if err := service.SetSelfSetting(user.ID, map[string]any(request)); err != nil {
		t.Fatalf("set self setting: %v", err)
	}
	var updated model.SysUser
	if err := db.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if updated.OriginSetting["theme"] != "dark" {
		t.Fatalf("origin setting = %#v", updated.OriginSetting)
	}
}

func TestJwtBlacklistQueryTreatsMissingRecordAsNotBlacklisted(t *testing.T) {
	ctx, db := newSystemTestContext(t)
	service := NewSysJwt(ctx)

	blacklisted, err := service.IsBlacklisted("missing-token")
	if err != nil {
		t.Fatalf("IsBlacklisted missing token error = %v", err)
	}
	if blacklisted {
		t.Fatal("missing token should not be blacklisted")
	}

	entry := model.JwtBlacklist{Jwt: "blocked-token"}
	if err := db.Create(&entry).Error; err != nil {
		t.Fatalf("create blacklist token: %v", err)
	}

	blacklisted, err = service.IsBlacklisted("blocked-token")
	if err != nil {
		t.Fatalf("IsBlacklisted blocked token error = %v", err)
	}
	if !blacklisted {
		t.Fatal("blocked token should be blacklisted")
	}
}

func TestSetRoleUsersKeepsPrimaryRoleValidAndRollsBack(t *testing.T) {
	ctx, db := newSystemTestContext(t)
	service := NewSysAuthority(ctx)
	config.GetConfig().System.UseStrictAuth = false
	t.Cleanup(func() { config.GetConfig().System.UseStrictAuth = false })
	if err := db.Create(&[]model.SysAuthority{
		{AuthorityId: 100, AuthorityName: "primary", ParentId: uintPtr(0)},
		{AuthorityId: 200, AuthorityName: "fallback", ParentId: uintPtr(0)},
	}).Error; err != nil {
		t.Fatalf("create authorities: %v", err)
	}
	user := model.SysUser{Username: "tester", AuthorityId: 100, Enable: 1}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&model.SysUserAuthority{SysUserId: user.ID, SysAuthorityAuthorityId: 100}).Error; err != nil {
		t.Fatalf("assign primary authority: %v", err)
	}

	if err := service.SetRoleUsers(888, 100, []uint{}); err == nil {
		t.Fatal("removed a user's only role")
	}
	var associationCount int64
	if err := db.Model(&model.SysUserAuthority{}).
		Where("sys_user_id = ? AND sys_authority_authority_id = ?", user.ID, 100).
		Count(&associationCount).Error; err != nil {
		t.Fatalf("count rolled-back association: %v", err)
	}
	if associationCount != 1 {
		t.Fatalf("association count after rollback = %d, want 1", associationCount)
	}

	if err := db.Create(&model.SysUserAuthority{SysUserId: user.ID, SysAuthorityAuthorityId: 200}).Error; err != nil {
		t.Fatalf("assign fallback authority: %v", err)
	}
	if err := service.SetRoleUsers(888, 100, []uint{}); err != nil {
		t.Fatalf("remove role with fallback available: %v", err)
	}
	var updated model.SysUser
	if err := db.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if updated.AuthorityId != 200 {
		t.Fatalf("primary authority = %d, want 200", updated.AuthorityId)
	}
}

func TestGetMenuTreeByAuthorityKeepsLateChildrenAndNestedData(t *testing.T) {
	ctx, db := newSystemTestContext(t)
	service := NewSysMenu(ctx)
	menus := []model.SysBaseMenu{
		{SystemModel: modelSystemModel(13), ParentId: 3, Name: "user", Sort: 4},
		{SystemModel: modelSystemModel(3), ParentId: 0, Name: "superAdmin", Sort: 3},
		{SystemModel: modelSystemModel(10), ParentId: 3, Name: "authority", Sort: 1},
		{SystemModel: modelSystemModel(11), ParentId: 3, Name: "menu", Sort: 2},
		{SystemModel: modelSystemModel(50), ParentId: 13, Name: "userChild", Sort: 1},
		{SystemModel: modelSystemModel(90), ParentId: 999, Name: "orphan", Sort: 2},
	}
	if err := db.Create(&menus).Error; err != nil {
		t.Fatalf("create menus: %v", err)
	}
	for _, menu := range menus {
		if err := db.Exec(
			"INSERT INTO sys_authority_menus (sys_base_menu_id, sys_authority_authority_id) VALUES (?, ?)",
			menu.ID, 888,
		).Error; err != nil {
			t.Fatalf("assign menu %d: %v", menu.ID, err)
		}
	}
	if err := db.Create(&model.SysBaseMenuParameter{
		SysBaseMenuID: 13, Type: "query", Key: "tab", Value: "users",
	}).Error; err != nil {
		t.Fatalf("create menu parameter: %v", err)
	}
	button := model.SysBaseMenuBtn{SystemModel: modelSystemModel(1), SysBaseMenuID: 13, Name: "create"}
	if err := db.Create(&button).Error; err != nil {
		t.Fatalf("create menu button: %v", err)
	}
	if err := db.Create(&model.SysAuthorityBtn{
		AuthorityId: 888, SysMenuID: 13, SysBaseMenuBtnID: button.ID,
	}).Error; err != nil {
		t.Fatalf("assign menu button: %v", err)
	}

	tree, err := service.GetMenuTreeByAuthority(888)
	if err != nil {
		t.Fatalf("get authority menu tree: %v", err)
	}
	if len(tree) != 2 || tree[0].Name != "orphan" || tree[1].Name != "superAdmin" {
		t.Fatalf("root menus = %#v", tree)
	}
	children := tree[1].Children
	if len(children) != 3 || children[0].Name != "authority" || children[1].Name != "menu" || children[2].Name != "user" {
		t.Fatalf("superAdmin children = %#v", children)
	}
	userMenu := children[2]
	if len(userMenu.Children) != 1 || userMenu.Children[0].Name != "userChild" {
		t.Fatalf("user children = %#v", userMenu.Children)
	}
	if len(userMenu.Parameters) != 1 || userMenu.Parameters[0].Key != "tab" {
		t.Fatalf("user parameters = %#v", userMenu.Parameters)
	}
	if userMenu.Btns["create"] != 888 {
		t.Fatalf("user buttons = %#v", userMenu.Btns)
	}
}

func modelSystemModel(id uint) base.SystemModel {
	return base.SystemModel{ID: id}
}
