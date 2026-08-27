package service

import (
	"megin/internal/base"
	"megin/internal/config"
	commonDto "megin/internal/module/common/dto"
	systemDto "megin/internal/system/dto"
	"megin/internal/system/model"
	repo "megin/internal/system/repository"
	"megin/pkg/context/api"
	"megin/pkg/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SysApiToken struct {
	base.Service
	Repo     *repo.SysApiToken
	UserRepo *repo.SysUser
	Jwt      *SysJwt
}

func NewSysApiToken(ctx *api.Context) *SysApiToken {
	s := &SysApiToken{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysApiToken(ctx)
	s.UserRepo = repo.NewSysUser(ctx)
	s.Jwt = NewSysJwt(ctx)
	return s
}

func (s *SysApiToken) CreateApiToken(req *systemDto.CreateApiTokenReq) (string, error) {
	user, err := s.UserRepo.GetById(req.UserID)
	if err != nil {
		return "", s.Error(err, "查询用户失败")
	}
	if user.ID == 0 {
		return "", s.ErrorMessage("用户不存在")
	}

	authorities, err := s.UserRepo.GetUserAuthorities(req.UserID)
	if err != nil {
		return "", s.Error(err, "查询用户角色失败")
	}

	hasAuth := user.AuthorityId == req.AuthorityID
	if !hasAuth {
		for _, auth := range authorities {
			if auth.AuthorityId == req.AuthorityID {
				hasAuth = true
				break
			}
		}
	}
	if !hasAuth {
		return "", s.ErrorMessage("用户不具备该角色权限")
	}

	expireDuration := time.Duration(req.Days) * 24 * time.Hour
	if req.Days == -1 {
		expireDuration = 100 * 365 * 24 * time.Hour
	}
	if expireDuration <= 0 {
		expireDuration = 30 * 24 * time.Hour
	}
	expiresAt := time.Now().Add(expireDuration)

	claims := &commonDto.Claims{
		UserID:   int(user.ID),
		Username: user.Username,
		RoleId:   int(req.AuthorityID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(config.GetConfig().Jwt.Secret))
	if err != nil {
		return "", s.Error(err, "生成token失败")
	}

	now := time.Now()
	apiToken := model.SysApiToken{
		UserID:      user.ID,
		AuthorityID: req.AuthorityID,
		Token:       tokenStr,
		Status:      true,
		ExpiresAt:   expiresAt,
		Remark:      req.Remark,
		SystemModel: base.SystemModel{
			CreatedAt: utils.TimePtr(now),
			UpdatedAt: utils.TimePtr(now),
		},
	}
	if err := s.Repo.Create(&apiToken); err != nil {
		return "", s.Error(err, "保存API Token失败")
	}

	return tokenStr, nil
}

func (s *SysApiToken) GetApiTokenList(req *systemDto.GetApiTokenListReq) (*systemDto.PageResult[model.SysApiToken], error) {
	query := s.Repo.DB().Model(&model.SysApiToken{}).Preload("User")
	if req.UserID != 0 {
		query = query.Where("user_id = ?", req.UserID)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, s.Error(err, "查询API Token列表失败")
	}

	var list []model.SysApiToken
	if err := query.Order("created_at desc").
		Limit(req.PageSize).
		Offset((req.PageNo - 1) * req.PageSize).
		Find(&list).Error; err != nil {
		return nil, s.Error(err, "查询API Token列表失败")
	}

	return &systemDto.PageResult[model.SysApiToken]{
		PageNo:    req.PageNo,
		PageSize:  req.PageSize,
		TotalSize: total,
		List:      list,
	}, nil
}

func (s *SysApiToken) DeleteApiToken(id uint) error {
	apiToken, err := s.Repo.GetById(id)
	if err != nil {
		return s.Error(err, "查询API Token失败")
	}
	if apiToken.ID == 0 {
		return s.ErrorMessage("API Token不存在")
	}

	if err := s.Jwt.JsonInBlacklist(apiToken.Token); err != nil {
		return s.Error(err, "作废Token失败")
	}

	apiToken.Status = false
	apiToken.UpdatedAt = utils.TimePtr(time.Now())
	return s.Repo.Save(&apiToken)
}
