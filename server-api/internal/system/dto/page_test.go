package dto

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPageResultUsesFrontendContract(t *testing.T) {
	b, err := json.Marshal(PageResult[int]{PageNo: 2, PageSize: 20, TotalSize: 1, List: []int{7}})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	got := string(b)
	for _, key := range []string{`"page":2`, `"pageSize":20`, `"total":1`, `"list":[7]`} {
		if !strings.Contains(got, key) {
			t.Fatalf("response %s does not contain %s", got, key)
		}
	}
	for _, oldKey := range []string{"page_no", "page_size", "total_size"} {
		if strings.Contains(got, oldKey) {
			t.Fatalf("response still contains legacy key %q: %s", oldKey, got)
		}
	}
}
