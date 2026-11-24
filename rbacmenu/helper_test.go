package rbacmenu_test

import (
	"sort"
	"testing"
)

func TestURLMenuLengthSort(t *testing.T) {
	urls := []string{
		"/dashboard",
		"/ppid",
		"/ppid/permohonan",
		"/ppid/permohonan/masuk",
		"/ppid/permohonan/masuk/detail",
	}

	sort.Slice(urls, func(i, j int) bool {
		return len(urls[i]) > len(urls[j])
	})

	if urls[0] != "/ppid/permohonan/masuk/detail" {
		t.Errorf("expected longest URL first, got %s", urls[0])
	}
	if urls[len(urls)-1] != "/ppid" {
		t.Errorf("expected shortest URL last, got %s", urls[len(urls)-1])
	}
}

