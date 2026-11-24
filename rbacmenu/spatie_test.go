package rbacmenu_test

import (
	"testing"

	"github.com/HumaCode/go-rbac-menu/rbacmenu"
)

func TestMatchMenuURL(t *testing.T) {
	menuURLs := []string{
		"/dashboard",
		"/ppid/permohonan",
		"/ppid/permohonan/masuk",
		"/system/users",
	}

	tests := []struct {
		requestPath string
		expected    string
	}{
		{"/dashboard", "/dashboard"},
		{"/ppid/permohonan", "/ppid/permohonan"},
		{"/ppid/permohonan/123", "/ppid/permohonan"},
		{"/ppid/permohonan/masuk/detail", "/ppid/permohonan/masuk"},
		{"/system/users/01HQ999/edit", "/system/users"},
		{"/unknown/path", ""},
	}

	for _, tc := range tests {
		matched := rbacmenu.MatchMenuURL(tc.requestPath, menuURLs)
		if matched != tc.expected {
			t.Errorf("for path %s expected matched %s, got %s", tc.requestPath, tc.expected, matched)
		}
	}
}
