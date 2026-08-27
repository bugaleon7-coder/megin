package service

import (
	"megin/internal/config"
	"sync"
	"testing"

	"gorm.io/gorm"
)

func TestNormalizeCasbinPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "admin api path", path: "/admin-api/user/getUserInfo", want: "/user/getUserInfo"},
		{name: "prefixless path", path: "/user/getUserInfo", want: "/user/getUserInfo"},
		{name: "empty path", path: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeCasbinPath(tt.path); got != tt.want {
				t.Fatalf("NormalizeCasbinPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsDefaultAuthenticatedPolicy(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		method string
		want   bool
	}{
		{name: "totp status with admin prefix", path: "/admin-api/user/getTotpStatus", method: "GET", want: true},
		{name: "totp enable without prefix", path: "/user/enableTotp", method: "POST", want: true},
		{name: "self setting", path: "/admin-api/user/setSelfSetting", method: "PUT", want: true},
		{name: "wrong method", path: "/admin-api/user/getTotpStatus", method: "POST", want: false},
		{name: "non default endpoint", path: "/admin-api/user/getUserList", method: "POST", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDefaultAuthenticatedPolicy(tt.path, tt.method); got != tt.want {
				t.Fatalf("IsDefaultAuthenticatedPolicy(%q, %q) = %v, want %v", tt.path, tt.method, got, tt.want)
			}
		})
	}
}

func TestGetEnforcerReusesSharedInstance(t *testing.T) {
	ctx, db := newSystemTestContext(t)
	resetCasbinEnforcerForTest(t, db)

	service := NewSysCasbin(ctx)
	first, err := service.getEnforcer()
	if err != nil {
		t.Fatalf("first getEnforcer() error = %v", err)
	}
	second, err := service.getEnforcer()
	if err != nil {
		t.Fatalf("second getEnforcer() error = %v", err)
	}
	if first != second {
		t.Fatal("getEnforcer() should reuse shared enforcer instance")
	}
}

func resetCasbinEnforcerForTest(t *testing.T, db *gorm.DB) {
	t.Helper()
	casbinEnforcerMu.Lock()
	defer casbinEnforcerMu.Unlock()
	casbinEnforcerInitOnce = sync.Once{}
	casbinEnforcerInstance = nil
	casbinEnforcerInitErr = nil
	casbinDBProvider = func() *gorm.DB { return db }
	t.Cleanup(func() {
		casbinEnforcerMu.Lock()
		defer casbinEnforcerMu.Unlock()
		casbinEnforcerInitOnce = sync.Once{}
		casbinEnforcerInstance = nil
		casbinEnforcerInitErr = nil
		casbinDBProvider = config.GetMysqlDB
	})
}
