package captcha

import (
	"sync"

	"github.com/mojocn/base64Captcha"
)

// CaptchaResponse 验证码响应结构体
type CaptchaResponse struct {
	CaptchaId     string `json:"captchaId"`
	PicPath       string `json:"picPath"`
	CaptchaLength int    `json:"captchaLength"`
	OpenCaptcha   bool   `json:"openCaptcha"`
}

var (
	captchaStore = base64Captcha.DefaultMemStore
	once         sync.Once
)

// 配置验证码生成器
func init() {
	once.Do(func() {
		// 设置验证码存储
		captchaStore = base64Captcha.DefaultMemStore
	})
}

// Generate 生成验证码
func Generate() (CaptchaResponse, error) {
	// 配置验证码
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, captchaStore)

	// 生成验证码
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		return CaptchaResponse{}, err
	}

	return CaptchaResponse{
		CaptchaId:     id,
		PicPath:       b64s,
		CaptchaLength: 5,
		OpenCaptcha:   true,
	}, nil
}

// Verify 验证验证码
func Verify(id, code string) bool {
	return captchaStore.Verify(id, code, true)
}
