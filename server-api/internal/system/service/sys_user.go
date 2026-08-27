package service

import (
	"errors"
	"megin/internal/base"
	"megin/internal/config"
	commonDto "megin/internal/module/common/dto"
	systemDto "megin/internal/system/dto"
	"megin/internal/system/model"
	repo "megin/internal/system/repository"
	"megin/pkg/captcha"
	"megin/pkg/context/api"
	"megin/pkg/utils"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SysUser struct {
	base.Service
	Repo      *repo.SysUser
	Blacklist *SysJwt
}

func NewSysUser(ctx *api.Context) *SysUser {
	s := &SysUser{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysUser(ctx)
	s.Blacklist = NewSysJwt(ctx)
	return s
}

func (s *SysUser) Register(adminAuthorityID uint, req *systemDto.RegisterReq) (*model.SysUser, error) {
	existing, err := s.Repo.GetByUsername(req.Username)
	if err != nil {
		return nil, s.Error(err, "查询用户失败")
	}
	if existing != nil {
		return nil, s.ErrorMessage("用户名已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, s.Error(err, "密码加密失败")
	}

	authorityId := req.AuthorityId
	if authorityId == 0 {
		authorityId = 888
	}
	authorityIDs := uniqueAuthorityIDs(req.AuthorityIds)
	if len(authorityIDs) == 0 {
		authorityIDs = []uint{authorityId}
	}
	if !containsAuthorityID(authorityIDs, authorityId) {
		return nil, s.ErrorMessage("主角色必须包含在用户角色列表中")
	}
	authorityService := NewSysAuthority(s.Ctx)
	for _, id := range authorityIDs {
		if err := authorityService.CheckAuthorityIDAuth(adminAuthorityID, id); err != nil {
			return nil, err
		}
	}

	now := time.Now()
	user := model.SysUser{
		UUID:        uuid.New().String(),
		Username:    req.Username,
		Password:    string(hashedPassword),
		NickName:    req.NickName,
		HeaderImg:   req.HeaderImg,
		AuthorityId: authorityId,
		Phone:       req.Phone,
		Email:       req.Email,
		Enable:      req.Enable,
		SystemModel: base.SystemModel{
			CreatedAt: utils.TimePtr(now),
			UpdatedAt: utils.TimePtr(now),
		},
	}
	if user.Enable == 0 {
		user.Enable = 1
	}

	err = s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		var authorityCount int64
		if err := tx.Model(&model.SysAuthority{}).Where("authority_id IN ?", authorityIDs).Count(&authorityCount).Error; err != nil {
			return err
		}
		if authorityCount != int64(len(authorityIDs)) {
			return s.ErrorMessage("存在无效角色ID")
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		associations := make([]model.SysUserAuthority, 0, len(authorityIDs))
		for _, id := range authorityIDs {
			associations = append(associations, model.SysUserAuthority{
				SysUserId: user.ID, SysAuthorityAuthorityId: id,
			})
		}
		return tx.Create(&associations).Error
	})
	if err != nil {
		return nil, s.Error(err, "创建用户失败")
	}

	return &user, nil
}

func (s *SysUser) Login(req *systemDto.LoginReq) (*systemDto.SysUser, string, int64, error) {
	user, err := s.Repo.GetByUsername(req.Username)
	if err != nil {
		return nil, "", 0, s.ErrorMessage("用户不存在")
	}
	if user == nil {
		return nil, "", 0, s.ErrorMessage("用户名或密码错误")
	}

	if user.Enable != 1 {
		return nil, "", 0, s.ErrorMessage("用户已被冻结")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", 0, s.ErrorMessage("用户名或密码错误")
	}

	// 图形验证码关闭时不读取也不校验请求中的验证码字段，便于本地开发调试。
	// 生产环境应通过 captcha.enable=true 开启校验，不能只依赖前端隐藏或展示验证码。
	if config.GetConfig().Admin.Captcha.Enable && !captcha.Verify(req.CaptchaId, req.Captcha) {
		return nil, "", 0, s.ErrorMessage("验证码错误")
	}

	if config.GetConfig().Admin.TOTP.Enable && user.TOTPEnabled {
		secret, err := decryptTOTPSecret(user.TOTPSecret)
		if err != nil {
			return nil, "", 0, s.Error(err, "读取Google TOTP配置失败")
		}
		if !verifyTOTPCode(secret, req.OTP, time.Now()) {
			return nil, "", 0, s.ErrorMessage("Google验证码错误")
		}
	}

	conf := config.GetConfig()
	expireDuration := time.Duration(conf.Jwt.ExpireSeconds) * time.Second
	if expireDuration <= 0 {
		expireDuration = 24 * time.Hour
	}
	expiresAtTime := time.Now().Add(expireDuration)

	claims := &commonDto.Claims{
		UserID:   int(user.ID),
		Username: user.Username,
		RoleId:   int(user.AuthorityId),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAtTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	jwtSecret := []byte(conf.Jwt.Secret)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		return nil, "", 0, s.Error(err, "生成token失败")
	}

	expiresAt := expiresAtTime.UnixMilli()
	userDTO := s.toDTO(*user)
	return &userDTO, tokenStr, expiresAt, nil
}

func (s *SysUser) GetTOTPStatus(userID uint) (*systemDto.TotpStatusResponse, error) {
	user, err := s.Repo.GetById(userID)
	if err != nil || user.ID == 0 {
		return nil, s.ErrorMessage("用户不存在")
	}
	conf := config.GetConfig()
	issuer := strings.TrimSpace(conf.Admin.TOTP.Issuer)
	if issuer == "" {
		issuer = "xadmin"
	}
	return &systemDto.TotpStatusResponse{
		Enabled:       user.TOTPEnabled,
		BoundAt:       user.TOTPBoundAt,
		Issuer:        issuer,
		Account:       user.Username,
		NeedSetup:     conf.Admin.TOTP.Enable && !user.TOTPEnabled,
		SystemEnabled: conf.Admin.TOTP.Enable,
	}, nil
}

func (s *SysUser) InitTOTP(userID uint) (*systemDto.TotpSetupResponse, error) {
	if !config.GetConfig().Admin.TOTP.Enable {
		return nil, s.ErrorMessage("系统未开启Google TOTP")
	}
	user, err := s.Repo.GetById(userID)
	if err != nil || user.ID == 0 {
		return nil, s.ErrorMessage("用户不存在")
	}

	secret, err := generateTOTPSecret()
	if err != nil {
		return nil, s.Error(err, "生成Google TOTP密钥失败")
	}
	encrypted, err := encryptTOTPSecret(secret)
	if err != nil {
		return nil, s.Error(err, "保存Google TOTP密钥失败")
	}

	if err := s.Repo.DB().Model(&model.SysUser{}).Where("id = ?", userID).Updates(map[string]any{
		"totp_secret":   encrypted,
		"totp_enabled":  false,
		"totp_bound_at": nil,
		"updated_at":    utils.TimePtr(time.Now()),
	}).Error; err != nil {
		return nil, s.Error(err, "初始化Google TOTP失败")
	}

	issuer := strings.TrimSpace(config.GetConfig().Admin.TOTP.Issuer)
	if issuer == "" {
		issuer = "xadmin"
	}
	otpAuthURL := buildTOTPAuthURL(issuer, user.Username, secret)
	return &systemDto.TotpSetupResponse{
		Secret:     secret,
		OTPAuthURL: otpAuthURL,
		QRContent:  otpAuthURL,
	}, nil
}

func (s *SysUser) EnableTOTP(userID uint, req *systemDto.TotpCodeReq) error {
	if !config.GetConfig().Admin.TOTP.Enable {
		return s.ErrorMessage("系统未开启Google TOTP")
	}
	user, err := s.Repo.GetById(userID)
	if err != nil || user.ID == 0 {
		return s.ErrorMessage("用户不存在")
	}
	if strings.TrimSpace(user.TOTPSecret) == "" {
		return s.ErrorMessage("请先初始化Google TOTP")
	}

	secret, err := decryptTOTPSecret(user.TOTPSecret)
	if err != nil {
		return s.Error(err, "读取Google TOTP配置失败")
	}
	if !verifyTOTPCode(secret, req.OTP, time.Now()) {
		return s.ErrorMessage("Google验证码错误")
	}

	now := time.Now()
	return s.Repo.DB().Model(&model.SysUser{}).Where("id = ?", userID).Updates(map[string]any{
		"totp_enabled":  true,
		"totp_bound_at": utils.TimePtr(now),
		"updated_at":    utils.TimePtr(now),
	}).Error
}

func (s *SysUser) DisableTOTP(userID uint, req *systemDto.TotpCodeReq) error {
	user, err := s.Repo.GetById(userID)
	if err != nil || user.ID == 0 {
		return s.ErrorMessage("用户不存在")
	}
	if !user.TOTPEnabled {
		return s.ErrorMessage("Google TOTP未启用")
	}

	secret, err := decryptTOTPSecret(user.TOTPSecret)
	if err != nil {
		return s.Error(err, "读取Google TOTP配置失败")
	}
	if !verifyTOTPCode(secret, req.OTP, time.Now()) {
		return s.ErrorMessage("Google验证码错误")
	}

	now := time.Now()
	return s.Repo.DB().Model(&model.SysUser{}).Where("id = ?", userID).Updates(map[string]any{
		"totp_secret":   "",
		"totp_enabled":  false,
		"totp_bound_at": nil,
		"updated_at":    utils.TimePtr(now),
	}).Error
}

func (s *SysUser) ChangePassword(userID uint, req *systemDto.ChangePasswordReq) error {
	user, err := s.Repo.GetById(userID)
	if err != nil || user.ID == 0 {
		return s.ErrorMessage("用户不存在")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return s.ErrorMessage("原密码错误")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return s.Error(err, "密码加密失败")
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = utils.TimePtr(time.Now())
	return s.Repo.Save(&user)
}

func (s *SysUser) GetUserInfoList(req *systemDto.GetUserListReq) (*systemDto.PageResult[systemDto.SysUser], error) {
	db := s.Repo.DB().Model(&model.SysUser{})

	if req.Username != "" {
		db = db.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.NickName != "" {
		db = db.Where("nick_name LIKE ?", "%"+req.NickName+"%")
	}
	if req.Phone != "" {
		db = db.Where("phone LIKE ?", "%"+req.Phone+"%")
	}
	if req.Email != "" {
		db = db.Where("email LIKE ?", "%"+req.Email+"%")
	}

	// Order key whitelist (original GVA behavior)
	orderStr := "id desc"
	if req.OrderKey != "" {
		allowedOrders := map[string]bool{
			"id":        true,
			"username":  true,
			"nick_name": true,
			"phone":     true,
			"email":     true,
		}
		if allowedOrders[req.OrderKey] {
			orderStr = req.OrderKey
			if req.Desc {
				orderStr = req.OrderKey + " desc"
			}
		}
	}

	users, total, err := s.Repo.GetUserInfoList(db, req.PageNo, req.PageSize, orderStr)
	if err != nil {
		return nil, s.Error(err, "查询用户列表失败")
	}

	userDtos := make([]systemDto.SysUser, len(users))
	for i, u := range users {
		userDtos[i] = s.toDTO(u)
	}

	return &systemDto.PageResult[systemDto.SysUser]{
		PageNo:    req.PageNo,
		PageSize:  req.PageSize,
		TotalSize: total,
		List:      userDtos,
	}, nil
}

func (s *SysUser) SetUserAuthority(userID uint, authorityId uint) error {
	user, err := s.Repo.GetById(userID)
	if err != nil || user.ID == 0 {
		return s.ErrorMessage("用户不存在")
	}

	// Validate user has this role assigned (original GVA behavior)
	var userAuth model.SysUserAuthority
	if err := s.Repo.DB().Where("sys_user_id = ? AND sys_authority_authority_id = ?", userID, authorityId).
		First(&userAuth).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return s.ErrorMessage("该用户无此角色")
	} else if err != nil {
		return s.Error(err, "查询用户角色失败")
	}

	// Validate target authority exists and has valid default router
	var authority model.SysAuthority
	if err := s.Repo.DB().Where("authority_id = ?", authorityId).First(&authority).Error; err != nil {
		return s.ErrorMessage("角色不存在")
	}

	// Check if the authority's default router is among its assigned menus.
	var menuCount int64
	if err := s.Repo.DB().Table("sys_authority_menus").
		Joins("JOIN sys_base_menus ON sys_base_menus.id = sys_authority_menus.sys_base_menu_id").
		Where("sys_authority_menus.sys_authority_authority_id = ? AND sys_base_menus.name = ?", authorityId, authority.DefaultRouter).
		Count(&menuCount).Error; err != nil {
		return s.Error(err, "查询角色菜单失败")
	}
	if menuCount == 0 {
		return s.ErrorMessage("找不到默认路由,无法切换本角色")
	}

	return s.Repo.DB().Model(&model.SysUser{}).Where("id = ?", userID).Updates(map[string]any{
		"authority_id": authorityId,
		"updated_at":   utils.TimePtr(time.Now()),
	}).Error
}

func (s *SysUser) SetUserAuthorities(adminAuthorityID, id uint, authorityIds []uint) error {
	authorityIds = uniqueAuthorityIDs(authorityIds)
	if len(authorityIds) == 0 {
		return s.ErrorMessage("用户至少需要一个角色")
	}
	user, err := s.Repo.GetById(id)
	if err != nil || user.ID == 0 {
		return s.ErrorMessage("用户不存在")
	}

	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		authorityService := NewSysAuthority(s.Ctx)
		for _, authorityID := range authorityIds {
			if err := authorityService.CheckAuthorityIDAuth(adminAuthorityID, authorityID); err != nil {
				return err
			}
		}
		var authorityCount int64
		if err := tx.Model(&model.SysAuthority{}).Where("authority_id IN ?", authorityIds).Count(&authorityCount).Error; err != nil {
			return err
		}
		if authorityCount != int64(len(authorityIds)) {
			return s.ErrorMessage("存在无效角色ID")
		}
		// Delete all existing role associations
		if err := tx.Where("sys_user_id = ?", id).Delete(&model.SysUserAuthority{}).Error; err != nil {
			return err
		}

		// Insert new role associations
		var uas []model.SysUserAuthority
		for _, aid := range authorityIds {
			uas = append(uas, model.SysUserAuthority{
				SysUserId:               id,
				SysAuthorityAuthorityId: aid,
			})
		}
		if err := tx.Create(&uas).Error; err != nil {
			return err
		}

		// Update primary role to the first authority ID (original GVA behavior)
		return tx.Model(&model.SysUser{}).Where("id = ?", id).Update("authority_id", authorityIds[0]).Error
	})
}

func (s *SysUser) DeleteUser(id, currentUserId, currentAuthorityID uint) error {
	if id == currentUserId {
		return s.ErrorMessage("不能删除自身")
	}

	user, err := s.Repo.GetById(id)
	if err != nil {
		return err
	}
	if user.ID == 0 {
		return s.ErrorMessage("用户不存在")
	}
	if err := NewSysAuthority(s.Ctx).CheckAuthorityIDAuth(currentAuthorityID, user.AuthorityId); err != nil {
		return err
	}

	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		// Delete user
		if err := tx.Where("id = ?", id).Delete(&model.SysUser{}).Error; err != nil {
			return err
		}
		// Clean up user-role associations
		if err := tx.Where("sys_user_id = ?", id).Delete(&model.SysUserAuthority{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *SysUser) SetUserInfo(req *systemDto.ChangeUserInfoReq) error {
	user, err := s.Repo.GetById(req.ID)
	if err != nil || user.ID == 0 {
		return s.ErrorMessage("用户不存在")
	}

	if req.NickName != "" {
		user.NickName = req.NickName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.HeaderImg != "" {
		user.HeaderImg = req.HeaderImg
	}
	if req.Enable != 0 {
		user.Enable = req.Enable
	}
	user.UpdatedAt = utils.TimePtr(time.Now())

	return s.Repo.Save(&user)
}

func (s *SysUser) SetSelfInfo(userID uint, req *systemDto.ChangeUserInfoReq) error {
	req.ID = userID
	return s.SetUserInfo(req)
}

func (s *SysUser) SetSelfSetting(uid uint, setting map[string]any) error {
	return s.Repo.DB().Model(&model.SysUser{}).Where("id = ?", uid).
		Update("origin_setting", model.JSONMap(setting)).Error
}

func uniqueAuthorityIDs(ids []uint) []uint {
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

func containsAuthorityID(ids []uint, target uint) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func (s *SysUser) GetUserInfo(id uint) (*systemDto.SysUserResponse, error) {
	user, err := s.Repo.GetById(id)
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, s.ErrorMessage("用户不存在")
	}
	watermark := ""
	if config.GetConfig().Admin.Watermark {
		// 登录账号是可唯一定位操作人的标识；同时展示昵称，方便截图流转时人工识别。
		watermark = user.Username
		if nickname := strings.TrimSpace(user.NickName); nickname != "" {
			watermark = nickname + "（" + user.Username + "）"
		}
	}
	return &systemDto.SysUserResponse{
		UserInfo:  s.toDTOPtr(user),
		Watermark: watermark,
	}, nil
}

func (s *SysUser) FindUserById(id uint) (*systemDto.SysUser, error) {
	user, err := s.Repo.GetById(id)
	if err != nil || user.ID == 0 {
		return nil, s.ErrorMessage("用户不存在")
	}
	dtoUser := s.toDTO(user)
	return &dtoUser, nil
}

func (s *SysUser) ResetPassword(adminAuthorityID, id uint, password string) error {
	user, err := s.Repo.GetById(id)
	if err != nil || user.ID == 0 {
		return s.ErrorMessage("用户不存在")
	}
	if err := NewSysAuthority(s.Ctx).CheckAuthorityIDAuth(adminAuthorityID, user.AuthorityId); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return s.Error(err, "密码加密失败")
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = utils.TimePtr(time.Now())
	return s.Repo.Save(&user)
}

func (s *SysUser) GetUserAuthorities(userID uint) ([]systemDto.SysAuthority, error) {
	authorities, err := s.Repo.GetUserAuthorities(userID)
	if err != nil {
		return nil, err
	}
	dtos := make([]systemDto.SysAuthority, len(authorities))
	for i, a := range authorities {
		dtos[i] = s.authorityToDTO(a)
	}
	return dtos, nil
}

func (s *SysUser) toDTO(u model.SysUser) systemDto.SysUser {
	authorities := make([]systemDto.SysAuthority, len(u.Authorities))
	for i, authority := range u.Authorities {
		authorities[i] = s.authorityToDTO(authority)
	}

	return systemDto.SysUser{
		ID:            u.ID,
		UUID:          u.UUID,
		Username:      u.Username,
		NickName:      u.NickName,
		HeaderImg:     u.HeaderImg,
		AuthorityId:   u.AuthorityId,
		Phone:         u.Phone,
		Email:         u.Email,
		Enable:        u.Enable,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		Authority:     s.authorityToDTO(u.Authority),
		Authorities:   authorities,
		OriginSetting: map[string]any(u.OriginSetting),
	}
}

func (s *SysUser) toDTOPtr(u model.SysUser) *systemDto.SysUser {
	dtoUser := s.toDTO(u)
	return &dtoUser
}

func (s *SysUser) authorityToDTO(a model.SysAuthority) systemDto.SysAuthority {
	return systemDto.SysAuthority{
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
		AuthorityId:   a.AuthorityId,
		AuthorityName: a.AuthorityName,
		ParentId:      a.ParentId,
		DefaultRouter: a.DefaultRouter,
	}
}

// Ensure model.Model interface requirement
var _ base.Model = (*model.SysUser)(nil)
