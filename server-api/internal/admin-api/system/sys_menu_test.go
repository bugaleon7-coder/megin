package system

import (
	"reflect"
	"testing"

	systemDto "megin/internal/system/dto"
)

func TestExtractMenuIDsRecursivelyAndDeduplicates(t *testing.T) {
	menus := []systemDto.SysBaseMenu{
		{ID: 3, Children: []systemDto.SysBaseMenu{
			{ID: 10},
			{ID: 13, Children: []systemDto.SysBaseMenu{{ID: 50}}},
		}},
		{ID: 13},
	}
	want := []uint{3, 10, 13, 50}
	if got := extractMenuIds(menus); !reflect.DeepEqual(got, want) {
		t.Fatalf("extractMenuIds() = %v, want %v", got, want)
	}
}
