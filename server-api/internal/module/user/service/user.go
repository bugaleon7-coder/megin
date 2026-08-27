package service

import (
	"crypto/rand"
	"encoding/hex"
	"megin/internal/base"
	bizcache "megin/internal/cache"
	userModel "megin/internal/module/user/model"
	userRepo "megin/internal/module/user/repository"
	"megin/pkg/context/api"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User 封装 C 端注册用户账号相关能力。
type User struct {
	base.Service
	Repo *userRepo.User
}

func NewUser(ctx *api.Context) *User {
	s := &User{}
	s.Initialize(ctx)
	s.Repo = userRepo.NewUser(ctx)
	return s
}

// Register 创建一个新的 C 端注册用户账号，并使用 bcrypt 保存密码。
func (s *User) Register(loginName, password string) (*userModel.UserInfo, error) {
	exists, err := s.Repo.ExistsByLoginName(loginName)
	if err != nil {
		return nil, s.Error(err, "检查前台用户账号是否存在失败")
	}
	if exists {
		return nil, s.ErrorMessage("账号已存在")
	}

	salt, err := generatePasswordSalt(16)
	if err != nil {
		return nil, s.Error(err, "生成前台用户密码盐值失败")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(buildPasswordWithSalt(password, salt)), bcrypt.DefaultCost)
	if err != nil {
		return nil, s.Error(err, "生成bcrypt密码摘要失败")
	}

	now := time.Now()
	user := &userModel.UserInfo{
		LoginName: loginName,
		Password:  string(hashedPassword),
		Salt:      salt,
		CreatedAt: &now,
		UpdatedAt: &now,
	}
	if err := s.Repo.Create(user); err != nil {
		return nil, s.Error(err, "创建前台用户失败")
	}
	return user, nil
}

// GetUserInfo 根据用户ID获取 C 端用户基础信息。
func (s *User) GetUserInfo(userID uint) (*userModel.UserInfo, error) {
	user, err := s.Repo.GetById(userID)
	if err != nil {
		return nil, s.Error(err, "查询前台用户失败")
	}
	if user.ID == 0 {
		return nil, s.ErrorMessage("用户不存在")
	}
	return &user, nil
}

// VerifyLogin 校验 C 端注册用户登录账号和密码。
func (s *User) VerifyLogin(loginName, password string) (*userModel.UserInfo, error) {
	user, err := s.Repo.GetByLoginName(loginName)
	if err != nil {
		return nil, s.Error(err, "查询前台用户失败")
	}
	if user == nil || user.ID == 0 {
		return nil, s.ErrorMessage("用户不存在")
	}
	matched, err := verifyUserPassword(user, password)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, s.ErrorMessage("用户名或密码错误")
	}
	return user, nil
}

// AfterLoginSuccess 在登录成功后回写最近登录时间和最新 token。
func (s *User) AfterLoginSuccess(userID uint, token string, expiresAtMillis int64) error {
	now := time.Now()
	if err := s.Repo.DB().Model(&userModel.UserInfo{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"token":           token,
			"last_login_time": &now,
			"updated_at":      &now,
		}).Error; err != nil {
		return s.Error(err, "更新前台用户登录状态失败")
	}
	ttl := time.Until(time.UnixMilli(expiresAtMillis))
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	if err := bizcache.SetRedisString(bizcache.GetApiUserLoginTokenKey(userID), token, ttl); err != nil {
		return s.Error(err, "写入前台用户登录token到Redis失败")
	}
	return nil
}

// ValidateLoginToken 校验 C 端用户当前 token 是否与 Redis 中保存的一致。
func (s *User) ValidateLoginToken(userID uint, token string) error {
	redisToken, err := bizcache.GetRedisString(bizcache.GetApiUserLoginTokenKey(userID))
	if err != nil {
		return s.Error(err, "读取前台用户登录token失败")
	}
	if redisToken == "" {
		return s.ErrorMessage("登录状态已失效,请重新登录", 403)
	}
	if redisToken != token {
		return s.ErrorMessage("您的帐户已在其他设备登录,请重新登录", 403)
	}
	return nil
}

// verifyUserPassword 校验前台用户密码，只接受 bcrypt 密码格式。
func verifyUserPassword(user *userModel.UserInfo, plainPassword string) (matched bool, err error) {
	storedPassword := strings.TrimSpace(user.Password)
	if storedPassword == "" {
		return false, nil
	}

	// bcrypt 是当前前台用户密码的标准存储格式。
	if strings.HasPrefix(storedPassword, "$2a$") || strings.HasPrefix(storedPassword, "$2b$") || strings.HasPrefix(storedPassword, "$2y$") {
		if err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(buildPasswordWithSalt(plainPassword, user.Salt))); err != nil {
			return false, nil
		}
		return true, nil
	}
	return false, nil
}

// buildPasswordWithSalt 构建参与 bcrypt 的最终密码输入。
// 当前规则为明文密码拼接数据库中的随机盐值。
func buildPasswordWithSalt(password string, salt string) string {
	return password + salt
}

// generatePasswordSalt 生成用于前台用户密码摘要的随机盐值。
func generatePasswordSalt(byteSize int) (string, error) {
	buf := make([]byte, byteSize)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
